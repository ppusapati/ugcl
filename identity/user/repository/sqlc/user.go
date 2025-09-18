package sqlcrepo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	sqlc "p9e.in/ugcl/identity/user/db/sqlc/generated"
	"p9e.in/ugcl/identity/user/mappers"
	"p9e.in/ugcl/identity/user/models"
)

type SQLCUserRepo struct {
	queries *sqlc.Queries
	// db      sqlc.DBTX
}

func NewSQLCUserRepo(q *sqlc.Queries) *SQLCUserRepo {
	return &SQLCUserRepo{queries: q}
}

func (r *SQLCUserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Uuid:               user.Uuid,
		Username:           user.Username,
		NormalizedUsername: user.NormalizedUsername,
		Fullname:           user.Fullname,
		Phone:              user.Phone,
		PhoneVerified:      &user.PhoneConfirmed,
		Email:              user.Email,
		NormalizedEmail:    user.NormalizedEmail,
		EmailVerified:      &user.EmailConfirmed,
		PasswordHash:       user.PasswordHash,
		Gender:             user.Gender,
		RolesCache:         user.TenantRoles,
		Avatar:             user.Avatar,
		TwoFactorEnabled:   &user.TwoFactorEnabled,
		Salt:               user.Salt,
		TwoFactorSecret:    user.TwoFactorSecret,
		IsActive:           user.IsActive,
		// Status:            user.,
	})
	return user, err
}

func (r *SQLCUserRepo) GetByID(ctx context.Context, id int32) (*models.User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, sqlc.GetUserByIDParams{
		ID: int64(id),
	})
	if err != nil {
		return nil, err
	}
	return mappers.UserSQLCToModel(dbUser), nil
}

func (r *SQLCUserRepo) GetByUUID(ctx context.Context, uuidStr string) (*models.User, error) {
	// Parse string UUID to uuid.UUID
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, err
	}

	dbUser, err := r.queries.GetUserByUUID(ctx, sqlc.GetUserByUUIDParams{
		Uuid: parsedUUID,
	})
	if err != nil {
		return nil, err
	}
	return mappers.UserSQLCToModel(dbUser), nil
}

func (r *SQLCUserRepo) Update(ctx context.Context, user *models.User) error {
	return r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		Username: user.Username,
		Email:    user.Email,
		IsActive: user.IsActive,
	})
}

func (r *SQLCUserRepo) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteUser(ctx, sqlc.DeleteUserParams{
		ID: int64(id),
	})
}

func (r *SQLCUserRepo) List(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error) {
	users, err := r.queries.ListUsers(ctx)
	if err != nil {
		fmt.Println("hey", err)
		return nil, err
	}
	fmt.Println(users)
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = mappers.UserSQLCToModel(users[i])
	}
	return result, nil
}

func (r *SQLCUserRepo) Exists(ctx context.Context, filter *models.UserSearchCriteria) (bool, error) {
	params := sqlc.UserExistsParams{
		Username: *filter.UsernameSearch,
		Email:    *filter.EmailSearch,
		Phone:    *filter.PhoneSearch,
		// Fullname: *filter.FullnameSearch,
	}

	return r.queries.UserExists(ctx, params)
}

func (r *SQLCUserRepo) AssignToUser(ctx context.Context, userID, roleID string) error {
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	roleIDInt, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}
	return r.queries.AssignRoleToUser(ctx, sqlc.AssignRoleToUserParams{
		UserID: userIDInt,
		RoleID: roleIDInt,
	})
}

func (r *SQLCUserRepo) RevokeFromUser(ctx context.Context, userID, roleID string) error {
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	roleIDInt, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}
	return r.queries.RemoveRoleFromUser(ctx, sqlc.RemoveRoleFromUserParams{
		UserID: userIDInt,
		RoleID: roleIDInt,
	})
}

func (r *SQLCUserRepo) ListUserRoles(ctx context.Context, userID string) ([]*models.Role, error) {
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	roles, err := r.queries.GetRolesByUserUUID(ctx, sqlc.GetRolesByUserUUIDParams{
		UserID: userIDInt,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*models.Role, len(roles))
	for i := range roles {
		result[i], _ = mappers.RoleSQLCToModel(roles[i])
	}
	return result, nil
}

func (r *SQLCUserRepo) GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error) {
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	permissions, err := r.queries.GetUserPermissions(ctx, sqlc.GetUserPermissionsParams{
		UserID: userIDInt,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*models.Permission, len(permissions))
	for i := range permissions {
		result[i] = mappers.PermissionSQLCToModel(permissions[i])
	}
	return result, nil
}
func (r *SQLCUserRepo) AssignPermissionDefToUser(ctx context.Context, userID string, def *models.PermissionDef, resource string) error {
	perm := mappers.MaterializeDefToPermission(def, userID, resource, nil)

	return r.queries.GrantPermission(ctx, sqlc.GrantPermissionParams{
		Namespace: perm.Namespace,
		Resource:  perm.Resource,
		Action:    perm.Action,
		Subject:   perm.Subject,
		Effect:    mappers.ModelToDBEffectInterface(perm.Effect),
	})
}
