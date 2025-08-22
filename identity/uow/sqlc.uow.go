package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "p9e.in/ugcl/identity/db/sqlc/generated"
	repo "p9e.in/ugcl/identity/repository/interfaces"
	sqlcrepo "p9e.in/ugcl/identity/repository/sqlc"
)

type SQLCUnitOfWork struct {
	tx   pgx.Tx
	user repo.UserRepository
	role repo.RoleRepository
	perm repo.PermissionRepository
	pdef repo.PermissionDefRepository
}

// Begin implements UnitOfWorkFactory.
func (u *SQLCUnitOfWork) Begin(ctx context.Context) (UnitOfWork, error) {
	return NewSQLCUnitOfWork(u.tx), nil
}

func NewSQLCUnitOfWork(tx pgx.Tx) *SQLCUnitOfWork {
	q := sqlc.New(tx)
	return &SQLCUnitOfWork{
		tx:   tx,
		user: sqlcrepo.NewSQLCUserRepo(q),
		role: sqlcrepo.NewSQLCRoleRepo(q),
		perm: sqlcrepo.NewSQLCPermissionRepo(q),
		pdef: sqlcrepo.NewSQLCPermissionDefRepo(q),
	}
}

func (u *SQLCUnitOfWork) UserRepo() repo.UserRepository {
	return u.user
}

func (u *SQLCUnitOfWork) RoleRepo() repo.RoleRepository {
	return u.role
}

func (u *SQLCUnitOfWork) PermissionRepo() repo.PermissionRepository {
	return u.perm
}

func (u *SQLCUnitOfWork) PermissionDefRepo() repo.PermissionDefRepository {
	return u.pdef
}

func (u *SQLCUnitOfWork) Commit(ctx context.Context) error {
	return u.tx.Commit(ctx)
}

func (u *SQLCUnitOfWork) Rollback(ctx context.Context) error {
	return u.tx.Rollback(ctx)
}

func BeginSQLCUnitOfWork(ctx context.Context, db *pgxpool.Pool) (UnitOfWork, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return NewSQLCUnitOfWork(tx), nil
}
