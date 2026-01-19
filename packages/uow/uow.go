package uow

import (
	"context"
)

type UnitOfWork interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Factory interface {
	Begin(ctx context.Context) (UnitOfWork, error)
}
