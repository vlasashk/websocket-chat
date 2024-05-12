package client

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/client/models"
	"github.com/vlasashk/websocket-chat/pkg/logger"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg config.ClientCfg) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log, err := logger.New(cfg.LoggerLVL)
	if err != nil {
		return err
	}

	user, err := models.NewUser(cfg)
	if err != nil {
		return err
	}
	defer user.Close(log)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return user.Receiver(gCtx, log)
	})
	g.Go(func() error {
		return user.Sender(gCtx, log)
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	return nil
}
