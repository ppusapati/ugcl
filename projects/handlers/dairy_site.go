package handlers

import (
	"context"

	"p9e.in/ugcl/projects/api/v2/dairy_site"
	dairysitemapper "p9e.in/ugcl/projects/mappers"
	"p9e.in/ugcl/projects/services"

	"connectrpc.com/connect"
)

// ConnectHandler wraps your gRPC service for Connect protocol
type DairySiteHandler struct {
	// pb.UnimplementedDairySiteServiceServer
	svc services.IDairySiteService
}

func NewDairySiteHandler(svc services.IDairySiteService) *DairySiteHandler {
	return &DairySiteHandler{
		svc: svc,
	}
}

func (h *DairySiteHandler) GetAllDairySites(
	ctx context.Context,
	req *connect.Request[dairy_site.GetDairySiteRequest],
) (*connect.Response[dairy_site.GetDairySiteResponse], error) {

	sites, err := h.svc.GetAllWithUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if len(sites) == 0 {
		return connect.NewResponse(&dairy_site.GetDairySiteResponse{}), nil
	}

	return connect.NewResponse(dairysitemapper.DBWithUserToProto(sites[0])), nil
}

func (h *DairySiteHandler) CreateDairySite(
	ctx context.Context,
	req *connect.Request[dairy_site.CreateDairySiteRequest],
) (*connect.Response[dairy_site.CreateDairySiteResponse], error) {

	params := dairysitemapper.ProtoToCreateParams(req.Msg.DairySite)

	created, err := h.svc.Create(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&dairy_site.CreateDairySiteResponse{
		DairySite: dairysitemapper.DBToProto(created),
	}), nil
}
