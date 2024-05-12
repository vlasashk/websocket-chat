package kakafka

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"github.com/vlasashk/websocket-chat/config"
)

type Producer struct {
	writer *kafka.Writer
	msgs   chan kafka.Message
	logger zerolog.Logger
}

func NewProducer(ctx context.Context, cfg config.KafkaCfg, logger zerolog.Logger) *Producer {
	w := &kafka.Writer{
		Addr:      kafka.TCP(cfg.Addr),
		Topic:     cfg.Topic,
		Balancer:  &kafka.LeastBytes{},
		BatchSize: cfg.BatchSize,
		Async:     cfg.Async,
	}

	producer := &Producer{
		writer: w,
		msgs:   make(chan kafka.Message, 100),
		logger: logger,
	}

	go producer.run(ctx)

	return producer
}

func (p *Producer) run(ctx context.Context) {
	defer func() {
		if err := p.writer.Close(); err != nil {
			p.logger.Error().Err(err).Msg("failed to close writer")
		}
	}()
	for {
		select {
		case msg := <-p.msgs:
			err := p.writer.WriteMessages(ctx, msg)
			if err != nil {
				p.logger.Error().Err(err).Msg("Failed to write messages")
			}
		case <-ctx.Done():
			p.logger.Info().Msg("closing producer")
			return
		}
	}
}

func (p *Producer) Write(ctx context.Context, data []byte) {
	kafkaMsg := kafka.Message{
		Value: data,
	}
	select {
	case p.msgs <- kafkaMsg:
	case <-ctx.Done():
		return
	}
}
