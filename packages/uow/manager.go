package uow

import (
	"context"
)

func WithTx(ctx context.Context, factory Factory, fn func(UnitOfWork) error) error {
	tx, err := factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func WithRead(ctx context.Context, factory Factory, fn func(UnitOfWork) error) error {
	tx, err := factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	return fn(tx)
}
