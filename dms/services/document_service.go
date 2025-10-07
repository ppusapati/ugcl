package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/google/uuid"

	db "p9e.in/ugcl/dms/db/generated"
	"p9e.in/ugcl/dms/mappers"
	"p9e.in/ugcl/dms/repository"
	pb "p9e.in/ugcl/dms/api/v1"
)

// DocumentService implements IDocumentService
type DocumentService struct {
	documentRepo repository.IDocumentRepository
	// TODO: Add storage service, processing service when wiring
}

// NewDocumentService creates a new document service
func NewDocumentService(documentRepo repository.IDocumentRepository) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
	}
}

func (s *DocumentService) UploadDocument(ctx context.Context, req *pb.UploadDocumentRequest) (*pb.UploadDocumentResponse, error) {
	// Validate request
	if req.TenantId == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.FileName == "" {
		return nil, fmt.Errorf("file_name is required")
	}
	if len(req.FileContent) == 0 {
		return nil, fmt.Errorf("file_content is required")
	}

	// TODO: Get current user from context
	currentUserID := uuid.New()

	// Calculate checksum
	hash := sha256.New()
	hash.Write(req.FileContent)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Check for duplicate (same checksum in tenant)
	// This would require a GetByChecksum query - skipping for now

	// Prepare storage path
	storagePath := fmt.Sprintf("%s/%s/%s", req.TenantId, uuid.New().String(), req.FileName)

	// TODO: Upload to MinIO storage service
	// For now, just create the database record

	// Build create params
	params := db.CreateDocumentParams{
		TenantID:         tenantID,
		DocumentType:     mappers.DocumentTypeProtoToDB(req.DocumentType),
		DocumentCategory: mappers.DocumentCategoryProtoToDB(req.DocumentCategory),
		FileName:         req.FileName,
		OriginalName:     req.FileName,
		MimeType:         detectMimeType(req.FileName, req.FileContent),
		SizeBytes:        int64(len(req.FileContent)),
		Checksum:         checksum,
		StoragePath:      storagePath,
		UploadedBy:       currentUserID,
		CreatedBy:        currentUserID,
		UpdatedBy:        currentUserID,
	}

	// Optional fields
	if req.OwnerEntityType != nil {
		entityType := req.OwnerEntityType.Value
		params.OwnerEntityType = &entityType
	}
	if req.OwnerEntityId != nil {
		entityID, err := uuid.Parse(req.OwnerEntityId.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid owner_entity_id: %w", err)
		}
		params.OwnerEntityID = &entityID
	}
	if req.DivisionId != nil {
		divID, err := uuid.Parse(req.DivisionId.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid division_id: %w", err)
		}
		params.DivisionID = &divID
	}
	if req.BranchId != nil {
		branchID, err := uuid.Parse(req.BranchId.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid branch_id: %w", err)
		}
		params.BranchID = &branchID
	}
	if req.DepartmentId != nil {
		deptID, err := uuid.Parse(req.DepartmentId.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid department_id: %w", err)
		}
		params.DepartmentID = &deptID
	}
	if req.Title != nil {
		title := req.Title.Value
		params.Title = &title
	}
	if req.Description != nil {
		desc := req.Description.Value
		params.Description = &desc
	}
	if req.ExpiresAt != nil {
		expiresAt := req.ExpiresAt.AsTime()
		params.ExpiresAt = &expiresAt
	}

	// Create document
	dbDoc, err := s.documentRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Convert to proto
	doc, err := mappers.DocumentDBToProto(dbDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to map document: %w", err)
	}

	return &pb.UploadDocumentResponse{
		Document: doc,
		UploadId: dbDoc.ID.String(),
		Message:  "Document uploaded successfully",
	}, nil
}

func (s *DocumentService) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.GetDocumentResponse, error) {
	docID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	dbDoc, err := s.documentRepo.GetByID(ctx, db.GetDocumentParams{
		ID:       docID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	doc, err := mappers.DocumentDBToProto(dbDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to map document: %w", err)
	}

	return &pb.GetDocumentResponse{
		Document: doc,
	}, nil
}

func (s *DocumentService) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	// Set defaults for pagination
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	pageNumber := req.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}

	offset := (pageNumber - 1) * pageSize

	// Build query params
	params := db.ListDocumentsParams{
		TenantID: tenantID,
		Limit:    pageSize,
		Offset:   offset,
	}

	// Optional filters
	if req.OwnerEntityType != nil {
		entityType := req.OwnerEntityType.Value
		params.OwnerEntityType = &entityType
	}
	if req.OwnerEntityId != nil {
		entityID, _ := uuid.Parse(req.OwnerEntityId.Value)
		params.OwnerEntityID = &entityID
	}
	if req.DivisionId != nil {
		divID, _ := uuid.Parse(req.DivisionId.Value)
		params.DivisionID = &divID
	}
	if req.BranchId != nil {
		branchID, _ := uuid.Parse(req.BranchId.Value)
		params.BranchID = &branchID
	}
	if req.DepartmentId != nil {
		deptID, _ := uuid.Parse(req.DepartmentId.Value)
		params.DepartmentID = &deptID
	}
	if req.DocumentType != nil {
		docType := mappers.DocumentTypeProtoToDB(pb.DocumentType(req.DocumentType.Value))
		docTypeStr := string(docType)
		params.DocumentType = &docTypeStr
	}
	if req.DocumentCategory != nil {
		docCat := mappers.DocumentCategoryProtoToDB(pb.DocumentCategory(req.DocumentCategory.Value))
		docCatStr := string(docCat)
		params.DocumentCategory = &docCatStr
	}

	// Get documents
	dbDocs, err := s.documentRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Get total count
	countParams := db.CountDocumentsParams{
		TenantID:         params.TenantID,
		OwnerEntityType:  params.OwnerEntityType,
		OwnerEntityID:    params.OwnerEntityID,
		DivisionID:       params.DivisionID,
		BranchID:         params.BranchID,
		DepartmentID:     params.DepartmentID,
		DocumentType:     params.DocumentType,
		DocumentCategory: params.DocumentCategory,
	}
	totalCount, err := s.documentRepo.Count(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// Convert to proto
	documents := make([]*pb.Document, 0, len(dbDocs))
	for i := range dbDocs {
		doc, err := mappers.DocumentDBToProto(&dbDocs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to map document: %w", err)
		}
		documents = append(documents, doc)
	}

	return &pb.ListDocumentsResponse{
		Documents:  documents,
		TotalCount: int32(totalCount),
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}, nil
}

func (s *DocumentService) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentRequest) (*pb.UpdateDocumentResponse, error) {
	docID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	// TODO: Get current user from context
	currentUserID := uuid.New()

	// Check if document exists
	existing, err := s.documentRepo.GetByID(ctx, db.GetDocumentParams{
		ID:       docID,
		TenantID: tenantID,
	})
	if err != nil || existing == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Build update params
	params := db.UpdateDocumentParams{
		UpdatedBy: currentUserID,
		ID:        docID,
		TenantID:  tenantID,
	}

	// Optional updates
	if req.Title != nil {
		title := req.Title.Value
		params.Title = &title
	}
	if req.Description != nil {
		desc := req.Description.Value
		params.Description = &desc
	}
	if req.DocumentType != nil {
		docType := mappers.DocumentTypeProtoToDB(pb.DocumentType(req.DocumentType.Value))
		params.DocumentType = &docType
	}
	if req.DocumentCategory != nil {
		docCat := mappers.DocumentCategoryProtoToDB(pb.DocumentCategory(req.DocumentCategory.Value))
		params.DocumentCategory = &docCat
	}
	if req.DivisionId != nil {
		divID, _ := uuid.Parse(req.DivisionId.Value)
		params.DivisionID = &divID
	}
	if req.BranchId != nil {
		branchID, _ := uuid.Parse(req.BranchId.Value)
		params.BranchID = &branchID
	}
	if req.DepartmentId != nil {
		deptID, _ := uuid.Parse(req.DepartmentId.Value)
		params.DepartmentID = &deptID
	}
	if req.ExpiresAt != nil {
		expiresAt := req.ExpiresAt.AsTime()
		params.ExpiresAt = &expiresAt
	}

	// Update document
	dbDoc, err := s.documentRepo.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// Convert to proto
	doc, err := mappers.DocumentDBToProto(dbDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to map document: %w", err)
	}

	return &pb.UpdateDocumentResponse{
		Document: doc,
	}, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	docID, err := uuid.Parse(req.Id)
	if err != nil {
		return &pb.DeleteDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid document ID: %v", err),
		}, nil
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return &pb.DeleteDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid tenant ID: %v", err),
		}, nil
	}

	// TODO: Get current user from context
	currentUserID := uuid.New()

	// Soft delete
	err = s.documentRepo.SoftDelete(ctx, db.SoftDeleteDocumentParams{
		DeletedBy: currentUserID,
		ID:        docID,
		TenantID:  tenantID,
	})
	if err != nil {
		return &pb.DeleteDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete document: %v", err),
		}, nil
	}

	return &pb.DeleteDocumentResponse{
		Success: true,
		Message: "Document deleted successfully",
	}, nil
}

// Stub implementations for remaining methods
func (s *DocumentService) GetDocumentVersions(ctx context.Context, req *pb.GetDocumentVersionsRequest) (*pb.GetDocumentVersionsResponse, error) {
	// TODO: Implement version history
	return &pb.GetDocumentVersionsResponse{
		Versions: []*pb.DocumentVersion{},
	}, nil
}

func (s *DocumentService) ProcessDocument(ctx context.Context, req *pb.ProcessDocumentRequest) (*pb.ProcessDocumentResponse, error) {
	// TODO: Integrate with processing service
	return &pb.ProcessDocumentResponse{
		JobId:   uuid.New().String(),
		Status:  "PENDING",
		Message: "Document processing queued",
	}, nil
}

func (s *DocumentService) GetProcessingStatus(ctx context.Context, req *pb.GetProcessingStatusRequest) (*pb.GetProcessingStatusResponse, error) {
	// TODO: Query processing job status
	return &pb.GetProcessingStatusResponse{}, nil
}

func (s *DocumentService) GenerateThumbnail(ctx context.Context, req *pb.GenerateThumbnailRequest) (*pb.GenerateThumbnailResponse, error) {
	// TODO: Integrate with thumbnail generator
	return &pb.GenerateThumbnailResponse{}, nil
}

func (s *DocumentService) ExtractText(ctx context.Context, req *pb.ExtractTextRequest) (*pb.ExtractTextResponse, error) {
	// TODO: Integrate with OCR service
	return &pb.ExtractTextResponse{}, nil
}

func (s *DocumentService) GetDocumentPreview(ctx context.Context, req *pb.GetDocumentPreviewRequest) (*pb.GetDocumentPreviewResponse, error) {
	// TODO: Generate preview
	return &pb.GetDocumentPreviewResponse{}, nil
}

func (s *DocumentService) GetDocumentPage(ctx context.Context, req *pb.GetDocumentPageRequest) (*pb.GetDocumentPageResponse, error) {
	// TODO: Extract page from PDF
	return &pb.GetDocumentPageResponse{}, nil
}

func (s *DocumentService) StreamDocument(ctx context.Context, req *pb.StreamDocumentRequest, stream pb.DMSService_StreamDocumentServer) error {
	// TODO: Implement streaming from storage
	return fmt.Errorf("not implemented")
}

// Helper functions

func detectMimeType(filename string, content []byte) string {
	// Simple mime type detection based on extension
	// TODO: Use proper mime type detection library
	ext := filename[len(filename)-4:]
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", "jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".doc", "docx":
		return "application/msword"
	case ".xls", "xlsx":
		return "application/vnd.ms-excel"
	default:
		return "application/octet-stream"
	}
}
