package mappers

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "p9e.in/ugcl/identity/entity/api/v1"
	db "p9e.in/ugcl/identity/entity/db/generated"
)

// EntityDBToProto converts database Entity to proto Entity
func EntityDBToProto(dbEntity *db.Entity) (*pb.Entity, error) {
	if dbEntity == nil {
		return nil, nil
	}

	metadata, err := jsonRawMessageToMap(dbEntity.Metadata)
	if err != nil {
		return nil, err
	}

	entity := &pb.Entity{
		Id:              dbEntity.ID.String(),
		TenantId:        dbEntity.TenantID.String(),
		EntityType:      entityTypeDBToProto(dbEntity.EntityType),
		UserId:          dbEntity.UserID.String(),
		ReferenceId:     dbEntity.ReferenceID.String(),
		ReferenceSource: referenceSourceDBToProto(dbEntity.ReferenceSource),
		Status:          entityStatusDBToProto(dbEntity.Status),
		Metadata:        metadata,
		CreatedAt:       timeToTimestamp(dbEntity.CreatedAt),
		UpdatedAt:       timeToTimestamp(dbEntity.UpdatedAt),
		CreatedBy:       dbEntity.CreatedBy.String(),
		UpdatedBy:       dbEntity.UpdatedBy.String(),
	}

	if dbEntity.DivisionID != nil {
		entity.DivisionId = wrapperspb.String(dbEntity.DivisionID.String())
	}
	if dbEntity.BranchID != nil {
		entity.BranchId = wrapperspb.String(dbEntity.BranchID.String())
	}
	if dbEntity.DepartmentID != nil {
		entity.DepartmentId = wrapperspb.String(dbEntity.DepartmentID.String())
	}

	return entity, nil
}

// CreateEntityRequestToDBParams converts proto CreateEntityRequest to SQLC params
func CreateEntityRequestToDBParams(req *pb.CreateEntityRequest, createdBy uuid.UUID) (db.CreateEntityParams, error) {
	tenantID, err := stringToUUID(req.TenantId)
	if err != nil {
		return db.CreateEntityParams{}, err
	}

	userID, err := stringToUUID(req.UserId)
	if err != nil {
		return db.CreateEntityParams{}, err
	}

	referenceID, err := stringToUUID(req.ReferenceId)
	if err != nil {
		return db.CreateEntityParams{}, err
	}

	metadata, err := mapToJSONRawMessage(req.Metadata)
	if err != nil {
		return db.CreateEntityParams{}, err
	}

	params := db.CreateEntityParams{
		TenantID:        tenantID,
		EntityType:      entityTypeProtoToDB(req.EntityType),
		UserID:          userID,
		ReferenceID:     referenceID,
		ReferenceSource: referenceSourceProtoToDB(req.ReferenceSource),
		Status:          entityStatusProtoToDB(req.Status),
		Metadata:        metadata,
		CreatedBy:       createdBy,
		UpdatedBy:       createdBy,
	}

	if req.DivisionId != nil {
		divID, err := stringToUUID(req.DivisionId.Value)
		if err != nil {
			return db.CreateEntityParams{}, err
		}
		params.DivisionID = &divID
	}

	if req.BranchId != nil {
		branchID, err := stringToUUID(req.BranchId.Value)
		if err != nil {
			return db.CreateEntityParams{}, err
		}
		params.BranchID = &branchID
	}

	if req.DepartmentId != nil {
		deptID, err := stringToUUID(req.DepartmentId.Value)
		if err != nil {
			return db.CreateEntityParams{}, err
		}
		params.DepartmentID = &deptID
	}

	return params, nil
}

// UpdateEntityRequestToDBParams converts proto UpdateEntityRequest to SQLC params
func UpdateEntityRequestToDBParams(req *pb.UpdateEntityRequest, updatedBy uuid.UUID) (db.UpdateEntityParams, error) {
	id, err := stringToUUID(req.Id)
	if err != nil {
		return db.UpdateEntityParams{}, err
	}

	// Extract tenant_id from somewhere - you may need to pass it separately
	// For now, assuming it's validated elsewhere
	params := db.UpdateEntityParams{
		UpdatedBy: updatedBy,
		ID:        id,
		// TenantID will be set by caller
	}

	if req.Status != pb.EntityStatus_ENTITY_STATUS_UNSPECIFIED {
		status := entityStatusProtoToDB(req.Status)
		params.Status = db.NullEntityStatus{EntityStatus: status, Valid: true}
	}

	if req.DivisionId != nil {
		divID, err := stringToUUID(req.DivisionId.Value)
		if err != nil {
			return db.UpdateEntityParams{}, err
		}
		params.DivisionID = &divID
	}

	if req.BranchId != nil {
		branchID, err := stringToUUID(req.BranchId.Value)
		if err != nil {
			return db.UpdateEntityParams{}, err
		}
		params.BranchID = &branchID
	}

	if req.DepartmentId != nil {
		deptID, err := stringToUUID(req.DepartmentId.Value)
		if err != nil {
			return db.UpdateEntityParams{}, err
		}
		params.DepartmentID = &deptID
	}

	if req.Metadata != nil {
		metadata, err := mapToJSONRawMessage(req.Metadata)
		if err != nil {
			return db.UpdateEntityParams{}, err
		}
		params.Metadata = metadata
	}

	return params, nil
}

// Type conversion helpers

func entityTypeDBToProto(et db.EntityType) pb.EntityType {
	switch et {
	case db.EntityTypeEMPLOYEE:
		return pb.EntityType_ENTITY_TYPE_EMPLOYEE
	case db.EntityTypeCONTRACTOR:
		return pb.EntityType_ENTITY_TYPE_CONTRACTOR
	case db.EntityTypeVENDOR:
		return pb.EntityType_ENTITY_TYPE_VENDOR
	case db.EntityTypeCLIENT:
		return pb.EntityType_ENTITY_TYPE_CLIENT
	case db.EntityTypeDEVICE:
		return pb.EntityType_ENTITY_TYPE_DEVICE
	case db.EntityTypeDRONE:
		return pb.EntityType_ENTITY_TYPE_DRONE
	case db.EntityTypeBOT:
		return pb.EntityType_ENTITY_TYPE_BOT
	case db.EntityTypeSYSTEM:
		return pb.EntityType_ENTITY_TYPE_SYSTEM
	case db.EntityTypeADMIN:
		return pb.EntityType_ENTITY_TYPE_ADMIN
	case db.EntityTypeAGENT:
		return pb.EntityType_ENTITY_TYPE_AGENT
	default:
		return pb.EntityType_ENTITY_TYPE_UNSPECIFIED
	}
}

func entityTypeProtoToDB(et pb.EntityType) db.EntityType {
	switch et {
	case pb.EntityType_ENTITY_TYPE_EMPLOYEE:
		return db.EntityTypeEMPLOYEE
	case pb.EntityType_ENTITY_TYPE_CONTRACTOR:
		return db.EntityTypeCONTRACTOR
	case pb.EntityType_ENTITY_TYPE_VENDOR:
		return db.EntityTypeVENDOR
	case pb.EntityType_ENTITY_TYPE_CLIENT:
		return db.EntityTypeCLIENT
	case pb.EntityType_ENTITY_TYPE_DEVICE:
		return db.EntityTypeDEVICE
	case pb.EntityType_ENTITY_TYPE_DRONE:
		return db.EntityTypeDRONE
	case pb.EntityType_ENTITY_TYPE_BOT:
		return db.EntityTypeBOT
	case pb.EntityType_ENTITY_TYPE_SYSTEM:
		return db.EntityTypeSYSTEM
	case pb.EntityType_ENTITY_TYPE_ADMIN:
		return db.EntityTypeADMIN
	case pb.EntityType_ENTITY_TYPE_AGENT:
		return db.EntityTypeAGENT
	default:
		return db.EntityTypeEMPLOYEE // default fallback
	}
}

func entityStatusDBToProto(es db.EntityStatus) pb.EntityStatus {
	switch es {
	case db.EntityStatusACTIVE:
		return pb.EntityStatus_ENTITY_STATUS_ACTIVE
	case db.EntityStatusINACTIVE:
		return pb.EntityStatus_ENTITY_STATUS_INACTIVE
	case db.EntityStatusSUSPENDED:
		return pb.EntityStatus_ENTITY_STATUS_SUSPENDED
	case db.EntityStatusARCHIVED:
		return pb.EntityStatus_ENTITY_STATUS_ARCHIVED
	default:
		return pb.EntityStatus_ENTITY_STATUS_UNSPECIFIED
	}
}

func entityStatusProtoToDB(es pb.EntityStatus) db.EntityStatus {
	switch es {
	case pb.EntityStatus_ENTITY_STATUS_ACTIVE:
		return db.EntityStatusACTIVE
	case pb.EntityStatus_ENTITY_STATUS_INACTIVE:
		return db.EntityStatusINACTIVE
	case pb.EntityStatus_ENTITY_STATUS_SUSPENDED:
		return db.EntityStatusSUSPENDED
	case pb.EntityStatus_ENTITY_STATUS_ARCHIVED:
		return db.EntityStatusARCHIVED
	default:
		return db.EntityStatusACTIVE // default fallback
	}
}

func referenceSourceDBToProto(rs db.ReferenceSource) pb.ReferenceSource {
	switch rs {
	case db.ReferenceSourceEMPLOYEE:
		return pb.ReferenceSource_REFERENCE_SOURCE_EMPLOYEE
	case db.ReferenceSourceCONTRACTOR:
		return pb.ReferenceSource_REFERENCE_SOURCE_CONTRACTOR
	case db.ReferenceSourceVENDOR:
		return pb.ReferenceSource_REFERENCE_SOURCE_VENDOR
	case db.ReferenceSourceCLIENT:
		return pb.ReferenceSource_REFERENCE_SOURCE_CLIENT
	case db.ReferenceSourceDEVICE:
		return pb.ReferenceSource_REFERENCE_SOURCE_DEVICE
	case db.ReferenceSourceDRONE:
		return pb.ReferenceSource_REFERENCE_SOURCE_DRONE
	case db.ReferenceSourceBOT:
		return pb.ReferenceSource_REFERENCE_SOURCE_BOT
	default:
		return pb.ReferenceSource_REFERENCE_SOURCE_UNSPECIFIED
	}
}

func referenceSourceProtoToDB(rs pb.ReferenceSource) db.ReferenceSource {
	switch rs {
	case pb.ReferenceSource_REFERENCE_SOURCE_EMPLOYEE:
		return db.ReferenceSourceEMPLOYEE
	case pb.ReferenceSource_REFERENCE_SOURCE_CONTRACTOR:
		return db.ReferenceSourceCONTRACTOR
	case pb.ReferenceSource_REFERENCE_SOURCE_VENDOR:
		return db.ReferenceSourceVENDOR
	case pb.ReferenceSource_REFERENCE_SOURCE_CLIENT:
		return db.ReferenceSourceCLIENT
	case pb.ReferenceSource_REFERENCE_SOURCE_DEVICE:
		return db.ReferenceSourceDEVICE
	case pb.ReferenceSource_REFERENCE_SOURCE_DRONE:
		return db.ReferenceSourceDRONE
	case pb.ReferenceSource_REFERENCE_SOURCE_BOT:
		return db.ReferenceSourceBOT
	default:
		return db.ReferenceSourceEMPLOYEE // default fallback
	}
}

// Public exported helpers for services
func ReferenceSourceProtoToDB(rs pb.ReferenceSource) db.ReferenceSource {
	return referenceSourceProtoToDB(rs)
}

func EntityTypeProtoToDB(et pb.EntityType) db.EntityType {
	return entityTypeProtoToDB(et)
}

func EntityStatusProtoToDB(es pb.EntityStatus) db.EntityStatus {
	return entityStatusProtoToDB(es)
}
