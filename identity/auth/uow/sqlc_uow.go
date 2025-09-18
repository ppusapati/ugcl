package uow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/ugcl/identity/auth/repository/interfaces"
	"p9e.in/ugcl/identity/auth/repository/sqlc"
)

type sqlcUnitOfWork struct {
	db *pgxpool.Pool
	tx pgx.Tx

	// Repository instances
	userRepo      interfaces.UserRepository
	sessionRepo   interfaces.SessionRepository
	apiKeyRepo    interfaces.ApiKeyRepository
	tokenRepo     interfaces.TokenRepository
	twoFactorRepo interfaces.TwoFactorRepository
	auditRepo     interfaces.AuditRepository
}

func NewSQLCUnitOfWork(db *pgxpool.Pool) UnitOfWork {
	return &sqlcUnitOfWork{
		db: db,
	}
}

// Transaction management

func (uow *sqlcUnitOfWork) Begin(ctx context.Context) error {
	if uow.tx != nil {
		return fmt.Errorf("transaction already started")
	}

	tx, err := uow.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	uow.tx = tx

	// Reset repository instances to use the transaction
	uow.resetRepositories()

	return nil
}

func (uow *sqlcUnitOfWork) Commit(ctx context.Context) error {
	if uow.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}

	err := uow.tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	uow.tx = nil
	uow.resetRepositories()

	return nil
}

func (uow *sqlcUnitOfWork) Rollback(ctx context.Context) error {
	if uow.tx == nil {
		return fmt.Errorf("no transaction to rollback")
	}

	err := uow.tx.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	uow.tx = nil
	uow.resetRepositories()

	return nil
}

// Repository access

func (uow *sqlcUnitOfWork) Users() interfaces.UserRepository {
	if uow.userRepo == nil {
		uow.userRepo = sqlc.NewUserRepository(uow.db, uow.tx)
	}
	return uow.userRepo
}

func (uow *sqlcUnitOfWork) Sessions() interfaces.SessionRepository {
	if uow.sessionRepo == nil {
		uow.sessionRepo = sqlc.NewSessionRepository(uow.db, uow.tx)
	}
	return uow.sessionRepo
}

func (uow *sqlcUnitOfWork) ApiKeys() interfaces.ApiKeyRepository {
	if uow.apiKeyRepo == nil {
		uow.apiKeyRepo = sqlc.NewApiKeyRepository(uow.db, uow.tx)
	}
	return uow.apiKeyRepo
}

func (uow *sqlcUnitOfWork) Tokens() interfaces.TokenRepository {
	if uow.tokenRepo == nil {
		uow.tokenRepo = sqlc.NewTokenRepository(uow.db, uow.tx)
	}
	return uow.tokenRepo
}

func (uow *sqlcUnitOfWork) TwoFactor() interfaces.TwoFactorRepository {
	if uow.twoFactorRepo == nil {
		uow.twoFactorRepo = sqlc.NewTwoFactorRepository(uow.db, uow.tx)
	}
	return uow.twoFactorRepo
}

func (uow *sqlcUnitOfWork) Audit() interfaces.AuditRepository {
	if uow.auditRepo == nil {
		uow.auditRepo = sqlc.NewAuditRepository(uow.db, uow.tx)
	}
	return uow.auditRepo
}

// Execution patterns

func (uow *sqlcUnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context, uow UnitOfWork) error) error {
	return fn(ctx, uow)
}

func (uow *sqlcUnitOfWork) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context, uow UnitOfWork) error) error {
	if err := uow.Begin(ctx); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = uow.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(ctx, uow); err != nil {
		if rollbackErr := uow.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("failed to rollback after error: %v (original error: %w)", rollbackErr, err)
		}
		return err
	}

	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Helper methods

func (uow *sqlcUnitOfWork) resetRepositories() {
	uow.userRepo = nil
	uow.sessionRepo = nil
	uow.apiKeyRepo = nil
	uow.tokenRepo = nil
	uow.twoFactorRepo = nil
	uow.auditRepo = nil
}
