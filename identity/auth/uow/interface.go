package uow

import (
	"context"

	"p9e.in/ugcl/identity/auth/repository/interfaces"
)

// UnitOfWork interface defines transaction boundaries and repository access
type UnitOfWork interface {
	// Transaction management
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	// Repository access
	Users() interfaces.UserRepository
	Sessions() interfaces.SessionRepository
	ApiKeys() interfaces.ApiKeyRepository
	Tokens() interfaces.TokenRepository
	TwoFactor() interfaces.TwoFactorRepository
	Audit() interfaces.AuditRepository

	// Execution patterns
	Execute(ctx context.Context, fn func(ctx context.Context, uow UnitOfWork) error) error
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context, uow UnitOfWork) error) error
}

// UnitOfWorkFactory creates new unit of work instances
type UnitOfWorkFactory interface {
	Create() UnitOfWork
}