package handlers

import (
	"context"

	pb "p9e.in/ugcl/contractors/api/v2/contractor"
	"p9e.in/ugcl/contractors/mappers"
	"p9e.in/ugcl/contractors/services"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ContractorHandler struct {
	srvc services.IContractorService
}

func NewContractorHandler(srvc services.IContractorService) *ContractorHandler {
	return &ContractorHandler{
		srvc: srvc,
	}
}

// CreateContractor implements ContractorService.CreateContractor
func (h *ContractorHandler) CreateContractor(
	ctx context.Context,
	req *connect.Request[pb.CreateContractorRequest],
) (*connect.Response[pb.Contractor], error) {

	// Validate UUID
	if req.Msg.Contractor.Id == "" {
		req.Msg.Contractor.Id = uuid.New().String()
	}

	// Map from proto to DB struct
	dbContractor, err := mappers.ProtoToDBContractor(req.Msg.Contractor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	// dbUser := umappers.UserProtoToModel(req.Msg.User)
	// if err != nil {
	// 	return nil, connect.NewError(connect.CodeInvalidArgument, err)
	// }

	// Call repository
	if _, err := h.srvc.Create(ctx, dbContractor, req.Msg.User); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map back to proto
	contractorProto, err := mappers.DBToProtoContractor(dbContractor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(contractorProto), nil
}

// UpdateContractor implements ContractorService.UpdateContractor
func (h *ContractorHandler) UpdateContractor(
	ctx context.Context,
	req *connect.Request[pb.UpdateContractorRequest],
) (*connect.Response[pb.Contractor], error) {

	// Fetch existing
	existing, err := h.srvc.GetByID(ctx, req.Msg.Contractor.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Apply field mask
	updated := mappers.ApplyFieldMask(existing, req.Msg.Contractor, req.Msg.UpdateMask)

	// Persist
	if _, err := h.srvc.Update(ctx, updated); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map back
	protoResp, err := mappers.DBToProtoContractor(updated)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoResp), nil
}

// GetContractor implements ContractorService.GetContractor
func (h *ContractorHandler) GetContractor(
	ctx context.Context,
	req *connect.Request[pb.ContractorIdentifier],
) (*connect.Response[pb.Contractor], error) {

	dbContractor, err := h.srvc.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	protoContractor, err := mappers.DBToProtoContractor(dbContractor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoContractor), nil
}

// ListContractors implements ContractorService.ListContractors
func (h *ContractorHandler) ListContractors(
	ctx context.Context,
	req *connect.Request[pb.ListContractorsRequest],
) (*connect.Response[pb.ListContractorsResponse], error) {

	dbList, total, err := h.srvc.List(ctx, req.Msg.PageSize, req.Msg.PageOffset, req.Msg.Filter, req.Msg.Sort)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoList := make([]*pb.Contractor, len(dbList))
	for i, c := range dbList {
		pc, err := mappers.DBToProtoContractor(&c)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoList[i] = pc
	}

	return connect.NewResponse(&pb.ListContractorsResponse{
		Contractors: protoList,
		TotalCount:  int32(total),
	}), nil
}

// DeleteContractor implements ContractorService.DeleteContractor
func (h *ContractorHandler) DeleteContractor(
	ctx context.Context,
	req *connect.Request[pb.ContractorIdentifier],
) (*connect.Response[emptypb.Empty], error) {

	if err := h.srvc.Delete(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetContractorByUserId implements ContractorService.GetContractorByUserId
func (h *ContractorHandler) GetContractorByUserId(
	ctx context.Context,
	req *connect.Request[pb.UserIdentifier],
) (*connect.Response[pb.Contractor], error) {

	dbContractor, err := h.srvc.GetByPersonID(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	protoContractor, err := mappers.DBToProtoContractor(dbContractor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoContractor), nil
}
