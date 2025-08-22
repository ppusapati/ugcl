package sqlcrepo

import (
	"context"

	sqlc "p9e.in/ugcl/identity/db/sqlc/generated"
	"p9e.in/ugcl/identity/mappers"
	"p9e.in/ugcl/identity/models"
)

type SQLCPermissionRepo struct {
	queries *sqlc.Queries
}

func NewSQLCPermissionRepo(q *sqlc.Queries) *SQLCPermissionRepo {
	return &SQLCPermissionRepo{queries: q}
}

func (r *SQLCPermissionRepo) GetBySubject(ctx context.Context, subject string) ([]*models.Permission, error) {
	rows, err := r.queries.GetPermissionsBySubject(ctx, sqlc.GetPermissionsBySubjectParams{
		Subject: subject,
	})
	if err != nil {
		return nil, err
	}
	return mappers.PermissionSQLCRowsToModels(rows), nil
}

func (r *SQLCPermissionRepo) Check(ctx context.Context, namespace, resource, action, subject string) (*models.Permission, error) {
	row, err := r.queries.CheckPermission(ctx, sqlc.CheckPermissionParams{
		Namespace: namespace,
		Resource:  resource,
		Action:    action,
		Subject:   subject,
	})
	if err != nil {
		return nil, err
	}
	return mappers.PermissionSQLCToModel(row), nil
}

func (r *SQLCPermissionRepo) Delete(ctx context.Context, namespace, resource, action, subject string) error {
	return r.queries.DeletePermission(ctx, sqlc.DeletePermissionParams{
		Namespace: namespace,
		Resource:  resource,
		Action:    action,
		Subject:   subject,
	})
}

func (r *SQLCPermissionRepo) ListAll(ctx context.Context) ([]*models.Permission, error) {
	rows, err := r.queries.ListAllPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return mappers.PermissionSQLCRowsToModels(rows), nil
}

func (r *SQLCPermissionRepo) Grant(ctx context.Context, permission *models.Permission) error {
	return r.queries.GrantPermission(ctx, mappers.PermissionModelToGrantSQLC(permission))
}
func (r *SQLCPermissionRepo) Revoke(ctx context.Context, namespace, resource, action, subject string) error {
	return r.queries.RevokePermission(ctx, sqlc.RevokePermissionParams{
		Namespace: namespace,
		Resource:  resource,
		Action:    action,
		Subject:   subject,
	})
}
