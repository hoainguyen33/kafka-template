package producer

import (
	"context"
	constants "kafka-test/constant"
	"kafka-test/constant/config"
	"kafka-test/constant/topics"
	"kafka-test/pkg/logger"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// AccountsProducer interface
type AccountsProducer interface {
	PublishCreate(ctx context.Context, msgs ...kafka.Message) error
	PublishUpdate(ctx context.Context, msgs ...kafka.Message) error
	Close()
	Run()
	GetNewKafkaWriter(topic string) *kafka.Writer
}

type accountsProducer struct {
	log         logger.Logger
	cfg         *config.Config
	loginWriter *kafka.Writer
	// createWriter *kafka.Writer
	// updateWriter *kafka.Writer
}

// NewAccountsProducer constructor
func NewAccountsProducer(log logger.Logger, cfg *config.Config) *accountsProducer {
	return &accountsProducer{log: log, cfg: cfg}
}

func (p *accountsProducer) NewReader(kafkaURLs []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:                kafkaURLs,
		GroupID:                groupID,
		Topic:                  topic,
		MinBytes:               constants.MinBytes,
		MaxBytes:               constants.MaxBytes,
		QueueCapacity:          constants.QueueCapacity,
		HeartbeatInterval:      constants.HeartbeatInterval,
		CommitInterval:         constants.CommitInterval,
		PartitionWatchInterval: constants.PartitionWatchInterval,
		Logger:                 kafka.LoggerFunc(p.log.Debugf),
		ErrorLogger:            kafka.LoggerFunc(p.log.Errorf),
		MaxAttempts:            constants.MaxAttempts,
		Dialer: &kafka.Dialer{
			Timeout: constants.DialTimeout,
		},
	})
}

// GetNewKafkaWriter Create new kafka writer
func (p *accountsProducer) NewWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(p.cfg.Kafka.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireNone,
		MaxAttempts:  constants.WriterMaxAttempts,
		Logger:       kafka.LoggerFunc(p.log.Debugf),
		ErrorLogger:  kafka.LoggerFunc(p.log.Errorf),
		Compression:  compress.Snappy,
		ReadTimeout:  constants.WriterReadTimeout,
		WriteTimeout: constants.WriterWriteTimeout,
	}
}

// Run init producers writers
func (p *accountsProducer) Run() {
	p.loginWriter = p.NewWriter(topics.LoginAccountTopic)
	// p.createWriter = p.NewWriter(topics.CreateAccountTopic)
	// p.updateWriter = p.NewWriter(topics.UpdateAccountTopic)
}

// Close close writers
func (p accountsProducer) Close() {
	p.loginWriter.Close()
	// p.createWriter.Close()
	// p.updateWriter.Close()
}

// PublishLogin publish messages to login topic
func (p *accountsProducer) PublishLogin(ctx context.Context, msgs ...kafka.Message) error {
	return p.loginWriter.WriteMessages(ctx, msgs...)
}

// PublishCreate publish messages to create topic
// func (p *accountsProducer) PublishCreate(ctx context.Context, msgs ...kafka.Message) error {
// 	return p.createWriter.WriteMessages(ctx, msgs...)
// }

// PublishUpdate publish messages to update topic
// func (p *accountsProducer) PublishUpdate(ctx context.Context, msgs ...kafka.Message) error {
// 	return p.updateWriter.WriteMessages(ctx, msgs...)
// }
