package p9context

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// AppContext holds the context for the application, including the dynamic database pool
type DBContext struct {
	DBPool            *pgxpool.Pool
	DBPoolShared      *pgxpool.Pool
	DBPoolIndependent map[string]*pgxpool.Pool
}
