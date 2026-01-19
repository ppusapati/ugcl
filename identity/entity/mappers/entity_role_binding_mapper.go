package mappers

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "p9e.in/ugcl/identity/entity/api/v1"
	db "p9e.in/ugcl/identity/entity/db/generated"
)

// EntityRoleBindingDBToProto converts database EntityRoleBinding to proto
func EntityRoleBindingDBToProto(dbBinding *db.EntityRoleBinding) *pb.EntityRoleBinding {
	if dbBinding == nil {
		return nil
	}

	binding := &pb.EntityRoleBinding{
		Id:        dbBinding.ID.String(),
		TenantId:  dbBinding.TenantID.String(),
		EntityId:  dbBinding.EntityID.String(),
		RoleId:    dbBinding.RoleID.String(),
		CreatedAt: timeToTimestamp(dbBinding.CreatedAt),
		UpdatedAt: timeToTimestamp(dbBinding.UpdatedAt),
		CreatedBy: dbBinding.CreatedBy.String(),
		UpdatedBy: dbBinding.UpdatedBy.String(),
	}

	if dbBinding.DivisionID != nil {
		binding.DivisionId = wrapperspb.String(dbBinding.DivisionID.String())
	}
	if dbBinding.BranchID != nil {
		binding.BranchId = wrapperspb.String(dbBinding.BranchID.String())
	}
	if dbBinding.DepartmentID != nil {
		binding.DepartmentId = wrapperspb.String(dbBinding.DepartmentID.String())
	}
	if dbBinding.ValidFrom != nil {
		binding.ValidFrom = timeToTimestamp(*dbBinding.ValidFrom)
	}
	if dbBinding.ValidUntil != nil {
		binding.ValidUntil = timeToTimestamp(*dbBinding.ValidUntil)
	}

	return binding
}

// CreateEntityRoleBindingRequestToDBParams converts proto request to SQLC params
func CreateEntityRoleBindingRequestToDBParams(req *pb.CreateEntityRoleBindingRequest, createdBy uuid.UUID) (db.CreateEntityRoleBindingParams, error) {
	tenantID, err := stringToUUID(req.TenantId)
	if err != nil {
		return db.CreateEntityRoleBindingParams{}, err
	}

	entityID, err := stringToUUID(req.EntityId)
	if err != nil {
		return db.CreateEntityRoleBindingParams{}, err
	}

	roleID, err := stringToUUID(req.RoleId)
	if err != nil {
		return db.CreateEntityRoleBindingParams{}, err
	}

	params := db.CreateEntityRoleBindingParams{
		TenantID:  tenantID,
		EntityID:  entityID,
		RoleID:    roleID,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	if req.DivisionId != nil {
		divID, err := stringToUUID(req.DivisionId.Value)
		if err != nil {
			return db.CreateEntityRoleBindingParams{}, err
		}
		params.DivisionID = &divID
	}

	if req.BranchId != nil {
		branchID, err := stringToUUID(req.BranchId.Value)
		if err != nil {
			return db.CreateEntityRoleBindingParams{}, err
		}
		params.BranchID = &branchID
	}

	if req.DepartmentId != nil {
		deptID, err := stringToUUID(req.DepartmentId.Value)
		if err != nil {
			return db.CreateEntityRoleBindingParams{}, err
		}
		params.DepartmentID = &deptID
	}

	if req.ValidFrom != nil {
		validFrom := req.ValidFrom.AsTime()
		params.ValidFrom = &validFrom
	}

	if req.ValidUntil != nil {
		validUntil := req.ValidUntil.AsTime()
		params.ValidUntil = &validUntil
	}

	return params, nil
}
