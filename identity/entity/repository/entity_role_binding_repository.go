package repository

import (
	"context"

	db "p9e.in/ugcl/identity/entity/db/generated"
)

// EntityRoleBindingRepository implements IEntityRoleBindingRepository
type EntityRoleBindingRepository struct {
	queries *db.Queries
}

// NewEntityRoleBindingRepository creates a new entity role binding repository
func NewEntityRoleBindingRepository(queries *db.Queries) *EntityRoleBindingRepository {
	return &EntityRoleBindingRepository{
		queries: queries,
	}
}

func (r *EntityRoleBindingRepository) Create(ctx context.Context, arg db.CreateEntityRoleBindingParams) (db.EntityRoleBinding, error) {
	return r.queries.CreateEntityRoleBinding(ctx, arg)
}

func (r *EntityRoleBindingRepository) GetByEntityID(ctx context.Context, arg db.GetEntityRoleBindingsParams) ([]db.EntityRoleBinding, error) {
	return r.queries.GetEntityRoleBindings(ctx, arg)
}

func (r *EntityRoleBindingRepository) GetActiveByEntityID(ctx context.Context, arg db.GetActiveEntityRoleBindingsParams) ([]db.EntityRoleBinding, error) {
	return r.queries.GetActiveEntityRoleBindings(ctx, arg)
}

func (r *EntityRoleBindingRepository) Delete(ctx context.Context, arg db.DeleteEntityRoleBindingParams) error {
	return r.queries.DeleteEntityRoleBinding(ctx, arg)
}

func (r *EntityRoleBindingRepository) DeleteAllForEntity(ctx context.Context, arg db.DeleteAllEntityRoleBindingsParams) error {
	return r.queries.DeleteAllEntityRoleBindings(ctx, arg)
}
