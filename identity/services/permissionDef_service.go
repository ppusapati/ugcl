package services

import (
	"context"

	"p9e.in/ugcl/identity/models"
	"p9e.in/ugcl/identity/uow"
)

var _ IPermissionDefService = (*PermissionDefService)(nil)

type IPermissionDefService interface {
	Create(ctx context.Context, def *models.PermissionDef) (*models.PermissionDef, error)
	Update(ctx context.Context, def *models.PermissionDef) (*models.PermissionDef, error)
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*models.PermissionDef, error)
	GetByName(ctx context.Context, name string) (*models.PermissionDef, error)
	List(ctx context.Context) ([]*models.PermissionDef, error)
	ListByRole(ctx context.Context, roleID string) ([]*models.PermissionDef, error)
	AssignPermissionDefToRoleName(ctx context.Context, roleID string, name []string) error
	RevokeFromRole(ctx context.Context, roleID string, name []string) error
}

type PermissionDefService struct {
	factory uow.UnitOfWorkFactory
}

func NewPermissionDefService(factory uow.UnitOfWorkFactory) IPermissionDefService {
	return &PermissionDefService{factory: factory}
}

func (s *PermissionDefService) Create(ctx context.Context, def *models.PermissionDef) (*models.PermissionDef, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	if err := uow.PermissionDefRepo().Create(ctx, def); err != nil {
		return nil, err
	}
	return def, nil
}

func (s *PermissionDefService) Update(ctx context.Context, def *models.PermissionDef) (*models.PermissionDef, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	var updated *models.PermissionDef
	d, err := uow.PermissionDefRepo().Update(ctx, def)
	if err != nil {
		return nil, err
	}
	updated = d
	return updated, nil
}

func (s *PermissionDefService) Delete(ctx context.Context, uuid string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	return uow.PermissionDefRepo().Delete(ctx, uuid)
}

func (s *PermissionDefService) GetAll(ctx context.Context) ([]*models.PermissionDef, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.PermissionDefRepo().ListAll(ctx)
}

func (s *PermissionDefService) List(ctx context.Context) ([]*models.PermissionDef, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.PermissionDefRepo().ListAll(ctx)
}

func (s *PermissionDefService) GetByName(ctx context.Context, name string) (*models.PermissionDef, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.PermissionDefRepo().GetByName(ctx, name)
}

func (s *PermissionDefService) AssignPermissionDefToRoleName(ctx context.Context, roleID string, name []string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	for _, n := range name {
		if err := uow.PermissionDefRepo().AssignPermissionDefToRoleName(ctx, roleID, n); err != nil {
			return err
		}
	}
	return nil
}

func (s *PermissionDefService) ListByRole(ctx context.Context, roleID string) ([]*models.PermissionDef, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.PermissionDefRepo().ListByRole(ctx, roleID)
}

func (s *PermissionDefService) RevokeFromRole(ctx context.Context, roleID string, name []string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	for _, n := range name {
		if err := uow.PermissionDefRepo().RevokeFromRole(ctx, roleID, n); err != nil {
			return err
		}
	}
	return nil
}
