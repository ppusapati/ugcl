package mappers

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/organization/api/v1/organization"
	db "p9e.in/ugcl/organization/db/generated"
)

// DepartmentDBToProto converts database Department model to protobuf Department
func DepartmentDBToProto(dbDept *db.Department) *pb.Department {
	if dbDept == nil {
		return nil
	}

	dept := &pb.Department{
		Id:           dbDept.ID.String(),
		TenantId:     dbDept.TenantID.String(),
		Code:         dbDept.Code,
		Name:         dbDept.Name,
		IsActive:     boolPtrToBool(dbDept.IsActive),
		DisplayOrder: int32PtrToInt32(dbDept.DisplayOrder),
		CreatedAt:    timestamppb.New(dbDept.CreatedAt),
		UpdatedAt:    timestamppb.New(dbDept.UpdatedAt),
	}

	if dbDept.DivisionID.Valid {
		dept.DivisionId = dbDept.DivisionID.UUID.String()
	}
	if dbDept.Description != nil {
		dept.Description = *dbDept.Description
	}
	if dbDept.DepartmentType != nil {
		dept.DepartmentType = *dbDept.DepartmentType
	}
	if dbDept.HeadUserID != nil {
		dept.HeadUserId = *dbDept.HeadUserID
	}
	if dbDept.ParentDepartmentID.Valid {
		dept.ParentDepartmentId = dbDept.ParentDepartmentID.UUID.String()
	}
	if dbDept.CreatedBy != nil {
		dept.CreatedBy = *dbDept.CreatedBy
	}
	if dbDept.UpdatedBy != nil {
		dept.UpdatedBy = *dbDept.UpdatedBy
	}
	if dbDept.Metadata != nil {
		metadata, _ := structpb.NewStruct(jsonRawMessageToMap(dbDept.Metadata))
		dept.Metadata = metadata
	}

	return dept
}

// DepartmentHierarchyDBToProto converts database DepartmentHierarchy to protobuf DepartmentHierarchy
func DepartmentHierarchyDBToProto(dbHier *db.DepartmentHierarchy) *pb.DepartmentHierarchy {
	if dbHier == nil {
		return nil
	}

	hier := &pb.DepartmentHierarchy{
		Id:           dbHier.ID.String(),
		TenantId:     dbHier.TenantID.String(),
		Code:         dbHier.Code,
		Name:         dbHier.Name,
		IsActive:     boolPtrToBool(dbHier.IsActive),
		DisplayOrder: int32PtrToInt32(dbHier.DisplayOrder),
		Level:        dbHier.Level,
		FullPath:     dbHier.FullPath,
	}

	if dbHier.DivisionID.Valid {
		hier.DivisionId = dbHier.DivisionID.UUID.String()
	}
	if dbHier.Description != nil {
		hier.Description = *dbHier.Description
	}
	if dbHier.DepartmentType != nil {
		hier.DepartmentType = *dbHier.DepartmentType
	}
	if dbHier.HeadUserID != nil {
		hier.HeadUserId = *dbHier.HeadUserID
	}
	if dbHier.ParentDepartmentID.Valid {
		hier.ParentDepartmentId = dbHier.ParentDepartmentID.UUID.String()
	}

	// Convert path from interface{} to []string
	if dbHier.Path != nil {
		if pathArray, ok := dbHier.Path.([]interface{}); ok {
			strPath := make([]string, len(pathArray))
			for i, p := range pathArray {
				if s, ok := p.(string); ok {
					strPath[i] = s
				}
			}
			hier.Path = strPath
		}
	}

	return hier
}

// CreateDepartmentProtoToDB converts protobuf CreateDepartmentRequest to database params
func CreateDepartmentProtoToDB(req *pb.CreateDepartmentRequest) (db.CreateDepartmentParams, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return db.CreateDepartmentParams{}, err
	}

	params := db.CreateDepartmentParams{
		TenantID:     tenantID,
		Code:         req.Code,
		Name:         req.Name,
		IsActive:     boolToBoolPtr(req.IsActive),
		DisplayOrder: int32ToInt32Ptr(req.DisplayOrder),
	}

	if req.DivisionId != "" {
		params.DivisionID = stringToUUIDNull(req.DivisionId)
	}
	if req.Description != "" {
		params.Description = &req.Description
	}
	if req.DepartmentType != "" {
		params.DepartmentType = &req.DepartmentType
	}
	if req.HeadUserId != "" {
		params.HeadUserID = &req.HeadUserId
	}
	if req.ParentDepartmentId != "" {
		params.ParentDepartmentID = stringToUUIDNull(req.ParentDepartmentId)
	}
	if req.CreatedBy != "" {
		params.CreatedBy = &req.CreatedBy
	}
	if req.Metadata != nil {
		params.Metadata = structToJSONRawMessage(req.Metadata)
	}

	return params, nil
}

// UpdateDepartmentProtoToDB converts protobuf UpdateDepartmentRequest to database params
func UpdateDepartmentProtoToDB(req *pb.UpdateDepartmentRequest) (db.UpdateDepartmentParams, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return db.UpdateDepartmentParams{}, err
	}

	params := db.UpdateDepartmentParams{
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
	if req.DepartmentType != nil {
		deptType := req.DepartmentType.Value
		params.DepartmentType = &deptType
	}
	if req.HeadUserId != nil {
		headID := req.HeadUserId.Value
		params.HeadUserID = &headID
	}
	if req.ParentDepartmentId != nil {
		parentID := req.ParentDepartmentId.Value
		params.ParentDepartmentID = stringToUUIDNull(parentID)
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
