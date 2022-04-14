package grpc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SuccessMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "accounts_success_incoming_grpc_messages_total",
		Help: "The total number of success incoming success gRPC messages",
	})
	ErrorMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "accounts_error_incoming_grpc_message_total",
		Help: "The total number of error incoming success gRPC messages",
	})
	CreateMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "accounts_create_incoming_grpc_requests_total",
		Help: "The total number of incoming create account gRPC messages",
	})
	UpdateMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "accounts_update_incoming_grpc_requests_total",
		Help: "The total number of incoming update account gRPC messages",
	})
	GetByIdMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "accounts_get_by_id_incoming_grpc_requests_total",
		Help: "The total number of incoming get by id account gRPC messages",
	})
	SearchMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "accounts_search_incoming_grpc_requests_total",
		Help: "The total number of incoming search accounts gRPC messages",
	})
)
