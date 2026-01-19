package uow

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlcUnitOfWorkFactory struct {
	db *pgxpool.Pool
}

func NewSQLCFactory(db *pgxpool.Pool) UnitOfWorkFactory {
	return &sqlcUnitOfWorkFactory{
		db: db,
	}
}

func (f *sqlcUnitOfWorkFactory) Create() UnitOfWork {
	return NewSQLCUnitOfWork(f.db)
}