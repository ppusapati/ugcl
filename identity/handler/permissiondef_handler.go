package handler

import (
	"context"

	"p9e.in/ugcl/identity/api/v2/permissiondef"
	"p9e.in/ugcl/identity/mappers"
	"p9e.in/ugcl/identity/services"

	"google.golang.org/protobuf/types/known/emptypb"
)

type PermissionDefHandler struct {
	// permissiondef.UnimplementedPermissionDefServiceServer
	Service services.IPermissionDefService
}

func NewPermissionDefHandler(service services.IPermissionDefService) *PermissionDefHandler {
	return &PermissionDefHandler{Service: service}
}

func (h *PermissionDefHandler) Create(ctx context.Context, req *permissiondef.PermissionDef) (*permissiondef.PermissionDef, error) {
	model := mappers.PermissionDefProtoToModel(req)
	created, err := h.Service.Create(ctx, model)
	if err != nil {
		return nil, err
	}
	return mappers.PermissionDefModelToProto(created), nil
}

func (h *PermissionDefHandler) Update(ctx context.Context, req *permissiondef.PermissionDef) (*permissiondef.PermissionDef, error) {
	model := mappers.PermissionDefProtoToModel(req)
	updated, err := h.Service.Update(ctx, model)
	if err != nil {
		return nil, err
	}
	return mappers.PermissionDefModelToProto(updated), nil
}

func (h *PermissionDefHandler) Delete(ctx context.Context, req *permissiondef.PermissionDefIdentifier) (*emptypb.Empty, error) {
	err := h.Service.Delete(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *PermissionDefHandler) GetByName(ctx context.Context, req *permissiondef.PermissionDefIdentifier) (*permissiondef.PermissionDef, error) {
	model, err := h.Service.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return mappers.PermissionDefModelToProto(model), nil
}

func (h *PermissionDefHandler) GetAll(ctx context.Context, _ *emptypb.Empty) (*permissiondef.ListPermissionDefsResponse, error) {
	items, err := h.Service.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return &permissiondef.ListPermissionDefsResponse{
		Items: mappers.PermissionDefsToProtoList(items),
	}, nil
}

func (h *PermissionDefHandler) ListByRole(ctx context.Context, req *permissiondef.RoleIdentifier) (*permissiondef.ListPermissionDefsResponse, error) {
	items, err := h.Service.ListByRole(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &permissiondef.ListPermissionDefsResponse{
		Items: mappers.PermissionDefsToProtoList(items),
	}, nil
}

func (h *PermissionDefHandler) AssignPermissionDefsToRole(ctx context.Context, req *permissiondef.AssignPermissionDefsRequest) (*emptypb.Empty, error) {
	if err := h.Service.AssignPermissionDefToRoleName(ctx, req.RoleId, req.Names); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *PermissionDefHandler) RevokePermissionDefsFromRole(ctx context.Context, req *permissiondef.RevokePermissionDefsRequest) (*emptypb.Empty, error) {
	if err := h.Service.RevokeFromRole(ctx, req.RoleId, req.Names); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
