package mappers

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/organization/api/v1/organization"
	db "p9e.in/ugcl/organization/db/generated"
)

// DivisionDBToProto converts database Division model to protobuf Division
func DivisionDBToProto(dbDiv *db.Division) *pb.Division {
	if dbDiv == nil {
		return nil
	}

	division := &pb.Division{
		Id:          dbDiv.ID.String(),
		TenantId:    dbDiv.TenantID.String(),
		Code:        dbDiv.Code,
		Name:        dbDiv.Name,
		IsActive:    boolPtrToBool(dbDiv.IsActive),
		DisplayOrder: int32PtrToInt32(dbDiv.DisplayOrder),
		CreatedAt:   timestamppb.New(dbDiv.CreatedAt),
		UpdatedAt:   timestamppb.New(dbDiv.UpdatedAt),
	}

	if dbDiv.Description != nil {
		division.Description = *dbDiv.Description
	}
	if dbDiv.HeadUserID != nil {
		division.HeadUserId = *dbDiv.HeadUserID
	}
	if dbDiv.CreatedBy != nil {
		division.CreatedBy = *dbDiv.CreatedBy
	}
	if dbDiv.UpdatedBy != nil {
		division.UpdatedBy = *dbDiv.UpdatedBy
	}
	if dbDiv.Metadata != nil {
		metadata, _ := structpb.NewStruct(jsonRawMessageToMap(dbDiv.Metadata))
		division.Metadata = metadata
	}

	return division
}

// DivisionSummaryDBToProto converts database DivisionSummary to protobuf DivisionSummary
func DivisionSummaryDBToProto(dbSummary *db.DivisionSummary) *pb.DivisionSummary {
	if dbSummary == nil {
		return nil
	}

	summary := &pb.DivisionSummary{
		Id:              dbSummary.ID.String(),
		TenantId:        dbSummary.TenantID.String(),
		Code:            dbSummary.Code,
		Name:            dbSummary.Name,
		IsActive:        boolPtrToBool(dbSummary.IsActive),
		DisplayOrder:    int32PtrToInt32(dbSummary.DisplayOrder),
		BranchCount:     dbSummary.BranchCount,
		DepartmentCount: dbSummary.DepartmentCount,
		CreatedAt:       timestamppb.New(dbSummary.CreatedAt),
		UpdatedAt:       timestamppb.New(dbSummary.UpdatedAt),
	}

	if dbSummary.Description != nil {
		summary.Description = *dbSummary.Description
	}
	if dbSummary.HeadUserID != nil {
		summary.HeadUserId = *dbSummary.HeadUserID
	}

	return summary
}

// CreateDivisionProtoToDB converts protobuf CreateDivisionRequest to database params
func CreateDivisionProtoToDB(req *pb.CreateDivisionRequest) (db.CreateDivisionParams, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return db.CreateDivisionParams{}, err
	}

	params := db.CreateDivisionParams{
		TenantID:     tenantID,
		Code:         req.Code,
		Name:         req.Name,
		IsActive:     boolToBoolPtr(req.IsActive),
		DisplayOrder: int32ToInt32Ptr(req.DisplayOrder),
	}

	if req.Description != "" {
		params.Description = &req.Description
	}
	if req.HeadUserId != "" {
		params.HeadUserID = &req.HeadUserId
	}
	if req.CreatedBy != "" {
		params.CreatedBy = &req.CreatedBy
	}
	if req.Metadata != nil {
		params.Metadata = structToJSONRawMessage(req.Metadata)
	}

	return params, nil
}

// UpdateDivisionProtoToDB converts protobuf UpdateDivisionRequest to database params
func UpdateDivisionProtoToDB(req *pb.UpdateDivisionRequest) (db.UpdateDivisionParams, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return db.UpdateDivisionParams{}, err
	}

	params := db.UpdateDivisionParams{
		ID: id,
	}

	if req.Code != nil {
		code := req.Code.Value
		params.Code = &code
	}
	if req.Name != nil {
		name := req.Name.Value
		params.Name = &name
	}
	if req.Description != nil {
		desc := req.Description.Value
		params.Description = &desc
	}
	if req.HeadUserId != nil {
		headID := req.HeadUserId.Value
		params.HeadUserID = &headID
	}
	if req.IsActive != nil {
		isActive := req.IsActive.Value
		params.IsActive = &isActive
	}
	if req.DisplayOrder != nil {
		order := req.DisplayOrder.Value
		params.DisplayOrder = &order
	}
	if req.UpdatedBy != "" {
		params.UpdatedBy = &req.UpdatedBy
	}
	if req.Metadata != nil {
		params.Metadata = structToJSONRawMessage(req.Metadata)
	}

	return params, nil
}
