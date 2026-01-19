package handlers

import (
	"context"

	"connectrpc.com/connect"

	pb "p9e.in/ugcl/identity/entity/api/v1"
	"p9e.in/ugcl/identity/entity/api/v1/entityv1connect"
	"p9e.in/ugcl/identity/entity/services"
)

// EntityHandler implements the Connect RPC handlers for EntityService
type EntityHandler struct {
	entityService            services.IEntityService
	entityRoleBindingService services.IEntityRoleBindingService
}

// NewEntityHandler creates a new entity handler
func NewEntityHandler(
	entityService services.IEntityService,
	entityRoleBindingService services.IEntityRoleBindingService,
) entityv1connect.EntityServiceHandler {
	return &EntityHandler{
		entityService:            entityService,
		entityRoleBindingService: entityRoleBindingService,
	}
}

// Entity CRUD handlers

func (h *EntityHandler) CreateEntity(
	ctx context.Context,
	req *connect.Request[pb.CreateEntityRequest],
) (*connect.Response[pb.Entity], error) {
	entity, err := h.entityService.CreateEntity(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(entity), nil
}

func (h *EntityHandler) UpdateEntity(
	ctx context.Context,
	req *connect.Request[pb.UpdateEntityRequest],
) (*connect.Response[pb.Entity], error) {
	entity, err := h.entityService.UpdateEntity(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(entity), nil
}

func (h *EntityHandler) GetEntity(
	ctx context.Context,
	req *connect.Request[pb.GetEntityRequest],
) (*connect.Response[pb.Entity], error) {
	entity, err := h.entityService.GetEntity(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(entity), nil
}

func (h *EntityHandler) GetEntityByUserId(
	ctx context.Context,
	req *connect.Request[pb.GetEntityByUserIdRequest],
) (*connect.Response[pb.Entity], error) {
	entity, err := h.entityService.GetEntityByUserId(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(entity), nil
}

func (h *EntityHandler) GetEntityByReference(
	ctx context.Context,
	req *connect.Request[pb.GetEntityByReferenceRequest],
) (*connect.Response[pb.Entity], error) {
	entity, err := h.entityService.GetEntityByReference(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(entity), nil
}

func (h *EntityHandler) ListEntities(
	ctx context.Context,
	req *connect.Request[pb.ListEntitiesRequest],
) (*connect.Response[pb.ListEntitiesResponse], error) {
	response, err := h.entityService.ListEntities(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *EntityHandler) DeleteEntity(
	ctx context.Context,
	req *connect.Request[pb.DeleteEntityRequest],
) (*connect.Response[pb.DeleteEntityResponse], error) {
	response, err := h.entityService.DeleteEntity(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

// Entity role binding handlers

func (h *EntityHandler) CreateEntityRoleBinding(
	ctx context.Context,
	req *connect.Request[pb.CreateEntityRoleBindingRequest],
) (*connect.Response[pb.EntityRoleBinding], error) {
	binding, err := h.entityRoleBindingService.CreateEntityRoleBinding(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(binding), nil
}

func (h *EntityHandler) GetEntityRoleBindings(
	ctx context.Context,
	req *connect.Request[pb.GetEntityRoleBindingsRequest],
) (*connect.Response[pb.GetEntityRoleBindingsResponse], error) {
	response, err := h.entityRoleBindingService.GetEntityRoleBindings(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *EntityHandler) DeleteEntityRoleBinding(
	ctx context.Context,
	req *connect.Request[pb.DeleteEntityRoleBindingRequest],
) (*connect.Response[pb.DeleteEntityRoleBindingResponse], error) {
	response, err := h.entityRoleBindingService.DeleteEntityRoleBinding(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}
