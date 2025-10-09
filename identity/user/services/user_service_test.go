package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/user/models"
	repointerfaces "p9e.in/ugcl/identity/user/repository/interfaces"
	"p9e.in/ugcl/identity/user/uow"
)

// -------------------- Mocks --------------------

type mockUOWFactory struct{ u *mockUOW }

func (f *mockUOWFactory) Begin(ctx context.Context) (uow.UnitOfWork, error) { return f.u, nil }

// Ensure mockUOWFactory implements interface
var _ uow.UnitOfWorkFactory = (*mockUOWFactory)(nil)

type mockUOW struct {
	userRepo       *mockUserRepo
	commitCalled   bool
	rollbackCalled bool
	commitErr      error
}

func (m *mockUOW) UserRepo() repointerfaces.UserRepository                   { return m.userRepo }
func (m *mockUOW) RoleRepo() repointerfaces.RoleRepository                   { return nil }
func (m *mockUOW) PermissionRepo() repointerfaces.PermissionRepository       { return nil }
func (m *mockUOW) PermissionDefRepo() repointerfaces.PermissionDefRepository { return nil }
func (m *mockUOW) Commit(ctx context.Context) error                          { m.commitCalled = true; return m.commitErr }
func (m *mockUOW) Rollback(ctx context.Context) error                        { m.rollbackCalled = true; return nil }

// Ensure mockUOW implements interface
var _ uow.UnitOfWork = (*mockUOW)(nil)

type mockUserRepo struct {
	ExistsFunc       func(ctx context.Context, search *models.UserSearchCriteria) (bool, error)
	CreateFunc       func(ctx context.Context, user *models.User) (*models.User, error)
	UpdateFunc       func(ctx context.Context, user *models.User) error
	DeleteFunc       func(ctx context.Context, id int32) error
	GetByIDFunc      func(ctx context.Context, id int32) (*models.User, error)
	GetByUUIDFunc    func(ctx context.Context, uuid string) (*models.User, error)
	ListFunc         func(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error)
	AssignToUserFunc func(ctx context.Context, userID, roleID string) error
	RevokeFromUserFn func(ctx context.Context, userID, roleID string) error
	ListUserRolesFn  func(ctx context.Context, userID string) ([]*models.Role, error)
	GetUserPermsFn   func(ctx context.Context, userID string) ([]*models.Permission, error)
	AssignPermDefFn  func(ctx context.Context, userID string, def *models.PermissionDef, resource string) error

	CreateCalled bool
	ExistsCalled bool

	LastCreateArg *models.User
}

func (m *mockUserRepo) Exists(ctx context.Context, search *models.UserSearchCriteria) (bool, error) {
	m.ExistsCalled = true
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, search)
	}
	return false, nil
}
func (m *mockUserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) {
	m.CreateCalled = true
	m.LastCreateArg = user
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return user, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int32) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	if m.GetByUUIDFunc != nil {
		return m.GetByUUIDFunc(ctx, uuid)
	}
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id int32) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
func (m *mockUserRepo) List(ctx context.Context, filter *models.UserSearchCriteria) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return nil, nil
}
func (m *mockUserRepo) AssignToUser(ctx context.Context, userID, roleID string) error {
	if m.AssignToUserFunc != nil {
		return m.AssignToUserFunc(ctx, userID, roleID)
	}
	return nil
}
func (m *mockUserRepo) RevokeFromUser(ctx context.Context, userID, roleID string) error {
	if m.RevokeFromUserFn != nil {
		return m.RevokeFromUserFn(ctx, userID, roleID)
	}
	return nil
}
func (m *mockUserRepo) ListUserRoles(ctx context.Context, userID string) ([]*models.Role, error) {
	if m.ListUserRolesFn != nil {
		return m.ListUserRolesFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockUserRepo) GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error) {
	if m.GetUserPermsFn != nil {
		return m.GetUserPermsFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockUserRepo) AssignPermissionDefToUser(ctx context.Context, userID string, def *models.PermissionDef, resource string) error {
	if m.AssignPermDefFn != nil {
		return m.AssignPermDefFn(ctx, userID, def, resource)
	}
	return nil
}

// Ensure mockUserRepo implements interface
var _ repointerfaces.UserRepository = (*mockUserRepo)(nil)

// Helpers to allocate pointers
func strptr(s string) *string { return &s }

// -------------------- Tests --------------------

func TestRegisterUser_AlreadyExists(t *testing.T) {
	usrRepo := &mockUserRepo{
		ExistsFunc: func(ctx context.Context, search *models.UserSearchCriteria) (bool, error) {
			return true, nil
		},
	}
	uowMock := &mockUOW{userRepo: usrRepo}
	factory := &mockUOWFactory{u: uowMock}
	svc := NewUserService(factory)

	user := &models.User{Email: strptr("taken@example.com"), Username: strptr("existing"), Phone: strptr("1234567890"), PasswordHash: "plaintext"}
	_, err := svc.RegisterUser(context.Background(), user)
	if err == nil {
		t.Fatalf("expected error for existing user, got nil")
	}
	if usrRepo.CreateCalled {
		t.Errorf("expected Create not to be called when user exists")
	}
	if uowMock.commitCalled {
		t.Errorf("expected Commit not to be called on exists error")
	}
	if !uowMock.rollbackCalled {
		t.Errorf("expected Rollback to be called via defer")
	}
}

func TestRegisterUser_Success_HashesPasswordAndCommits(t *testing.T) {
	plain := "P@ssword123"
	usrRepo := &mockUserRepo{
		ExistsFunc: func(ctx context.Context, search *models.UserSearchCriteria) (bool, error) { return false, nil },
		CreateFunc: func(ctx context.Context, u *models.User) (*models.User, error) {
			if u.PasswordHash == plain {
				t.Errorf("expected password to be hashed, still plaintext")
			}
			if len(u.Salt) == 0 {
				t.Errorf("expected salt to be set")
			}
			// Return created copy
			return &models.User{Uuid: uuid.New()}, nil
		},
	}
	uowMock := &mockUOW{userRepo: usrRepo}
	factory := &mockUOWFactory{u: uowMock}
	svc := NewUserService(factory)

	user := &models.User{Email: strptr("new@example.com"), Username: strptr("newuser"), Phone: strptr("9876543210"), PasswordHash: plain}
	created, err := svc.RegisterUser(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil {
		t.Fatalf("expected created user, got nil")
	}
	if !usrRepo.CreateCalled {
		t.Errorf("expected Create to be called")
	}
	if !uowMock.commitCalled {
		t.Errorf("expected Commit to be called on success")
	}
	if !uowMock.rollbackCalled {
		t.Errorf("expected Rollback to be called via defer even on success")
	}
}

func TestRegisterUser_CreateError(t *testing.T) {
	usrRepo := &mockUserRepo{
		ExistsFunc: func(ctx context.Context, search *models.UserSearchCriteria) (bool, error) { return false, nil },
		CreateFunc: func(ctx context.Context, u *models.User) (*models.User, error) {
			return nil, errors.New("create failed")
		},
	}
	uowMock := &mockUOW{userRepo: usrRepo}
	factory := &mockUOWFactory{u: uowMock}
	svc := NewUserService(factory)

	user := &models.User{Email: strptr("err@example.com"), Username: strptr("erruser"), Phone: strptr("555"), PasswordHash: "secret"}
	_, err := svc.RegisterUser(context.Background(), user)
	if err == nil {
		t.Fatalf("expected error from repository create, got nil")
	}
	if uowMock.commitCalled {
		t.Errorf("expected Commit not to be called on error")
	}
	if !uowMock.rollbackCalled {
		t.Errorf("expected Rollback to be called via defer")
	}
}
