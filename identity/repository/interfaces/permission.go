package repointerfaces

import (
	"context"

	"p9e.in/ugcl/identity/models"
)

type PermissionRepository interface {
	GetBySubject(ctx context.Context, subject string) ([]*models.Permission, error)
	Check(ctx context.Context, namespace, resource, action, subject string) (*models.Permission, error)
	Delete(ctx context.Context, namespace, resource, action, subject string) error
	ListAll(ctx context.Context) ([]*models.Permission, error)
	Grant(ctx context.Context, permission *models.Permission) error
	Revoke(ctx context.Context, namespace, resource, action, subject string) error
}
