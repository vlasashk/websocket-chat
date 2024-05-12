package usecase

import (
	"context"
)

type Repo interface {
	AddMessage(ctx context.Context, userID int, msg string) error
	AddUser(ctx context.Context, UserName string) (int, error)
}
