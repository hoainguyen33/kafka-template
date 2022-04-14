package middlewares

import (
	"kafka-test/constant/config"
	"kafka-test/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpTotalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "http_microservice_total_requests",
		Help: "The total number of incoming HTTP requests",
	})
)

// MiddlewareManager http middlewares
type middlewareManager struct {
	log logger.Logger
	cfg *config.Config
}

// MiddlewareManager interface
type MiddlewareManager interface {
	Metrics(next gin.HandlerFunc) gin.HandlerFunc
}

// NewMiddlewareManager constructor
func NewMiddlewareManager(log logger.Logger, cfg *config.Config) *middlewareManager {
	return &middlewareManager{log: log, cfg: cfg}
}

// Metrics prometheus metrics
func (m *middlewareManager) Metrics(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpTotalRequests.Inc()
		next(c)
	}
}
