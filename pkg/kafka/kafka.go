package kafka

import (
	"context"
	"kafka-test/constant/config"

	"github.com/segmentio/kafka-go"
)

func NewKafkaConn(cfg *config.Config) (*kafka.Conn, error) {
	return kafka.DialContext(context.Background(), "tcp", cfg.Kafka.Brokers[0])
}
