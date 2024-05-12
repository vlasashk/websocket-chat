package pgrepo

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	addMsgQuery  = `INSERT INTO messages (user_id, content) VALUES ($1, $2);`
	addUserQuery = `INSERT INTO users (username) VALUES ($1) RETURNING user_id;`
)

func (pg PgRepo) AddMessage(ctx context.Context, userID int, msg string) error {
	start := time.Now()
	if _, err := pg.Pool.Exec(ctx, addMsgQuery, userID, msg); err != nil {
		return err
	}
	log.Info().Dur("postgres msg add time", time.Since(start)).Send()
	return nil
}

func (pg PgRepo) AddUser(ctx context.Context, UserName string) (int, error) {
	var userID int
	start := time.Now()
	if err := pg.Pool.QueryRow(ctx, addUserQuery, UserName).Scan(&userID); err != nil {
		return 0, err
	}
	log.Info().Dur("postgres user add time", time.Since(start)).Send()
	return userID, nil
}
