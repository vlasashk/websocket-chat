package storage

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/storage/adapters/processor"
	"github.com/vlasashk/websocket-chat/internal/storage/resources"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg config.StorageConfig) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	g, gCtx := errgroup.WithContext(ctx)

	container := resources.New()

	log, err := container.GetLogger(cfg.LoggerLVL)
	if err != nil {
		return err
	}
	srv, err := container.GetHttp(gCtx, cfg)
	if err != nil {
		return err
	}
	repo, err := container.GetRepo(ctx, cfg)
	if err != nil {
		return err
	}
	kafkaConsumer, err := container.GetKafkaConsumer(gCtx, cfg)
	if err != nil {
		return err
	}

	g.Go(kafkaConsumer.Run)
	g.Go(func() error {
		log.Info().Msg("starting processing kafka events")
		return processor.NewProcessor(kafkaConsumer, log, repo).ProcessEvents(gCtx)
	})
	g.Go(func() error {
		log.Info().Msg(fmt.Sprintf("starting server: %s", net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port)))
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	<-gCtx.Done()
	log.Info().Msg("Got interruption signal")

	// Additional timeout for shutting down process
	ctxDown, cancelDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelDown()

	if err = srv.Shutdown(ctxDown); err != nil {
		return err
	}

	if err = g.Wait(); err != nil {
		log.Error().Err(err).Send()
	}
	log.Info().Msg("server was gracefully shut down")

	return nil
}
