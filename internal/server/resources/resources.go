package resources

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/server/adapters/manager"
	"github.com/vlasashk/websocket-chat/internal/server/adapters/rediska"
	"github.com/vlasashk/websocket-chat/pkg/kakafka"
	"github.com/vlasashk/websocket-chat/pkg/logger"
)

type Resources struct {
	Cfg           config.ServerCfg
	Log           zerolog.Logger
	ClientManager ClientManager
	RedisRepo     CacheRepo
	KafkaWriter   MessageBroker
}

func New(ctx context.Context, cfg config.ServerCfg) (*Resources, error) {
	log, err := logger.New(cfg.LoggerLVL)
	if err != nil {
		return nil, err
	}

	res := Resources{
		Cfg:           cfg,
		Log:           log,
		ClientManager: manager.New(log),
		KafkaWriter:   kakafka.NewProducer(ctx, cfg.Kafka, log),
	}

	repo, err := rediska.NewClient(res.Cfg.Redis)
	if err != nil {
		return nil, err
	}
	res.RedisRepo = repo

	return &res, nil
}
