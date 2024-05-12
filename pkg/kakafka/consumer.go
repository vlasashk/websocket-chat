package kakafka

import (
	"context"
	"errors"
	"io"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"github.com/vlasashk/websocket-chat/config"
)

type Consumer struct {
	Messages <-chan kafka.Message
	Commiter func(ctx context.Context, msgs ...kafka.Message) error
	handler  func() error
}

func NewConsumer(ctx context.Context, log zerolog.Logger, cfg config.KafkaCfg) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Addr},
		GroupID: cfg.GroupID,
		Topic:   cfg.Topic,
	})

	msgChan := make(chan kafka.Message)

	handler := func() error {
		defer close(msgChan)
		for {
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
					return nil
				}
				log.Error().Err(err).Send()
				return err
			}
			select {
			case msgChan <- m:
			case <-ctx.Done():
				if err = reader.Close(); err != nil {
					log.Error().Err(err).Send()
					return err
				}
				return nil
			}
		}
	}
	return &Consumer{
		Messages: msgChan,
		Commiter: reader.CommitMessages,
		handler:  handler,
	}
}

func (c *Consumer) Run() error {
	if c == nil {
		return errors.New("consumer is not initialized")
	}
	return c.handler()
}
