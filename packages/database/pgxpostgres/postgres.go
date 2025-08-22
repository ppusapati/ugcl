package pgxpostgres

import (
	"context"
	"fmt"
	"time"

	"p9e.in/ugcl/packages/api/v1/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxOpenConns    = 30
	connMaxLifetime = 60
	maxIdleConns    = 10
	connMaxIdleTime = 10
)

// AppContext holds the context for the application, including the dynamic database pool
type DBContext struct {
	DBPool            *pgxpool.Pool
	DBPoolShared      *pgxpool.Pool
	DBPoolIndependent map[string]*pgxpool.Pool
}

// NewDBContext creates and returns a new instance of DBContext.
func NewDBContext(conn *pgxpool.Pool) *DBContext {
	return &DBContext{
		DBPoolShared:      conn,
		DBPoolIndependent: make(map[string]*pgxpool.Pool),
	}
}

// NewPgx initializes a new PostgreSQL connection pool and returns it along with a cleanup function.
func NewPgx(c *config.Data) (*pgxpool.Pool, func(), error) {
	dataSourceName := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable connect_timeout=5 statement_timeout=15000 idle_in_transaction_session_timeout=15000 pool_max_conns=%d",
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Dbname,
		maxOpenConns,
	)

	fmt.Printf("Attempting to connect to PostgreSQL with max connections: %d\n", maxOpenConns)

	config, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	// Add connection logging
	config.BeforeConnect = func(ctx context.Context, config *pgx.ConnConfig) error {
		fmt.Printf("Attempting new database connection to %s:%d\n", config.Host, config.Port)
		return nil
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		fmt.Printf("Successfully established new connection to %s:%d\n", conn.Config().Host, conn.Config().Port)
		return nil
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Configure connection pool settings
	config.MaxConnLifetime = connMaxLifetime * time.Second
	config.MaxConnIdleTime = connMaxIdleTime * time.Second
	config.MaxConns = int32(maxOpenConns)
	config.MinConns = int32(maxIdleConns)

	// Test the connection with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Testing database connection with ping...")
	if err = conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("database connection test failed: %w", err)
	}
	fmt.Println("Ping successful")

	// Test a simple query to verify full connectivity
	fmt.Println("Testing database connection with query...")
	if _, err = conn.Exec(ctx, "SELECT 1"); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("database query test failed: %w", err)
	}
	fmt.Println("Query test successful")

	// Log initial pool statistics
	stats := conn.Stat()
	fmt.Printf("Connection pool initialized - Total: %d, Acquired: %d, Idle: %d\n",
		stats.TotalConns(), stats.AcquiredConns(), stats.IdleConns())

	// Cleanup function to close the connection pool
	cleanup := func() {
		fmt.Println("Closing connection pool...")
		conn.Close()
	}

	return conn, cleanup, nil
}
