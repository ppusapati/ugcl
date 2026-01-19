package mappers

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"

	db "p9e.in/ugcl/dms/db/generated"
	pb "p9e.in/ugcl/dms/api/v1"
)

// DocumentShareDBToProto converts database DocumentShare to proto
func DocumentShareDBToProto(dbShare *db.DocumentShare) *pb.DocumentShare {
	if dbShare == nil {
		return nil
	}

	share := &pb.DocumentShare{
		Id:              dbShare.ID.String(),
		TenantId:        dbShare.TenantID.String(),
		DocumentId:      dbShare.DocumentID.String(),
		CanView:         dbShare.CanView,
		CanDownload:     dbShare.CanDownload,
		CanEdit:         dbShare.CanEdit,
		CanDelete:       dbShare.CanDelete,
		IsExpired:       dbShare.IsExpired,
		AccessCount:     dbShare.AccessCount,
		CreatedAt:       TimeToTimestamp(dbShare.CreatedAt),
		UpdatedAt:       TimeToTimestamp(dbShare.UpdatedAt),
		CreatedBy:       dbShare.CreatedBy.String(),
		UpdatedBy:       dbShare.UpdatedBy.String(),
	}

	// Optional fields
	if dbShare.SharedWithEntityID != nil {
		share.SharedWithEntityId = wrapperspb.String(dbShare.SharedWithEntityID.String())
	}
	if dbShare.SharedWithUserID != nil {
		share.SharedWithUserId = wrapperspb.String(dbShare.SharedWithUserID.String())
	}
	if dbShare.SharedWithEmail != nil {
		share.SharedWithEmail = wrapperspb.String(*dbShare.SharedWithEmail)
	}
	if dbShare.DivisionID != nil {
		share.DivisionId = wrapperspb.String(dbShare.DivisionID.String())
	}
	if dbShare.BranchID != nil {
		share.BranchId = wrapperspb.String(dbShare.BranchID.String())
	}
	if dbShare.DepartmentID != nil {
		share.DepartmentId = wrapperspb.String(dbShare.DepartmentID.String())
	}
	if dbShare.ShareLink != nil {
		share.ShareLink = wrapperspb.String(*dbShare.ShareLink)
	}
	if dbShare.ExpiresAt != nil {
		share.ExpiresAt = TimeToTimestamp(*dbShare.ExpiresAt)
	}
	if dbShare.LastAccessedAt != nil {
		share.LastAccessedAt = TimeToTimestamp(*dbShare.LastAccessedAt)
	}

	return share
}

// CreateDocumentShareRequestToDBParams converts proto request to SQLC params
func CreateDocumentShareRequestToDBParams(req *pb.CreateDocumentShareRequest, createdBy uuid.UUID) (db.CreateDocumentShareParams, error) {
	tenantID, err := StringToUUID(req.TenantId)
	if err != nil {
		return db.CreateDocumentShareParams{}, err
	}

	documentID, err := StringToUUID(req.DocumentId)
	if err != nil {
		return db.CreateDocumentShareParams{}, err
	}

	params := db.CreateDocumentShareParams{
		TenantID:    tenantID,
		DocumentID:  documentID,
		CanView:     req.CanView,
		CanDownload: req.CanDownload,
		CanEdit:     req.CanEdit,
		CanDelete:   req.CanDelete,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	// Optional entity share
	if req.SharedWithEntityId != nil {
		entityID, err := StringToUUID(req.SharedWithEntityId.Value)
		if err != nil {
			return db.CreateDocumentShareParams{}, err
		}
		params.SharedWithEntityID = &entityID
	}

	// Optional user share
	if req.SharedWithUserId != nil {
		userID, err := StringToUUID(req.SharedWithUserId.Value)
		if err != nil {
			return db.CreateDocumentShareParams{}, err
		}
		params.SharedWithUserID = &userID
	}

	// Optional email share
	if req.SharedWithEmail != nil {
		email := req.SharedWithEmail.Value
		params.SharedWithEmail = &email
	}

	// Optional organizational scope
	if req.DivisionId != nil {
		divID, err := StringToUUID(req.DivisionId.Value)
		if err != nil {
			return db.CreateDocumentShareParams{}, err
		}
		params.DivisionID = &divID
	}

	if req.BranchId != nil {
		branchID, err := StringToUUID(req.BranchId.Value)
		if err != nil {
			return db.CreateDocumentShareParams{}, err
		}
		params.BranchID = &branchID
	}

	if req.DepartmentId != nil {
		deptID, err := StringToUUID(req.DepartmentId.Value)
		if err != nil {
			return db.CreateDocumentShareParams{}, err
		}
		params.DepartmentID = &deptID
	}

	// Optional expiration
	if req.ExpiresAt != nil {
		expiresAt := req.ExpiresAt.AsTime()
		params.ExpiresAt = &expiresAt
	}

	// Optional share link/password
	if req.SharePassword != nil {
		password := req.SharePassword.Value
		params.SharePassword = &password
	}

	return params, nil
}
