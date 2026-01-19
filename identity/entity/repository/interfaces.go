package repository

import (
	"context"

	db "p9e.in/ugcl/identity/entity/db/generated"
)

// IEntityRepository defines the contract for entity data access
type IEntityRepository interface {
	Create(ctx context.Context, arg db.CreateEntityParams) (*db.Entity, error)
	Update(ctx context.Context, arg db.UpdateEntityParams) (*db.Entity, error)
	GetByID(ctx context.Context, arg db.GetEntityParams) (*db.Entity, error)
	GetByUserID(ctx context.Context, arg db.GetEntityByUserIdParams) (*db.Entity, error)
	GetByReference(ctx context.Context, arg db.GetEntityByReferenceParams) (*db.Entity, error)
	List(ctx context.Context, arg db.ListEntitiesParams) ([]db.Entity, error)
	Count(ctx context.Context, arg db.CountEntitiesParams) (int64, error)
	Delete(ctx context.Context, arg db.DeleteEntityParams) error
}

// IEntityRoleBindingRepository defines the contract for entity role binding data access
type IEntityRoleBindingRepository interface {
	Create(ctx context.Context, arg db.CreateEntityRoleBindingParams) (*db.EntityRoleBinding, error)
	GetByEntityID(ctx context.Context, arg db.GetEntityRoleBindingsParams) ([]db.EntityRoleBinding, error)
	GetActiveByEntityID(ctx context.Context, arg db.GetActiveEntityRoleBindingsParams) ([]db.EntityRoleBinding, error)
	Delete(ctx context.Context, arg db.DeleteEntityRoleBindingParams) error
	DeleteAllForEntity(ctx context.Context, arg db.DeleteAllEntityRoleBindingsParams) error
}
