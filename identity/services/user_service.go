package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"p9e.in/ugcl/packages/utils"

	"p9e.in/ugcl/identity/models"
	"p9e.in/ugcl/identity/uow"
)

var _ IUserService = (*UserService)(nil)

// IUserService defines the interface for user-related operations,
// including user registration, role assignments, and retrieval of
// user information. It provides methods to manage users, such as
// creating, updating, deleting, and resetting passwords. Additionally,
// it supports operations for fetching user roles and assigning or
// revoking roles from users.
type IUserService interface {
	RegisterUser(ctx context.Context, user *models.User) (*models.User, error)
	AssignRoles(ctx context.Context, userID string, roleIDs []string) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id int32) error
	ResetPassword(ctx context.Context, userID string, newHash string) error

	GetUserByID(ctx context.Context, id int32) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetByIdentifier(ctx context.Context, ident *models.UserIdentifier) (*models.User, error)

	ListUsers(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]*models.Role, error)

	AssignUserToRole(ctx context.Context, userID, roleID string) error
	RevokeUserFromRole(ctx context.Context, userID, roleId string) error
}

type UserService struct {
	factory uow.UnitOfWorkFactory
}

func NewUserService(factory uow.UnitOfWorkFactory) IUserService {
	return &UserService{factory: factory}
}

func (s *UserService) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	exists, err := uow.UserRepo().Exists(ctx, &models.UserSearchCriteria{
		EmailSearch:    user.Email,
		UsernameSearch: user.Username,
		PhoneSearch:    user.Phone,
	})
	if err != nil {
		return nil, fmt.Errorf("check existing user failed: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already exists with given email/username/phone")
	}

	hash, salt, err := utils.EncryptPassword(user.PasswordHash, nil)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}
	user.PasswordHash = hash
	user.Salt = salt
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	res, err := uow.UserRepo().Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	return res, uow.Commit(ctx)
}

func (s *UserService) AssignRoles(ctx context.Context, userID string, roleIDs []string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	for _, roleID := range roleIDs {
		if err := uow.UserRepo().AssignToUser(ctx, userID, roleID); err != nil {
			return fmt.Errorf("assign role %s failed: %w", roleID, err)
		}
	}

	return uow.Commit(ctx)
}

func (s *UserService) GetUserByID(ctx context.Context, id int32) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return uow.UserRepo().GetByID(ctx, id)
}

func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return uow.UserRepo().GetByUUID(ctx, uuid)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	user.UpdatedAt = time.Now()
	if err := uow.UserRepo().Update(ctx, user); err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	return uow.Commit(ctx)
}

func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	if err := uow.UserRepo().Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return uow.Commit(ctx)
}

func (s *UserService) ListUsers(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(filter)
	return uow.UserRepo().List(ctx, filter)
}

func (s *UserService) GetUserRoles(ctx context.Context, userID string) ([]*models.Role, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return uow.UserRepo().ListUserRoles(ctx, userID)
}

func (s *UserService) ResetPassword(ctx context.Context, userID string, newHash string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	user, err := uow.UserRepo().GetByUUID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	hash, salt, err := utils.EncryptPassword(newHash, nil)
	if err != nil {
		return fmt.Errorf("hashing password failed: %w", err)
	}
	user.PasswordHash = hash
	user.Salt = salt
	user.UpdatedAt = time.Now()

	if err := uow.UserRepo().Update(ctx, user); err != nil {
		return fmt.Errorf("reset password failed: %w", err)
	}

	return uow.Commit(ctx)
}

func (s *UserService) GetByIdentifier(ctx context.Context, ident *models.UserIdentifier) (*models.User, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	switch {
	case ident.Id != 0:
		return uow.UserRepo().GetByID(ctx, ident.Id)
	case ident.Uuid != "":
		return uow.UserRepo().GetByUUID(ctx, ident.Uuid)
	default:
		return nil, errors.New("must specify id or uuid")
	}
}

func (s *UserService) AssignUserToRole(ctx context.Context, userID, roleID string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	if err := uow.UserRepo().AssignToUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}
	return uow.Commit(ctx)
}

func (s *UserService) RevokeUserFromRole(ctx context.Context, userID, roleID string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	if err := uow.UserRepo().RevokeFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to revoke role to user: %w", err)
	}
	return uow.Commit(ctx)
}
