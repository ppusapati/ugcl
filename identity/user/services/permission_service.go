package services

import (
	"context"
	"fmt"

	"p9e.in/ugcl/identity/user/helper"
	"p9e.in/ugcl/identity/user/mappers"
	"p9e.in/ugcl/identity/user/models"
	"p9e.in/ugcl/identity/user/uow"

	pb "p9e.in/ugcl/identity/user/api/v2/permission"
	"github.com/google/uuid"
)

var _ IPermissionService = (*PermissionService)(nil)

type IPermissionService interface {
	// Legacy model-based methods (for backward compatibility)
	GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error)
	CheckPermission(ctx context.Context, userID, namespace, resource, action string) (*models.Effect, error)
	GrantPermission(ctx context.Context, permission *models.Permission) error
	RevokePermission(ctx context.Context, userID, namespace, resource, action string) error
	GetPermissions(ctx context.Context, subject []string) ([]*models.Permission, error)
	GetRolePermissions(ctx context.Context, uuid uuid.UUID) (*models.Role, []*models.Permission, error)

	// New proto-based methods (enhanced permission system)
	GetPermissionsProto(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error)
	ListPermissionsByRole(ctx context.Context, req *pb.RoleIdentifier) (*pb.ListPermissionsResponse, error)
	ListPermissionsByUser(ctx context.Context, req *pb.UserIdentifier) (*pb.ListPermissionsResponse, error)
}

type PermissionService struct {
	factory uow.UnitOfWorkFactory
}

func NewPermissionService(factory uow.UnitOfWorkFactory) IPermissionService {
	return &PermissionService{factory: factory}
}

func (s *PermissionService) GetPermissions(ctx context.Context, subjects []string) ([]*models.Permission, error) {

	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	allPerms := []*models.Permission{}
	for _, subj := range subjects {
		perms, err := uow.PermissionRepo().GetBySubject(ctx, subj)
		if err != nil {
			return nil, err
		}
		allPerms = append(allPerms, perms...)
	}

	return allPerms, nil
}

func (s *PermissionService) GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	return uow.UserRepo().GetUserPermissions(ctx, userID)
}
func (s *PermissionService) GetRolePermissions(ctx context.Context, uuid uuid.UUID) (*models.Role, []*models.Permission, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer uow.Rollback(ctx)

	return uow.RoleRepo().ListPermissionsForRole(ctx, uuid.String())
}

func (s *PermissionService) CheckPermission(ctx context.Context, subject, namespace, resource, action string) (*models.Effect, error) {
	_, _, err := helper.ParseSubjectSafe(subject)
	if err != nil {
		return nil, err
	}

	// You can use prefix if needed later (e.g., logging or resolving differently)
	// But for now, just pass the raw subject ID to permission check
	return s.checkEnforcement(ctx, subject, namespace, resource, action)
}

func (s *PermissionService) GrantPermission(ctx context.Context, permission *models.Permission) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	return uow.PermissionRepo().Grant(ctx, permission)
}

func (s *PermissionService) RevokePermission(ctx context.Context, subject, namespace, resource, action string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	return uow.PermissionRepo().Revoke(ctx, subject, namespace, resource, action)
}

func (s *PermissionService) AssignDefToUser(ctx context.Context, userID, defName, resource string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	def, err := uow.PermissionDefRepo().GetByName(ctx, defName)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}
	return uow.UserRepo().AssignPermissionDefToUser(ctx, userID, def, resource)

}

func (s *PermissionService) AssignDefToRole(ctx context.Context, roleID, defName, resource string) error {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	def, err := uow.PermissionDefRepo().GetByName(ctx, defName)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}
	return uow.PermissionDefRepo().AssignPermissionDefToRole(ctx, roleID, def, resource)
}

func (s *PermissionService) checkEnforcement(ctx context.Context, subjectID, namespace, resource, action string) (*models.Effect, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	perm, err := uow.PermissionRepo().Check(ctx, namespace, resource, action, subjectID)
	if err != nil {
		return nil, err
	}
	eff := models.EffectUnknown
	if perm != nil {
		eff = perm.Effect
	}
	return &eff, nil
}

// Proto-based permission methods

// GetPermissionsProto returns permissions for subjects using proto format
func (s *PermissionService) GetPermissionsProto(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	uow, err := s.factory.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	allPerms := []*models.Permission{}
	for _, subj := range req.Subjects {
		perms, err := uow.PermissionRepo().GetBySubject(ctx, subj)
		if err != nil {
			return nil, err
		}
		allPerms = append(allPerms, perms...)
	}

	// Convert to proto
	protoPerms := make([]*pb.Permission, len(allPerms))
	for i, perm := range allPerms {
		protoPerms[i] = mappers.PermissionModelToProto(perm)
	}

	return &pb.ListPermissionsResponse{
		Permissions: protoPerms,
	}, nil
}

// ListPermissionsByRole returns all permissions for a role
func (s *PermissionService) ListPermissionsByRole(ctx context.Context, req *pb.RoleIdentifier) (*pb.ListPermissionsResponse, error) {
	uid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	_, perms, err := s.GetRolePermissions(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Convert to proto
	protoPerms := make([]*pb.Permission, len(perms))
	for i, perm := range perms {
		protoPerms[i] = mappers.PermissionModelToProto(perm)
	}

	return &pb.ListPermissionsResponse{
		Permissions: protoPerms,
	}, nil
}

// ListPermissionsByUser returns all permissions for a user
func (s *PermissionService) ListPermissionsByUser(ctx context.Context, req *pb.UserIdentifier) (*pb.ListPermissionsResponse, error) {
	var userID string
	if req.Uuid != "" {
		userID = req.Uuid
	} else {
		return nil, fmt.Errorf("user ID required")
	}

	perms, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to proto
	protoPerms := make([]*pb.Permission, len(perms))
	for i, perm := range perms {
		protoPerms[i] = mappers.PermissionModelToProto(perm)
	}

	return &pb.ListPermissionsResponse{
		Permissions: protoPerms,
	}, nil
}
