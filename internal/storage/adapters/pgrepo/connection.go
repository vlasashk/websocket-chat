package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/migrations"
)

type PgRepo struct {
	Pool *pgxpool.Pool
}

type tracer struct {
	log zerolog.Logger
}

var timeout = 10 * time.Second

func New(ctx context.Context, cfg config.RepoCfg, logger zerolog.Logger) (*PgRepo, error) {
	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s", cfg.Schema, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}

	pgxCfg.ConnConfig.Tracer = &tracer{log: logger}

	dbPool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err = dbPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping connection pool: %w", err)
	}

	if err = migrations.Up(dbPool, cfg.MigrationPath); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	return &PgRepo{dbPool}, nil
}

func (t *tracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	t.log.Info().Str("sql", data.SQL).Any("args", data.Args).Msg("Executing command")
	return ctx
}

func (t *tracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
}
