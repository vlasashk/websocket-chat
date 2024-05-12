package httpchi

import (
	"context"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlasashk/websocket-chat/internal/server/resources"
)

func NewServer(ctx context.Context, container *resources.Resources) *http.Server {
	cfg := container.Cfg
	return &http.Server{
		Addr:    net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		Handler: newRouter(ctx, container),
	}
}

func newRouter(ctx context.Context, container *resources.Resources) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.URLFormat)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)
	r.Get("/chat", EstablishWS(ctx, container))
	r.Get("/healthz", HealthCheck)
	return r
}
