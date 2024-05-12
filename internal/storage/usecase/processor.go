package usecase

import "context"

type Processor interface {
	ProcessEvents(ctx context.Context) error
}
