package processor

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"github.com/vlasashk/websocket-chat/internal/storage/usecase"
	"github.com/vlasashk/websocket-chat/pkg/kakafka"
	"github.com/vlasashk/websocket-chat/pkg/response"
)

type KafkaProc struct {
	consumer *kakafka.Consumer
	logger   zerolog.Logger
	repo     usecase.Repo
}

func NewProcessor(consumer *kakafka.Consumer, logger zerolog.Logger, repo usecase.Repo) *KafkaProc {
	return &KafkaProc{
		consumer: consumer,
		logger:   logger,
		repo:     repo,
	}
}

func (p *KafkaProc) ProcessEvents(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-p.consumer.Messages:
			if !ok {
				return nil
			}
			var userMsg response.Msg
			if err := json.Unmarshal(msg.Value, &userMsg); err != nil {
				p.logger.Error().Err(err).Send()
				if err = p.consumer.Commiter(ctx, msg); err != nil {
					p.logger.Error().Err(err).Send()
					return err
				}
				continue
			}

			if userMsg.UserID == 0 {
				p.logger.Error().Msg("user ID was not provided in the message")
				if err := p.consumer.Commiter(ctx, msg); err != nil {
					return err
				}
				continue
			}

			if err := p.repo.AddMessage(ctx, userMsg.UserID, userMsg.Text); err != nil {
				p.logger.Error().Err(err).Send()
			}
			if err := p.consumer.Commiter(ctx, msg); err != nil {
				return err
			}
		}
	}
}
