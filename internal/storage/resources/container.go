package resources

import (
	"context"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/storage/adapters/pgrepo"
	"github.com/vlasashk/websocket-chat/internal/storage/ports/httpchi"
	"github.com/vlasashk/websocket-chat/internal/storage/usecase"
	"github.com/vlasashk/websocket-chat/pkg/kakafka"
	"github.com/vlasashk/websocket-chat/pkg/logger"
)

type resource[T any] struct {
	sync.Once
	value   T
	initErr error
}

type Container struct {
	logger      *resource[zerolog.Logger]
	httpServ    *resource[*http.Server]
	pgRepo      *resource[usecase.Repo]
	kafkaReader *resource[*kakafka.Consumer]
}

func New() *Container {
	return &Container{
		logger:      &resource[zerolog.Logger]{},
		httpServ:    &resource[*http.Server]{},
		pgRepo:      &resource[usecase.Repo]{},
		kafkaReader: &resource[*kakafka.Consumer]{},
	}
}

func (c *resource[T]) get(init func() (T, error)) (T, error) {
	c.Do(func() {
		c.value, c.initErr = init()
	})
	if c.initErr != nil {
		return *new(T), c.initErr
	}
	return c.value, nil
}

func (r *Container) GetHttp(ctx context.Context, cfg config.StorageConfig) (*http.Server, error) {
	repo, err := r.GetRepo(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return r.httpServ.get(func() (*http.Server, error) {
		return httpchi.New(ctx, cfg.HTTP, repo), nil
	})
}

func (r *Container) GetLogger(lvl string) (zerolog.Logger, error) {
	return r.logger.get(func() (zerolog.Logger, error) {
		return logger.New(lvl)
	})
}

func (r *Container) GetRepo(ctx context.Context, cfg config.StorageConfig) (usecase.Repo, error) {
	log, err := r.GetLogger(cfg.LoggerLVL)
	if err != nil {
		return nil, err
	}

	return r.pgRepo.get(func() (usecase.Repo, error) {
		repo, err := pgrepo.New(ctx, cfg.Repo, log)
		if err != nil {
			return nil, err
		}
		return repo, nil
	})
}

func (r *Container) GetKafkaConsumer(ctx context.Context, cfg config.StorageConfig) (*kakafka.Consumer, error) {
	log, err := r.GetLogger(cfg.LoggerLVL)
	if err != nil {
		return nil, err
	}

	return r.kafkaReader.get(func() (*kakafka.Consumer, error) {
		return kakafka.NewConsumer(ctx, log, cfg.Kafka), nil
	})
}
