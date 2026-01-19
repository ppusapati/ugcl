package sqlcrepo

import (
	"context"

	sqlc "p9e.in/ugcl/identity/user/db/sqlc/generated"
	"p9e.in/ugcl/identity/user/mappers"
	"p9e.in/ugcl/identity/user/models"
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
	return r.queries.CreateRole(ctx, *params)
}

func (r *SQLCRoleRepo) GetByID(ctx context.Context, id string) (*models.Role, error) {
	// Convert string ID to int64
	var numID int64
	if id != "" {
		// Simple hash of string ID to int64
		for _, b := range []byte(id) {
			numID = numID*31 + int64(b)
		}
	}

	res, err := r.queries.GetRoleByID(ctx, sqlc.GetRoleByIDParams{
		ID: numID,
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
		ID:          params.ID,
		ParentID:    params.ParentID,
		Name:        role.Name,
		IsPreserved: role.IsPreserved,
		Metadata:    params.Metadata,
	})
}

func (r *SQLCRoleRepo) Delete(ctx context.Context, id string) error {
	// Convert string ID to int64
	var numID int64
	if id != "" {
		// Simple hash of string ID to int64
		for _, b := range []byte(id) {
			numID = numID*31 + int64(b)
		}
	}
	return r.queries.DeleteRole(ctx, sqlc.DeleteRoleParams{
		ID: numID,
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
	// Convert string ID to int64
	var numID int64
	if roleID != "" {
		// Simple hash of string ID to int64
		for _, b := range []byte(roleID) {
			numID = numID*31 + int64(b)
		}
	}

	roleRow, errs := r.queries.GetRoleByID(ctx, sqlc.GetRoleByIDParams{
		ID: numID,
	})
	if errs != nil {
		return nil, nil, errs
	}
	rows, err := r.queries.ListPermissionsForRole(ctx, sqlc.ListPermissionsForRoleParams{
		RoleID: numID,
	})
	if err != nil {
		return nil, nil, err
	}

	perms := make([]*models.Permission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, mappers.PermissionSQLCToModel(row))
	}
	role, err := mappers.RoleSQLCToModel(roleRow)
	if err != nil {
		return nil, nil, err
	}
	return role, perms, nil
}
