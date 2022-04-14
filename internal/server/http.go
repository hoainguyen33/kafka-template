package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *server) runHttpServer() {
	s.gin.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Ok")
	})
	// s.gin.GET("/metrics", gin.WrapH(promhttp.Handler()))
	// s.mapRoutes()

	go func() {
		sr := &http.Server{
			Addr:           ":8080",
			Handler:        s.gin,
			ReadTimeout:    time.Second * s.cfg.Http.ReadTimeout,
			WriteTimeout:   time.Second * s.cfg.Http.WriteTimeout,
			MaxHeaderBytes: maxHeaderBytes,
		}
		sr.ListenAndServe()
		if err := s.gin.RunTLS(s.cfg.Http.Port, certFile, keyFile); err != nil {
			s.log.Fatalf("Error starting TLS Server: ", err)
		}
	}()
}
