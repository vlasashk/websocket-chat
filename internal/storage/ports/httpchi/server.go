package httpchi

import (
	"context"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/storage/usecase"
	"github.com/vlasashk/websocket-chat/pkg/response"
)

func New(ctx context.Context, cfg config.StorageAddr, repo usecase.Repo) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.URLFormat)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", HealthCheck)
	r.Post("/register", RegisterUser(ctx, repo))

	return &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: r,
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, render.M{
		"status": "ok",
	})
}

func RegisterUser(ctx context.Context, repo usecase.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var userReq response.RegisterReq
		if err := render.DecodeJSON(r.Body, &userReq); err != nil {
			log.Error().Err(err).Msg("error decoding body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrResp{Error: "bad json"})
			return
		}

		userID, err := repo.AddUser(ctx, userReq.Username)
		if err != nil {
			log.Error().Err(err).Msg("error decoding body")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrResp{Error: "failed to register"})
			return
		}

		resp := response.RegisterResp{UserID: userID}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, resp)
	}
}
