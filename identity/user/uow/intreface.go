// Make sure the interface is exported (capitalized):
package uow

import (
	"context"

	repo "p9e.in/ugcl/identity/user/repository/interfaces"
)

type UnitOfWork interface {
	UserRepo() repo.UserRepository
	RoleRepo() repo.RoleRepository
	PermissionRepo() repo.PermissionRepository
	PermissionDefRepo() repo.PermissionDefRepository

	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UnitOfWorkFactory interface {
	Begin(ctx context.Context) (UnitOfWork, error)
}
