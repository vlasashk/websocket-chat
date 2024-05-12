package migrations

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

func Up(pool *pgxpool.Pool, path string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.Up(db, path); err != nil {
		return err
	}

	if err := db.Close(); err != nil {
		log.Error().Err(err).Send()
	}

	return nil
}

func Down(pool *pgxpool.Pool, path string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.Down(db, path); err != nil {
		return err
	}

	if err := db.Close(); err != nil {
		log.Error().Err(err).Send()
	}

	return nil
}
