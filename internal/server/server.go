package server

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
	"github.com/vlasashk/websocket-chat/internal/server/ports/httpchi"
	"github.com/vlasashk/websocket-chat/internal/server/resources"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg config.ServerCfg) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	g, gCtx := errgroup.WithContext(ctx)

	container, err := resources.New(gCtx, cfg)
	if err != nil {
		return err
	}

	srv := httpchi.NewServer(gCtx, container)

	g.Go(func() error {
		container.Log.Info().Msg(fmt.Sprintf("starting server: %s", net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)))
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	<-gCtx.Done()
	container.Log.Info().Msg("Got interruption signal")

	// Additional timeout for shutting down process
	ctxDown, cancelDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelDown()

	if err = srv.Shutdown(ctxDown); err != nil {
		return err
	}

	if err = g.Wait(); err != nil {
		container.Log.Error().Err(err).Send()
	}
	container.Log.Info().Msg("server was gracefully shut down")

	return nil
}
