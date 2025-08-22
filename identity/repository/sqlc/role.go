package sqlcrepo

import (
	"context"

	sqlc "p9e.in/ugcl/identity/db/sqlc/generated"
	"p9e.in/ugcl/identity/mappers"
	"p9e.in/ugcl/identity/models"
)

type SQLCRoleRepo struct {
	queries *sqlc.Queries
}

func NewSQLCRoleRepo(q *sqlc.Queries) *SQLCRoleRepo {
	return &SQLCRoleRepo{queries: q}
}

func (r *SQLCRoleRepo) Create(ctx context.Context, role *models.Role) error {
	params, err := mappers.RoleModelToSQLC(role)
	if err != nil {
		return err
	}
	return r.queries.CreateRole(ctx, sqlc.CreateRoleParams{
		ID:       role.ID,
		Name:     role.Name,
		Metadata: params.Metadata,
	})
}

func (r *SQLCRoleRepo) GetByID(ctx context.Context, id string) (*models.Role, error) {
	res, err := r.queries.GetRoleByID(ctx, sqlc.GetRoleByIDParams{
		ID: id,
	})
	if err != nil {
		return nil, err
	}
	role, err := mappers.RoleSQLCToModel(res)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *SQLCRoleRepo) Update(ctx context.Context, role *models.Role) error {
	params, err := mappers.RoleModelToSQLC(role)
	if err != nil {
		return err
	}
	return r.queries.UpdateRole(ctx, sqlc.UpdateRoleParams{
		ID:       role.ID,
		Name:     role.Name,
		Metadata: params.Metadata,
	})
}

func (r *SQLCRoleRepo) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteRole(ctx, sqlc.DeleteRoleParams{
		ID: id,
	})
}

func (r *SQLCRoleRepo) List(ctx context.Context) ([]*models.Role, error) {
	roles, err := r.queries.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Role, len(roles))
	for i := range roles {
		result[i], _ = mappers.RoleSQLCToModel(roles[i])
	}
	return result, nil
}

func (r *SQLCRoleRepo) ListPermissionsForRole(ctx context.Context, roleID string) (*models.Role, []*models.Permission, error) {
	row, errs := r.queries.GetRoleByID(ctx, sqlc.GetRoleByIDParams{
		ID: roleID,
	})
	if errs != nil {
		return nil, nil, errs
	}
	rows, err := r.queries.ListPermissionsForRole(ctx, sqlc.ListPermissionsForRoleParams{
		RoleID: roleID,
	})
	if err != nil {
		return nil, nil, err
	}

	perms := make([]*models.Permission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, mappers.PermissionSQLCToModel(row))
	}
	role, err := mappers.RoleSQLCToModel(row)
	if err != nil {
		return nil, nil, err
	}
	return role, perms, nil
}
