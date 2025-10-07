package handlers

import (
	"context"

	pb "p9e.in/ugcl/vendors/api/v2/vendor"
	"p9e.in/ugcl/vendors/mappers"
	"p9e.in/ugcl/vendors/services"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type VendorHandler struct {
	srvc services.IVendorService
}

func NewVendorHandler(srvc services.IVendorService) *VendorHandler {
	return &VendorHandler{
		srvc: srvc,
	}
}

// CreateVendor implements VendorService.CreateVendor
func (h *VendorHandler) CreateVendor(
	ctx context.Context,
	req *connect.Request[pb.CreateVendorRequest],
) (*connect.Response[pb.Vendor], error) {

	// Validate UUID
	if req.Msg.Vendor.Id == "" {
		req.Msg.Vendor.Id = uuid.New().String()
	}

	// Map from proto to DB struct
	dbVendor, err := mappers.ProtoToDBVendor(req.Msg.Vendor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call service
	if _, err := h.srvc.Create(ctx, dbVendor, req.Msg.User); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map back to proto
	vendorProto, err := mappers.DBToProtoVendor(dbVendor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(vendorProto), nil
}

// UpdateVendor implements VendorService.UpdateVendor
func (h *VendorHandler) UpdateVendor(
	ctx context.Context,
	req *connect.Request[pb.UpdateVendorRequest],
) (*connect.Response[pb.Vendor], error) {

	// Fetch existing
	existing, err := h.srvc.GetByID(ctx, req.Msg.Vendor.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Apply field mask
	updated := mappers.ApplyFieldMask(existing, req.Msg.Vendor, req.Msg.UpdateMask)

	// Persist
	if _, err := h.srvc.Update(ctx, updated); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map back
	protoResp, err := mappers.DBToProtoVendor(updated)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoResp), nil
}

// GetVendor implements VendorService.GetVendor
func (h *VendorHandler) GetVendor(
	ctx context.Context,
	req *connect.Request[pb.VendorIdentifier],
) (*connect.Response[pb.Vendor], error) {

	dbVendor, err := h.srvc.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	protoVendor, err := mappers.DBToProtoVendor(dbVendor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoVendor), nil
}

// ListVendors implements VendorService.ListVendors
func (h *VendorHandler) ListVendors(
	ctx context.Context,
	req *connect.Request[pb.ListVendorsRequest],
) (*connect.Response[pb.ListVendorsResponse], error) {

	dbList, total, err := h.srvc.List(ctx, req.Msg.PageSize, req.Msg.PageOffset, req.Msg.Filter, req.Msg.Sort)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoList := make([]*pb.Vendor, len(dbList))
	for i, v := range dbList {
		pv, err := mappers.DBToProtoVendor(&v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoList[i] = pv
	}

	return connect.NewResponse(&pb.ListVendorsResponse{
		Vendors:    protoList,
		TotalCount: int32(total),
	}), nil
}

// DeleteVendor implements VendorService.DeleteVendor
func (h *VendorHandler) DeleteVendor(
	ctx context.Context,
	req *connect.Request[pb.VendorIdentifier],
) (*connect.Response[emptypb.Empty], error) {

	if err := h.srvc.Delete(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetVendorByUserId implements VendorService.GetVendorByUserId
func (h *VendorHandler) GetVendorByUserId(
	ctx context.Context,
	req *connect.Request[pb.UserIdentifier],
) (*connect.Response[pb.Vendor], error) {

	dbVendor, err := h.srvc.GetByPersonID(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	protoVendor, err := mappers.DBToProtoVendor(dbVendor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoVendor), nil
}

// GetVendorsByCategory implements VendorService.GetVendorsByCategory
func (h *VendorHandler) GetVendorsByCategory(
	ctx context.Context,
	req *connect.Request[pb.CategoryIdentifier],
) (*connect.Response[pb.ListVendorsResponse], error) {

	dbList, err := h.srvc.ListByCategory(ctx, req.Msg.Category)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoList := make([]*pb.Vendor, len(dbList))
	for i, v := range dbList {
		pv, err := mappers.DBToProtoVendor(&v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoList[i] = pv
	}

	return connect.NewResponse(&pb.ListVendorsResponse{
		Vendors:    protoList,
		TotalCount: int32(len(dbList)),
	}), nil
}
