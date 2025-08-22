package handler

import (
	"context"

	"p9e.in/ugcl/identity/api/v2/permission"

	permspb "p9e.in/ugcl/identity/api/v2/permission"
	"p9e.in/ugcl/identity/mappers"
	services "p9e.in/ugcl/identity/services"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PermissionHandler struct {
	// permspb.UnimplementedPermissionServiceServer
	Service services.IPermissionService
}

func NewPermissionHandler(service services.IPermissionService) *PermissionHandler {
	return &PermissionHandler{Service: service}
}

func (h *PermissionHandler) Grant(ctx context.Context, req *permspb.Permission) (*emptypb.Empty, error) {
	return nil, h.Service.GrantPermission(ctx, mappers.PermissionProtoToModel(req))
}

func (h *PermissionHandler) Revoke(ctx context.Context, req *permspb.RevokePermissionRequest) (*emptypb.Empty, error) {
	return nil, h.Service.RevokePermission(ctx, req.Subject, req.Namespace, req.Resource, req.Action)
}

func (h *PermissionHandler) Check(ctx context.Context, req *permspb.CheckPermissionRequest) (*permspb.CheckPermissionResponse, error) {
	effect, err := h.Service.CheckPermission(ctx, req.Subject, req.Namespace, req.Resource, req.Action)
	if err != nil {
		return nil, err
	}

	return &permission.CheckPermissionResponse{
		Effect: permission.Effect(*effect), // cast from model to proto enum
	}, nil
}

func (h *PermissionHandler) GetPermissions(ctx context.Context, req *permspb.ListPermissionsRequest) (*permspb.ListPermissionsResponse, error) {
	allPerms, err := h.Service.GetPermissions(ctx, req.Subjects)
	if err != nil {
		return nil, err
	}
	return &permspb.ListPermissionsResponse{
		Permissions: mappers.PermissionsToProtoList(allPerms)}, nil
}

func (h *PermissionHandler) ListPermissionByRole(ctx context.Context, req *permspb.RoleIdentifier) (*permspb.ListPermissionsResponse, error) {
	uid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, err
	}
	_, res, err := h.Service.GetRolePermissions(ctx, uid)
	if err != nil {
		return nil, err
	}
	var protoPerms []*permission.Permission
	for _, perm := range res {
		protoPerms = append(protoPerms, mappers.PermissionModelToProto(perm))
	}
	return &permspb.ListPermissionsResponse{Permissions: protoPerms}, nil
}

func (h *PermissionHandler) ListPermissionByUser(ctx context.Context, req *permspb.UserIdentifier) (*permspb.ListPermissionsResponse, error) {
	res, err := h.Service.GetUserPermissions(ctx, req.Uuid)
	if err != nil {
		return nil, err
	}
	var protoPerms []*permission.Permission
	for _, perm := range res {
		protoPerms = append(protoPerms, mappers.PermissionModelToProto(perm))
	}
	return &permspb.ListPermissionsResponse{Permissions: protoPerms}, nil
}
