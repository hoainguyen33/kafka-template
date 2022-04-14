package consumer

import (
	"context"
	"encoding/json"
	constants "kafka-test/constant"
	"kafka-test/internal/models"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/segmentio/kafka-go"
)

const (
	retryAttempts = 1
	retryDelay    = 1 * time.Second
)

func (ac *AccountsConsumer) LoginWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	r *kafka.Reader,
	w *kafka.Writer,
	wg *sync.WaitGroup,
	workerID int,
) {
	defer wg.Done()
	defer cancel()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			ac.log.Errorf("FetchMessage", err)
			return
		}

		ac.log.Infof(
			"WORKER: %v, message at topic/partition/offset %v/%v/%v: %s = %s\n",
			workerID,
			m.Topic,
			m.Partition,
			m.Offset,
			string(m.Key),
			string(m.Value),
		)
		constants.IncomingMessages.Inc()

		var acc models.Account
		if err := json.Unmarshal(m.Value, &acc); err != nil {
			constants.ErrorMessages.Inc()
			ac.log.Errorf("json.Unmarshal", err)
			continue
		}
		if err := ac.validate.StructCtx(ctx, acc); err != nil {
			constants.ErrorMessages.Inc()
			ac.log.Errorf("validate.StructCtx", err)
			continue
		}

		if err := retry.Do(func() error {
			// test username and password demo
			if acc.Username != "test" || acc.Password != "12345678" {
				return err
			}
			ac.log.Infof("login: %v", acc)
			return nil
		},
			retry.Attempts(retryAttempts),
			retry.Delay(retryDelay),
			retry.Context(ctx),
		); err != nil {
			constants.ErrorMessages.Inc()

			if err := ac.publishErrorMessage(ctx, w, m, err); err != nil {
				ac.log.Errorf("publishErrorMessage", err)
				continue
			}
			ac.log.Errorf("accountsUC.Create.publishErrorMessage", err)
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			constants.ErrorMessages.Inc()
			ac.log.Errorf("CommitMessages", err)
			continue
		}

		constants.SuccessMessages.Inc()
	}
}

func (acg *AccountsConsumer) publishErrorMessage(ctx context.Context, w *kafka.Writer, m kafka.Message, err error) error {
	errMsg := &models.ErrorMessage{
		Offset:    m.Offset,
		Error:     err.Error(),
		Time:      m.Time.UTC(),
		Partition: m.Partition,
		Topic:     m.Topic,
	}

	errMsgBytes, err := json.Marshal(errMsg)
	if err != nil {
		return err
	}

	return w.WriteMessages(ctx, kafka.Message{
		Value: errMsgBytes,
	})
}
