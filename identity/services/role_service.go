package services

import (
	"context"
	"fmt"

	"p9e.in/ugcl/identity/models"
	"p9e.in/ugcl/identity/uow"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ IRoleService = (*RoleService)(nil)

type IRoleService interface {
	Create(ctx context.Context, role *models.Role) (*models.Role, error)
	Update(ctx context.Context, role *models.Role, mask *fieldmaskpb.FieldMask) (*models.Role, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.Role, error)
	List(ctx context.Context) ([]*models.Role, error)
	GetWithPermissions(ctx context.Context, roleID string) (*models.Role, []*models.Permission, error)
}

type RoleService struct {
	factory uow.UnitOfWorkFactory
}

func NewRoleService(factory uow.UnitOfWorkFactory) IRoleService {
	return &RoleService{factory: factory}
}

func (s *RoleService) Create(ctx context.Context, role *models.Role) (*models.Role, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	if err := uow.RoleRepo().Create(ctx, role); err != nil {
		return nil, err
	}
	if err := uow.Commit(ctx); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) Get(ctx context.Context, id string) (*models.Role, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.RoleRepo().GetByID(ctx, id)
}

func (s *RoleService) Delete(ctx context.Context, id string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	if err := uow.RoleRepo().Delete(ctx, id); err != nil {
		return err
	}
	return uow.Commit(ctx)
}

func (s *RoleService) Update(ctx context.Context, role *models.Role, mask *fieldmaskpb.FieldMask) (*models.Role, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	existing, err := uow.RoleRepo().GetByID(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	for _, path := range mask.Paths {
		switch path {
		case "name":
			existing.Name = role.Name
		case "metadata":
			existing.Metadata = role.Metadata
		// Add other fields as needed
		default:
			return nil, fmt.Errorf("unsupported field: %s", path)
		}
	}

	if err := uow.RoleRepo().Update(ctx, existing); err != nil {
		return nil, err
	}
	if err := uow.Commit(ctx); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *RoleService) List(ctx context.Context) ([]*models.Role, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.RoleRepo().List(ctx)
}

func (s *RoleService) GetWithPermissions(ctx context.Context, roleId string) (*models.Role, []*models.Permission, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer uow.Rollback(ctx)
	return uow.RoleRepo().ListPermissionsForRole(ctx, roleId)
}
