package repointerfaces

import (
	"context"

	"p9e.in/ugcl/identity/models"
)

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id string) (*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*models.Role, error)

	ListPermissionsForRole(ctx context.Context, roleID string) (*models.Role, []*models.Permission, error)
	// ListRolePermissions(ctx context.Context, roleID string) (any, any)

}
