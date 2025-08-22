// handlers/role_handler.go
package handler

import (
	"context"

	rolepb "p9e.in/ugcl/identity/api/v2/role"
	"p9e.in/ugcl/identity/mappers"
	"p9e.in/ugcl/identity/services"

	"github.com/golang/protobuf/ptypes/empty"
)

type RoleHandler struct {
	// rolepb.UnimplementedRoleServiceServer
	srvc services.IRoleService
}

func NewRoleHandler(service services.IRoleService) *RoleHandler {
	return &RoleHandler{srvc: service}
}

func (h *RoleHandler) Create(ctx context.Context, req *rolepb.Role) (*rolepb.Role, error) {
	role, err := h.srvc.Create(ctx, mappers.RoleProtoToModel(req))
	if err != nil {
		return nil, err
	}
	return mappers.RoleModelToProto(role), nil
}

func (h *RoleHandler) Update(ctx context.Context, req *rolepb.UpdateRoleRequest) (*rolepb.Role, error) {
	role := mappers.RoleProtoToModel(req.GetRole())
	updated, err := h.srvc.Update(ctx, role, req.GetUpdateMask())
	if err != nil {
		return nil, err
	}
	return mappers.RoleModelToProto(updated), nil
}

func (h *RoleHandler) Delete(ctx context.Context, req *rolepb.RoleIdentifier) (*empty.Empty, error) {
	err := h.srvc.Delete(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (h *RoleHandler) GetByID(ctx context.Context, req *rolepb.RoleIdentifier) (*rolepb.Role, error) {
	role, err := h.srvc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return mappers.RoleModelToProto(role), nil
}

func (h *RoleHandler) GetAll(ctx context.Context, req *rolepb.ListRolesRequest) (*rolepb.ListRolesResponse, error) {
	roles, err := h.srvc.List(ctx)
	if err != nil {
		return nil, err
	}

	var protoRoles []*rolepb.Role
	for _, r := range roles {
		protoRoles = append(protoRoles, mappers.RoleModelToProto(r))
	}

	return &rolepb.ListRolesResponse{
		Roles: protoRoles,
		// TotalSize: int32(total),
	}, nil
}

func (h *RoleHandler) GetWithPermissions(ctx context.Context, req *rolepb.RoleIdentifier) (*rolepb.RoleWithPermissions, error) {
	role, perms, err := h.srvc.GetWithPermissions(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &rolepb.RoleWithPermissions{
		Role:        mappers.RoleModelToProto(role),
		Permissions: mappers.PermissionsToProtoList(perms),
	}, nil
}
