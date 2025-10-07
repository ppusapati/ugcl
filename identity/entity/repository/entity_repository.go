package repository

import (
	"context"

	db "p9e.in/ugcl/identity/entity/db/generated"
)

// EntityRepository implements IEntityRepository
type EntityRepository struct {
	queries *db.Queries
}

// NewEntityRepository creates a new entity repository
func NewEntityRepository(queries *db.Queries) *EntityRepository {
	return &EntityRepository{
		queries: queries,
	}
}

func (r *EntityRepository) Create(ctx context.Context, arg db.CreateEntityParams) (db.Entity, error) {
	return r.queries.CreateEntity(ctx, arg)
}

func (r *EntityRepository) Update(ctx context.Context, arg db.UpdateEntityParams) (db.Entity, error) {
	return r.queries.UpdateEntity(ctx, arg)
}

func (r *EntityRepository) GetByID(ctx context.Context, arg db.GetEntityParams) (db.Entity, error) {
	return r.queries.GetEntity(ctx, arg)
}

func (r *EntityRepository) GetByUserID(ctx context.Context, arg db.GetEntityByUserIdParams) (db.Entity, error) {
	return r.queries.GetEntityByUserId(ctx, arg)
}

func (r *EntityRepository) GetByReference(ctx context.Context, arg db.GetEntityByReferenceParams) (db.Entity, error) {
	return r.queries.GetEntityByReference(ctx, arg)
}

func (r *EntityRepository) List(ctx context.Context, arg db.ListEntitiesParams) ([]db.Entity, error) {
	return r.queries.ListEntities(ctx, arg)
}

func (r *EntityRepository) Count(ctx context.Context, arg db.CountEntitiesParams) (int64, error) {
	return r.queries.CountEntities(ctx, arg)
}

func (r *EntityRepository) Delete(ctx context.Context, arg db.DeleteEntityParams) error {
	return r.queries.DeleteEntity(ctx, arg)
}
