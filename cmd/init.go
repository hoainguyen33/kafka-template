package main

import (
	"kafka-test/constant/config"
	"kafka-test/internal/server"
	"kafka-test/pkg/jaeger"
	"kafka-test/pkg/kafka"
	"kafka-test/pkg/logger"
	"kafka-test/pkg/redis"
	"log"

	"github.com/opentracing/opentracing-go"
)

func init() {
	log.Println("Starting accounts microservice")
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}
	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()
	appLogger.Info("Starting user server")
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, DevelopmentMode: %s",
		cfg.AppVersion,
		cfg.Logger.Level,
		cfg.Server.Development,
	)
	appLogger.Infof("Success parsed config: %#v", cfg.AppVersion)
	tracer, closer, err := jaeger.InitJaeger(cfg)
	if err != nil {
		appLogger.Fatal("cannot create tracer", err)
	}
	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	appLogger.Info("Opentracing connected")

	// connect Database
	// mongoDBConn, err := mongodb.NewMongoDBConn(ctx, cfg)
	// if err != nil {
	// 	appLogger.Fatal("cannot connect mongodb", err)
	// }
	// defer func() {
	// 	if err := mongoDBConn.Disconnect(ctx); err != nil {
	// 		appLogger.Fatal("mongoDBConn.Disconnect", err)
	// 	}
	// }()
	// appLogger.Infof("MongoDB connected: %v", mongoDBConn.NumberSessionsInProgress())
	// connect kafka
	conn, err := kafka.NewKafkaConn(cfg)
	if err != nil {
		appLogger.Fatal("NewKafkaConn", err)
	}
	defer conn.Close()
	brokers, err := conn.Brokers()
	if err != nil {
		appLogger.Fatal("conn.Brokers", err)
	}
	appLogger.Infof("Kafka connected: %v", brokers)
	redisClient := redis.NewRedisClient(cfg)
	appLogger.Info("Redis connected")
	s := server.NewServer(appLogger, cfg, tracer, redisClient)
	appLogger.Fatal(s.Run())
}
