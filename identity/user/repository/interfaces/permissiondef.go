package repointerfaces

import (
	"context"

	"p9e.in/ugcl/identity/user/models"
)

type PermissionDefRepository interface {
	Create(ctx context.Context, def *models.PermissionDef) error
	GetByName(ctx context.Context, name string) (*models.PermissionDef, error)
	GetByID(ctx context.Context, id int32) (*models.PermissionDef, error)
	Update(ctx context.Context, def *models.PermissionDef) (*models.PermissionDef, error)
	ListAll(ctx context.Context) ([]*models.PermissionDef, error)
	ListGroups(ctx context.Context) ([]*models.PermissionDefGroup, error)
	ListByRole(ctx context.Context, roleID string) ([]*models.PermissionDef, error)
	Delete(ctx context.Context, uuid string) error

	AssignPermissionDefToRoleName(ctx context.Context, roleID, defName string) error
	RevokeFromRole(ctx context.Context, roleID string, defName string) error
	AssignPermissionDefToRole(ctx context.Context, roleID string, def *models.PermissionDef, resource string) error
}
