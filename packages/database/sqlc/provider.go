package sqlc

import (
	"context"
	"fmt"
	"time"

	conf "p9e.in/ugcl/packages/api/v1/config"

	"github.com/jackc/pgx/v5/pgxpool"
	employeeDB "p9e.in/ugcl/employee/db/generated"
	formInstanceDb "p9e.in/ugcl/formbuilder/db/generated"
	formbuilderDb "p9e.in/ugcl/formbuilder/db/generated"
	workflowDb "p9e.in/ugcl/formbuilder/db/generated"
	tenantDB "p9e.in/ugcl/identity/tenant/db/generated"
	userDB "p9e.in/ugcl/identity/user/db/sqlc/generated"
	"p9e.in/ugcl/identity/user/uow"
	notificationDB "p9e.in/ugcl/notification/db/generated"
	organizationDB "p9e.in/ugcl/organization/db/generated"
	dairySiteDB "p9e.in/ugcl/projects/db/generated"
	contractorDB "p9e.in/ugcl/vendors/db/generated"
)

type DatabaseManager struct {
	Pool *pgxpool.Pool
}

func NewDatabaseManager(cfg *conf.Data) (*DatabaseManager, error) {
	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=disable search_path=testing2",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Dbname,
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 30 * time.Minute
	config.MaxConnLifetime = 2 * time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseManager{Pool: pool}, nil
}

// GetContractorQueries returns a new instance of contractor queries
func (m *DatabaseManager) GetContractorQueries() *contractorDB.Queries {
	return contractorDB.New(m.Pool)
}

func (m *DatabaseManager) GetUserQueries() *userDB.Queries {
	return userDB.New(m.Pool)
}

func (m *DatabaseManager) GetTenantQueries() *tenantDB.Queries {
	return tenantDB.New(m.Pool)
}

func (m *DatabaseManager) GetUserUOW() *uow.SQLCUnitOfWorkFactory {
	return uow.NewSQLCUnitOfWorkFactory(m.Pool)
}

func (m *DatabaseManager) GetDairySiteQueries() *dairySiteDB.Queries {
	return dairySiteDB.New(m.Pool)
}

func (m *DatabaseManager) GetFormBuilderQueries() *formbuilderDb.Queries {
	return formbuilderDb.New(m.Pool)
}

func (m *DatabaseManager) GetFormInstanceQueries() *formInstanceDb.Queries {
	return formInstanceDb.New(m.Pool)
}

func (m *DatabaseManager) GetWorkflowQueries() *workflowDb.Queries {
	return workflowDb.New(m.Pool)
}

// ExecRaw executes raw SQL queries for dynamic table operations
// This is a generic method that can be used by any repository that needs raw SQL execution
func (m *DatabaseManager) ExecRaw(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return m.Pool.Exec(ctx, query, args...)
}

func (m *DatabaseManager) GetNotificationQueries() *notificationDB.Queries {
	return notificationDB.New(m.Pool)
}

func (m *DatabaseManager) GetOrganizationQueries() *organizationDB.Queries {
	return organizationDB.New(m.Pool)
}

// GetEmployeeQueries returns a new instance of employee queries
func (m *DatabaseManager) GetEmployeeQueries() *employeeDB.Queries {
	return employeeDB.New(m.Pool)
}
