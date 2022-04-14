package consumer

import (
	"context"
	constants "kafka-test/constant"
	"kafka-test/constant/config"
	"kafka-test/constant/topics"
	"kafka-test/pkg/logger"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// AccountsConsumerGroup struct
type AccountsConsumer struct {
	Brokers  []string
	GroupID  string
	log      logger.Logger
	cfg      *config.Config
	validate *validator.Validate
}

// NewAccountsConsumerGroup constructor
func NewAccountsConsumerGroup(
	brokers []string,
	groupID string,
	log logger.Logger,
	cfg *config.Config,
	validate *validator.Validate,
) *AccountsConsumer {
	return &AccountsConsumer{
		Brokers:  brokers,
		GroupID:  groupID,
		log:      log,
		cfg:      cfg,
		validate: validate,
	}
}

func (ac *AccountsConsumer) NewReader(kafkaURLs []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: kafkaURLs,
		// GroupID:                groupID,
		Topic:                  topic,
		MinBytes:               constants.MinBytes,
		MaxBytes:               constants.MaxBytes,
		QueueCapacity:          constants.QueueCapacity,
		HeartbeatInterval:      constants.HeartbeatInterval,
		CommitInterval:         constants.CommitInterval,
		PartitionWatchInterval: constants.PartitionWatchInterval,
		Logger:                 kafka.LoggerFunc(ac.log.Debugf),
		ErrorLogger:            kafka.LoggerFunc(ac.log.Errorf),
		MaxAttempts:            constants.MaxAttempts,
		Dialer: &kafka.Dialer{
			Timeout: constants.DialTimeout,
		},
	})
}

func (ac *AccountsConsumer) NewWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(ac.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireNone,
		MaxAttempts:  constants.WriterMaxAttempts,
		Logger:       kafka.LoggerFunc(ac.log.Debugf),
		ErrorLogger:  kafka.LoggerFunc(ac.log.Errorf),
		Compression:  compress.Snappy,
		ReadTimeout:  constants.WriterReadTimeout,
		WriteTimeout: constants.WriterWriteTimeout,
	}
}

func (ac *AccountsConsumer) consumeLogin(
	ctx context.Context,
	cancel context.CancelFunc,
	groupID string,
	topic string,
	workersNum int,
) {
	r := ac.NewReader(ac.Brokers, topic, groupID)
	defer cancel()
	defer func() {
		if err := r.Close(); err != nil {
			ac.log.Errorf("r.Close", err)
			cancel()
		}
	}()

	w := ac.NewWriter(topics.DeadLetterQueueTopic)
	defer func() {
		if err := w.Close(); err != nil {
			ac.log.Errorf("w.Close", err)
			cancel()
		}
	}()

	ac.log.Infof("Starting consumer group: %v", r.Config().GroupID)

	wg := &sync.WaitGroup{}
	for i := 0; i <= workersNum; i++ {
		wg.Add(1)
		go ac.LoginWorker(ctx, cancel, r, w, wg, i)
	}
	wg.Wait()
}

// RunConsumers run kafka consumers
func (ac *AccountsConsumer) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go ac.consumeLogin(ctx, cancel, constants.AccountsGroupID, topics.LoginAccountTopic, constants.LoginAccountWorker)
}
