package mappers

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/organization/api/v1/organization"
	db "p9e.in/ugcl/organization/db/generated"
)

// BranchDBToProto converts database Branch model to protobuf Branch
func BranchDBToProto(dbBranch *db.Branch) *pb.Branch {
	if dbBranch == nil {
		return nil
	}

	branch := &pb.Branch{
		Id:           dbBranch.ID.String(),
		TenantId:     dbBranch.TenantID.String(),
		Code:         dbBranch.Code,
		Name:         dbBranch.Name,
		IsActive:     boolPtrToBool(dbBranch.IsActive),
		DisplayOrder: int32PtrToInt32(dbBranch.DisplayOrder),
		CreatedAt:    timestamppb.New(dbBranch.CreatedAt),
		UpdatedAt:    timestamppb.New(dbBranch.UpdatedAt),
	}

	if dbBranch.DivisionID.Valid {
		branch.DivisionId = dbBranch.DivisionID.UUID.String()
	}
	if dbBranch.BranchType != nil {
		branch.BranchType = *dbBranch.BranchType
	}

	// Address
	if dbBranch.AddressLine1 != nil || dbBranch.AddressLine2 != nil || dbBranch.City != nil {
		branch.Address = &pb.Address{
			Line1:      stringPtrToString(dbBranch.AddressLine1),
			Line2:      stringPtrToString(dbBranch.AddressLine2),
			City:       stringPtrToString(dbBranch.City),
			State:      stringPtrToString(dbBranch.State),
			Country:    stringPtrToString(dbBranch.Country),
			PostalCode: stringPtrToString(dbBranch.PostalCode),
		}
	}

	// Location
	if dbBranch.Latitude.Valid || dbBranch.Longitude.Valid {
		branch.Location = &pb.Location{
			Latitude:  sqlNullFloat64ToFloat64(dbBranch.Latitude),
			Longitude: sqlNullFloat64ToFloat64(dbBranch.Longitude),
		}
	}

	// Contact
	if dbBranch.Phone != nil || dbBranch.Email != nil {
		branch.Contact = &pb.ContactInfo{
			Phone: stringPtrToString(dbBranch.Phone),
			Email: stringPtrToString(dbBranch.Email),
		}
	}

	if dbBranch.BranchManagerUserID != nil {
		branch.BranchManagerUserId = *dbBranch.BranchManagerUserID
	}
	if dbBranch.CreatedBy != nil {
		branch.CreatedBy = *dbBranch.CreatedBy
	}
	if dbBranch.UpdatedBy != nil {
		branch.UpdatedBy = *dbBranch.UpdatedBy
	}
	if dbBranch.Metadata != nil {
		metadata, _ := structpb.NewStruct(jsonRawMessageToMap(dbBranch.Metadata))
		branch.Metadata = metadata
	}

	return branch
}

// BranchSummaryDBToProto converts database BranchSummary to protobuf BranchSummary
func BranchSummaryDBToProto(dbSummary *db.BranchSummary) *pb.BranchSummary {
	if dbSummary == nil {
		return nil
	}

	summary := &pb.BranchSummary{
		Id:           dbSummary.ID.String(),
		TenantId:     dbSummary.TenantID.String(),
		Code:         dbSummary.Code,
		Name:         dbSummary.Name,
		IsActive:     boolPtrToBool(dbSummary.IsActive),
		DisplayOrder: int32PtrToInt32(dbSummary.DisplayOrder),
		CreatedAt:    timestamppb.New(dbSummary.CreatedAt),
		UpdatedAt:    timestamppb.New(dbSummary.UpdatedAt),
	}

	if dbSummary.DivisionID.Valid {
		summary.DivisionId = dbSummary.DivisionID.UUID.String()
	}
	summary.DivisionName = dbSummary.DivisionName
	summary.DivisionCode = dbSummary.DivisionCode

	if dbSummary.BranchType != nil {
		summary.BranchType = *dbSummary.BranchType
	}
	if dbSummary.City != nil {
		summary.City = *dbSummary.City
	}
	if dbSummary.State != nil {
		summary.State = *dbSummary.State
	}
	if dbSummary.Country != nil {
		summary.Country = *dbSummary.Country
	}
	if dbSummary.BranchManagerUserID != nil {
		summary.BranchManagerUserId = *dbSummary.BranchManagerUserID
	}

	return summary
}

// CreateBranchProtoToDB converts protobuf CreateBranchRequest to database params
func CreateBranchProtoToDB(req *pb.CreateBranchRequest) (db.CreateBranchParams, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return db.CreateBranchParams{}, err
	}

	params := db.CreateBranchParams{
		TenantID:     tenantID,
		Code:         req.Code,
		Name:         req.Name,
		IsActive:     boolToBoolPtr(req.IsActive),
		DisplayOrder: int32ToInt32Ptr(req.DisplayOrder),
	}

	if req.DivisionId != "" {
		params.DivisionID = stringToUUIDNull(req.DivisionId)
	}
	if req.BranchType != "" {
		params.BranchType = &req.BranchType
	}

	// Address
	if req.Address != nil {
		if req.Address.Line1 != "" {
			params.AddressLine1 = &req.Address.Line1
		}
		if req.Address.Line2 != "" {
			params.AddressLine2 = &req.Address.Line2
		}
		if req.Address.City != "" {
			params.City = &req.Address.City
		}
		if req.Address.State != "" {
			params.State = &req.Address.State
		}
		if req.Address.Country != "" {
			params.Country = &req.Address.Country
		}
		if req.Address.PostalCode != "" {
			params.PostalCode = &req.Address.PostalCode
		}
	}

	// Location
	if req.Location != nil {
		params.Latitude = float64ToSqlNullFloat64(req.Location.Latitude)
		params.Longitude = float64ToSqlNullFloat64(req.Location.Longitude)
	}

	// Contact
	if req.Contact != nil {
		if req.Contact.Phone != "" {
			params.Phone = &req.Contact.Phone
		}
		if req.Contact.Email != "" {
			params.Email = &req.Contact.Email
		}
	}

	if req.BranchManagerUserId != "" {
		params.BranchManagerUserID = &req.BranchManagerUserId
	}
	if req.CreatedBy != "" {
		params.CreatedBy = &req.CreatedBy
	}
	if req.Metadata != nil {
		params.Metadata = structToJSONRawMessage(req.Metadata)
	}

	return params, nil
}

// UpdateBranchProtoToDB converts protobuf UpdateBranchRequest to database params
func UpdateBranchProtoToDB(req *pb.UpdateBranchRequest) (db.UpdateBranchParams, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return db.UpdateBranchParams{}, err
	}

	params := db.UpdateBranchParams{
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
	if req.BranchType != nil {
		branchType := req.BranchType.Value
		params.BranchType = &branchType
	}

	// Address
	if req.Address != nil {
		if req.Address.Line1 != "" {
			params.AddressLine1 = &req.Address.Line1
		}
		if req.Address.Line2 != "" {
			params.AddressLine2 = &req.Address.Line2
		}
		if req.Address.City != "" {
			params.City = &req.Address.City
		}
		if req.Address.State != "" {
			params.State = &req.Address.State
		}
		if req.Address.Country != "" {
			params.Country = &req.Address.Country
		}
		if req.Address.PostalCode != "" {
			params.PostalCode = &req.Address.PostalCode
		}
	}

	// Location
	if req.Location != nil {
		params.Latitude = float64ToSqlNullFloat64(req.Location.Latitude)
		params.Longitude = float64ToSqlNullFloat64(req.Location.Longitude)
	}

	// Contact
	if req.Contact != nil {
		if req.Contact.Phone != "" {
			params.Phone = &req.Contact.Phone
		}
		if req.Contact.Email != "" {
			params.Email = &req.Contact.Email
		}
	}

	if req.BranchManagerUserId != nil {
		managerID := req.BranchManagerUserId.Value
		params.BranchManagerUserID = &managerID
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
