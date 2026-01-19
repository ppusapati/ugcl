package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	db "p9e.in/ugcl/dms/db/generated"
	"p9e.in/ugcl/dms/mappers"
	"p9e.in/ugcl/dms/repository"
	pb "p9e.in/ugcl/dms/api/v1"
)

// DocumentShareService implements IDocumentShareService
type DocumentShareService struct {
	shareRepo    repository.IDocumentShareRepository
	documentRepo repository.IDocumentRepository
}

// NewDocumentShareService creates a new document share service
func NewDocumentShareService(
	shareRepo repository.IDocumentShareRepository,
	documentRepo repository.IDocumentRepository,
) *DocumentShareService {
	return &DocumentShareService{
		shareRepo:    shareRepo,
		documentRepo: documentRepo,
	}
}

func (s *DocumentShareService) CreateDocumentShare(ctx context.Context, req *pb.CreateDocumentShareRequest) (*pb.CreateDocumentShareResponse, error) {
	// Validate request
	if req.TenantId == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DocumentId == "" {
		return nil, fmt.Errorf("document_id is required")
	}

	// Must have at least one share target
	if req.SharedWithEntityId == nil && req.SharedWithUserId == nil && req.SharedWithEmail == nil {
		return nil, fmt.Errorf("must specify at least one share target (entity_id, user_id, or email)")
	}

	// Verify document exists
	docID, err := uuid.Parse(req.DocumentId)
	if err != nil {
		return nil, fmt.Errorf("invalid document_id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	existing, err := s.documentRepo.GetByID(ctx, db.GetDocumentParams{
		ID:       docID,
		TenantID: tenantID,
	})
	if err != nil || existing == nil {
		return nil, fmt.Errorf("document not found")
	}

	// TODO: Get current user from context
	currentUserID := uuid.New()

	// Convert request to DB params
	params, err := mappers.CreateDocumentShareRequestToDBParams(req, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to map request: %w", err)
	}

	// Create share
	dbShare, err := s.shareRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	// Convert to proto
	share := mappers.DocumentShareDBToProto(dbShare)

	return &pb.CreateDocumentShareResponse{
		Share: share,
	}, nil
}

func (s *DocumentShareService) GetDocumentShare(ctx context.Context, req *pb.GetDocumentShareRequest) (*pb.GetDocumentShareResponse, error) {
	shareID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid share id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	dbShare, err := s.shareRepo.GetByID(ctx, db.GetDocumentShareParams{
		ID:       shareID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("share not found: %w", err)
	}

	share := mappers.DocumentShareDBToProto(dbShare)

	return &pb.GetDocumentShareResponse{
		Share: share,
	}, nil
}

func (s *DocumentShareService) ListDocumentShares(ctx context.Context, req *pb.ListDocumentSharesRequest) (*pb.ListDocumentSharesResponse, error) {
	docID, err := uuid.Parse(req.DocumentId)
	if err != nil {
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	dbShares, err := s.shareRepo.List(ctx, db.ListDocumentSharesParams{
		DocumentID: docID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list shares: %w", err)
	}

	shares := make([]*pb.DocumentShare, 0, len(dbShares))
	for i := range dbShares {
		share := mappers.DocumentShareDBToProto(&dbShares[i])
		shares = append(shares, share)
	}

	return &pb.ListDocumentSharesResponse{
		Shares: shares,
	}, nil
}

func (s *DocumentShareService) GetUserShares(ctx context.Context, req *pb.GetUserSharesRequest) (*pb.GetUserSharesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	var userID uuid.UUID
	if req.UserId != nil {
		userID, err = uuid.Parse(req.UserId.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id: %w", err)
		}
	}

	var entityID uuid.UUID
	if req.EntityId != nil {
		entityID, err = uuid.Parse(req.EntityId.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid entity_id: %w", err)
		}
	}

	// Query shares for this user/entity
	dbShares, err := s.shareRepo.GetUserShares(ctx, db.GetUserDocumentSharesParams{
		TenantID:            tenantID,
		SharedWithUserID:    &userID,
		SharedWithEntityID:  &entityID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user shares: %w", err)
	}

	// Convert to proto
	shares := make([]*pb.DocumentShare, 0, len(dbShares))
	for i := range dbShares {
		// Map the joined result to DocumentShare
		share := &pb.DocumentShare{
			Id:         dbShares[i].ID.String(),
			TenantId:   dbShares[i].TenantID.String(),
			DocumentId: dbShares[i].DocumentID.String(),
			// Add more fields as needed
		}
		shares = append(shares, share)
	}

	return &pb.GetUserSharesResponse{
		Shares: shares,
	}, nil
}

func (s *DocumentShareService) DeleteDocumentShare(ctx context.Context, req *pb.DeleteDocumentShareRequest) (*pb.DeleteDocumentShareResponse, error) {
	shareID, err := uuid.Parse(req.Id)
	if err != nil {
		return &pb.DeleteDocumentShareResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid share ID: %v", err),
		}, nil
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return &pb.DeleteDocumentShareResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid tenant ID: %v", err),
		}, nil
	}

	err = s.shareRepo.Delete(ctx, db.DeleteDocumentShareParams{
		ID:       shareID,
		TenantID: tenantID,
	})
	if err != nil {
		return &pb.DeleteDocumentShareResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete share: %v", err),
		}, nil
	}

	return &pb.DeleteDocumentShareResponse{
		Success: true,
		Message: "Share deleted successfully",
	}, nil
}
