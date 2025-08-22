package repository

import (
	sqlc "p9e.in/ugcl/identity/db/sqlc/generated"
	ri "p9e.in/ugcl/identity/repository/interfaces"
	sqlcrepo "p9e.in/ugcl/identity/repository/sqlc"

	"go.uber.org/fx"
)

type RepositoryContainer struct {
	UserRepo          ri.UserRepository
	RoleRepo          ri.RoleRepository
	PermissionRepo    ri.PermissionRepository
	PermissionDefRepo ri.PermissionDefRepository
}

// Repository container provider
func ProvideRepositoryContainer(
	userRepo ri.UserRepository,
	roleRepo ri.RoleRepository,
	permRepo ri.PermissionRepository,
	permDefRepo ri.PermissionDefRepository,
) *RepositoryContainer {
	return &RepositoryContainer{
		UserRepo:          userRepo,
		RoleRepo:          roleRepo,
		PermissionRepo:    permRepo,
		PermissionDefRepo: permDefRepo,
	}
}

// Repository providers for SQLC implementation
func ProvideSQLCUserRepo(queries *sqlc.Queries) ri.UserRepository {
	return sqlcrepo.NewSQLCUserRepo(queries)
}

func ProvideSQLCRoleRepo(queries *sqlc.Queries) ri.RoleRepository {
	return sqlcrepo.NewSQLCRoleRepo(queries)
}

func ProvideSQLCPermissionRepo(queries *sqlc.Queries) ri.PermissionRepository {
	return sqlcrepo.NewSQLCPermissionRepo(queries)
}

func ProvideSQLCPermissionDefRepo(queries *sqlc.Queries) ri.PermissionDefRepository {
	return sqlcrepo.NewSQLCPermissionDefRepo(queries)
}

var SQLCRepoModule = fx.Module("sqlc_repo",
	fx.Provide(
		ProvideSQLCUserRepo,
		ProvideSQLCRoleRepo,
		ProvideSQLCPermissionRepo,
		ProvideSQLCPermissionDefRepo,
	),
)
