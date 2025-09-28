// =============================================================================
// repository/repository_manager.go - Repository manager for DataBridge module (Import Only)
// =============================================================================
package repository

import (
	"github.com/jackc/pgx/v5"
	"go.uber.org/fx"
)

// RepositoryManager manages import-related repositories for DataBridge
type RepositoryManager struct {
	Mapping IMappingRepository
	Job     IJobRepository
}

// NewRepositoryManager creates a new repository manager with import-only repositories
func NewRepositoryManager(database *pgx.Conn) *RepositoryManager {
	return &RepositoryManager{
		Mapping: NewMappingRepository(database),
		Job:     NewJobRepository(database),
	}
}

// RepositoryModule provides dependency injection for import repositories only
var RepositoryModule = fx.Module("repository",
	fx.Provide(
		NewMappingRepository,
		NewJobRepository,
		NewRepositoryManager,
	),
)