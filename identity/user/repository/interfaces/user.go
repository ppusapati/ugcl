package repointerfaces

import (
	"context"

	"p9e.in/ugcl/identity/user/models"
)

type UserRepository interface {
	Exists(ctx context.Context, search *models.UserSearchCriteria) (bool, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int32) (*models.User, error)
	GetByUUID(ctx context.Context, uuid string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error)
	AssignToUser(ctx context.Context, userID, roleID string) error
	RevokeFromUser(ctx context.Context, userID, roleID string) error
	ListUserRoles(ctx context.Context, userID string) ([]*models.Role, error)
	GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error)

	AssignPermissionDefToUser(ctx context.Context, userID string, def *models.PermissionDef, resource string) error
}
