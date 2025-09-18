package uow

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "p9e.in/ugcl/identity/user/db/sqlc/generated"
	sqlcrepo "p9e.in/ugcl/identity/user/repository/sqlc"
)

// --- SQLC Factory ---
type SQLCUnitOfWorkFactory struct {
	db *pgxpool.Pool
}

func NewSQLCUnitOfWorkFactory(db *pgxpool.Pool) *SQLCUnitOfWorkFactory {
	return &SQLCUnitOfWorkFactory{db: db}
}

func (f *SQLCUnitOfWorkFactory) Begin(ctx context.Context) (UnitOfWork, error) {
	tx, err := f.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	q := sqlc.New(tx)
	return &SQLCUnitOfWork{
		tx:   tx,
		user: sqlcrepo.NewSQLCUserRepo(q),
		role: sqlcrepo.NewSQLCRoleRepo(q),
		perm: sqlcrepo.NewSQLCPermissionRepo(q),
		pdef: sqlcrepo.NewSQLCPermissionDefRepo(q),
	}, nil
}
