// uow_helpers.go
package helper

import (
	"context"
	"fmt"

	"p9e.in/ugcl/identity/user/models"
	"p9e.in/ugcl/identity/user/uow"
)

// AssignUserToRole links a user to a role using UOW

// AssignRolePermission links a role to a permission def using UOW
func AssignRolePermission(ctx context.Context, u uow.UnitOfWork, roleID, defName string) error {
	repo := u.PermissionDefRepo()
	if err := repo.AssignPermissionDefToRoleName(ctx, roleID, defName); err != nil {
		u.Rollback(ctx)
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}
	return nil
}

// ListRolesForUser retrieves roles for a user via join table
func ListRolesForUser(ctx context.Context, u uow.UnitOfWork, userID string) ([]*models.Role, error) {
	repo := u.UserRepo()
	roles, err := repo.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}
	return roles, nil
}
