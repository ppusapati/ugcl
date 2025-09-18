package sqlcrepo

import (
	"context"
	"fmt"

	sqlc "p9e.in/ugcl/identity/user/db/sqlc/generated"
	"p9e.in/ugcl/identity/user/mappers"
	"p9e.in/ugcl/identity/user/models"
)

type SQLCPermissionDefRepo struct {
	queries *sqlc.Queries
}

func NewSQLCPermissionDefRepo(q *sqlc.Queries) *SQLCPermissionDefRepo {
	return &SQLCPermissionDefRepo{queries: q}
}

// Create inserts a new PermissionDef into the database.
func (r *SQLCPermissionDefRepo) Create(ctx context.Context, def *models.PermissionDef) error {
	params := mappers.PermissionDefModelToSQLC(def)
	_, err := r.queries.CreatePermissionDef(ctx, params)
	return err
}

// GetByName fetches a PermissionDef by name.
func (r *SQLCPermissionDefRepo) GetByName(ctx context.Context, name string) (*models.PermissionDef, error) {
	row, err := r.queries.GetPermissionDefByName(ctx, sqlc.GetPermissionDefByNameParams{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	return mappers.PermissionDefSQLCToModel(row), nil
}

// ListAll fetches all permission definitions.
func (r *SQLCPermissionDefRepo) ListAll(ctx context.Context) ([]*models.PermissionDef, error) {
	rows, err := r.queries.ListAllPermissionDefs(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*models.PermissionDef, 0, len(rows))
	for _, row := range rows {
		result = append(result, mappers.PermissionDefSQLCToModel(row))
	}
	return result, nil
}

// ListGroups fetches permission groups with their associated definitions.
func (r *SQLCPermissionDefRepo) ListGroups(ctx context.Context) ([]*models.PermissionDefGroup, error) {
	groups, err := r.queries.ListPermissionDefGroups(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.PermissionDefGroup, 0, len(groups))
	for _, g := range groups {
		// Assuming you will replace this with proper query:
		defs, err := r.queries.ListDefsByGroup(ctx, sqlc.ListDefsByGroupParams{
			Name: g.Name,
		})
		if err != nil {
			return nil, err
		}

		modelDefs := make([]*models.PermissionDef, 0, len(defs))
		for _, def := range defs {
			modelDefs = append(modelDefs, mappers.PermissionDefSQLCToModel(def))
		}

		result = append(result, &models.PermissionDefGroup{
			Name: g.Name,
			Defs: modelDefs,
			// Map other fields as needed
		})
	}
	return result, nil
}

// Delete removes a permission definition.
func (r *SQLCPermissionDefRepo) Delete(ctx context.Context, def string) error {
	return r.queries.DeletePermissionDef(ctx, sqlc.DeletePermissionDefParams{
		Name: def,
	})
}

func (r *SQLCPermissionDefRepo) GetByID(ctx context.Context, id int32) (*models.PermissionDef, error) {
	row, err := r.queries.GetPermissionDefByID(ctx, sqlc.GetPermissionDefByIDParams{
		ID: int64(id),
	})
	if err != nil {
		return nil, err
	}
	return mappers.PermissionDefSQLCToModel(row), nil
}

// Update modifies an existing PermissionDef.
func (r *SQLCPermissionDefRepo) Update(ctx context.Context, def *models.PermissionDef) (*models.PermissionDef, error) {
	params := mappers.PermissionDefModelToSQLCUpdate(def)
	err := r.queries.UpdatePermissionDef(ctx, params)
	if err != nil {
		return nil, err
	}
	return def, nil
}

func (r *SQLCPermissionDefRepo) AssignPermissionDefToRoleName(ctx context.Context, roleID, defName string) error {
	// Convert string role ID to int64
	var numRoleID int64
	if roleID != "" {
		// Simple hash of string ID to int64
		for _, b := range []byte(roleID) {
			numRoleID = numRoleID*31 + int64(b)
		}
	}

	return r.queries.AssignPermissionToRole(ctx, sqlc.AssignPermissionToRoleParams{
		RoleID:            numRoleID,
		PermissionDefName: defName,
	})
}

func (r *SQLCPermissionDefRepo) AssignPermissionDefToRole(ctx context.Context, roleID string, def *models.PermissionDef, resource string) error {
	perm := mappers.MaterializeDefToPermission(def, roleID, resource, nil)
	err := r.queries.GrantPermission(ctx, sqlc.GrantPermissionParams{
		Namespace: perm.Namespace,
		Resource:  perm.Resource,
		Action:    perm.Action,
		Subject:   perm.Subject,
		Effect:    mappers.ModelToDBEffectInterface(perm.Effect),
	})
	return err
}

func (r *SQLCPermissionDefRepo) RevokeFromRole(ctx context.Context, roleID, defName string) error {
	// Convert string role ID to int64
	var numRoleID int64
	if roleID != "" {
		// Simple hash of string ID to int64
		for _, b := range []byte(roleID) {
			numRoleID = numRoleID*31 + int64(b)
		}
	}

	// Step 1: Fetch the permission definition
	defRow, err := r.queries.GetPermissionDefByName(ctx, sqlc.GetPermissionDefByNameParams{
		Name: defName,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch permission def: %w", err)
	}

	// Step 2: Convert to model
	def := mappers.PermissionDefSQLCToModel(defRow)

	// Step 3: Materialize to actual permission (subject = roleID)
	perm := mappers.MaterializeDefToPermission(def, roleID, def.Resource, nil)

	// Step 4: Revoke the concrete permission
	if err := r.queries.RevokePermission(ctx, sqlc.RevokePermissionParams{
		Namespace: perm.Namespace,
		Resource:  perm.Resource,
		Action:    perm.Action,
		Subject:   perm.Subject,
	}); err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	// Step 5: Remove from role_permission_defs
	if err := r.queries.RemoveRolePermissionDef(ctx, sqlc.RemoveRolePermissionDefParams{
		RoleID:            numRoleID,
		PermissionDefName: defName,
	}); err != nil {
		return fmt.Errorf("failed to remove permission def link: %w", err)
	}

	return nil
}

func (r *SQLCPermissionDefRepo) ListByRole(ctx context.Context, roleID string) ([]*models.PermissionDef, error) {
	// Convert string role ID to int64
	var numRoleID int64
	if roleID != "" {
		// Simple hash of string ID to int64
		for _, b := range []byte(roleID) {
			numRoleID = numRoleID*31 + int64(b)
		}
	}

	rows, err := r.queries.ListPermissionDefsByRole(ctx, sqlc.ListPermissionDefsByRoleParams{
		RoleID: numRoleID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*models.PermissionDef, 0, len(rows))
	for _, row := range rows {
		result = append(result, mappers.PermissionDefSQLCToModel(row))
	}
	return result, nil
}
