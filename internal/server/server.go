package server

import (
	"context"
	"crypto/tls"
	"fmt"
	constants "kafka-test/constant"
	"kafka-test/constant/config"
	"kafka-test/internal/delivery/grpc/account"
	"kafka-test/internal/delivery/kafka/consumer"
	"kafka-test/internal/delivery/kafka/producer"
	"kafka-test/internal/interceptors"
	"kafka-test/pkg/logger"
	accountsService "kafka-test/proto/account"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	certFile        = "ssl/server-cert.pem"
	keyFile         = "ssl/server-key.pem"
	maxHeaderBytes  = 1 << 20
	gzipLevel       = 5
	stackSize       = 1 << 10 // 1 KB
	csrfTokenHeader = "X-CSRF-Token"
	bodyLimit       = "2M"
)

// server
type server struct {
	log    logger.Logger
	cfg    *config.Config
	tracer opentracing.Tracer
	gin    *gin.Engine
	redis  *redis.Client
}

// NewServer constructor
func NewServer(log logger.Logger, cfg *config.Config, tracer opentracing.Tracer, redis *redis.Client) *server {
	return &server{log: log, cfg: cfg, tracer: tracer, gin: gin.Default(), redis: redis}
}

// Run Start server
func (s *server) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validate := validator.New()

	accountsProducer := producer.NewAccountsProducer(s.log, s.cfg)
	accountsProducer.Run()
	defer accountsProducer.Close()
	// accountUC := usecase.NewAccountUC(accountMongoRepo, accountRedisRepo, s.log, accountsProducer)

	im := interceptors.NewInterceptorManager(s.log, s.cfg)
	// mw := middlewares.NewMiddlewareManager(s.log, s.cfg)
	fmt.Println(s.cfg.Server.Port)
	l, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}
	defer l.Close()

	// server cert
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		s.log.Fatalf("failed to load key pair: %s", err)
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(config)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
			Timeout:           s.cfg.Server.Timeout * time.Second,
			MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
			Time:              s.cfg.Server.Timeout * time.Minute,
		}),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpcrecovery.UnaryServerInterceptor(),
			im.Logger,
		),
	)

	accountService := account.NewAccountService(s.log, validate)
	accountsService.RegisterAccountsServiceServer(grpcServer, accountService)
	grpc_prometheus.Register(grpcServer)

	// v1 := s.gin.Group("/api/v1")
	// v1.Use(mw.Metrics)

	// accountHandlers := accountsHttpV1.NewAccountHandlers(s.log, accountUC, validate, v1.Group("/accounts"), mw)
	// accountHandlers.MapRoutes()

	accountsCG := consumer.NewAccountsConsumerGroup(s.cfg.Kafka.Brokers, constants.AccountsGroupID, s.log, s.cfg, validate)
	accountsCG.RunConsumers(ctx, cancel)

	go func() {
		s.log.Infof("Server is listening on PORT: %s", s.cfg.Http.Port)
		s.runHttpServer()
	}()

	go func() {
		s.log.Infof("GRPC Server is listening on port: %s", s.cfg.Server.Port)
		s.log.Fatal(grpcServer.Serve(l))
	}()

	if s.cfg.Server.Development {
		reflection.Register(grpcServer)
	}

	metricsServer := gin.Default()
	metricsServer.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.log.Infof("Metrics server is running on port: %s", s.cfg.Metrics.Port)
	srv := &http.Server{
		Addr:    s.cfg.Metrics.Port,
		Handler: metricsServer,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.log.Error(err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		s.log.Errorf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		s.log.Errorf("ctx.Done: %v", done)
	}

	// if err := s.gin.Shutdown(ctx); err != nil {
	// 	return errors.Wrap(err, "echo.Server.Shutdown")
	// }

	if err := srv.Shutdown(ctx); err != nil {
		s.log.Errorf("metricsServer.Shutdown: %v", err)
	}
	grpcServer.GracefulStop()
	s.log.Info("Server Exited Properly")

	return nil
}
