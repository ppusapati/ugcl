package repository

import (
	"context"

	"github.com/google/uuid"
	db "p9e.in/ugcl/dms/db/generated"
)

// DocumentRepository implements IDocumentRepository
type DocumentRepository struct {
	queries *db.Queries
}

// NewDocumentRepository creates a new instance of DocumentRepository
func NewDocumentRepository(queries *db.Queries) IDocumentRepository {
	return &DocumentRepository{
		queries: queries,
	}
}

// Create creates a new document
func (r *DocumentRepository) Create(ctx context.Context, arg db.CreateDocumentParams) (*db.Document, error) {
	doc, err := r.queries.CreateDocument(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, arg db.GetDocumentParams) (*db.Document, error) {
	doc, err := r.queries.GetDocument(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// List retrieves a list of documents with filters
func (r *DocumentRepository) List(ctx context.Context, arg db.ListDocumentsParams) ([]db.Document, error) {
	return r.queries.ListDocuments(ctx, arg)
}

// Count counts documents with filters
func (r *DocumentRepository) Count(ctx context.Context, arg db.CountDocumentsParams) (int64, error) {
	return r.queries.CountDocuments(ctx, arg)
}

// Update updates a document
func (r *DocumentRepository) Update(ctx context.Context, arg db.UpdateDocumentParams) (*db.Document, error) {
	doc, err := r.queries.UpdateDocument(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// SoftDelete soft deletes a document
func (r *DocumentRepository) SoftDelete(ctx context.Context, arg db.SoftDeleteDocumentParams) error {
	return r.queries.SoftDeleteDocument(ctx, arg)
}

// GetExpired retrieves expired documents for a tenant
func (r *DocumentRepository) GetExpired(ctx context.Context, tenantID uuid.UUID) ([]db.Document, error) {
	return r.queries.GetExpiredDocuments(ctx, tenantID)
}

// SearchByTags searches documents by tags
func (r *DocumentRepository) SearchByTags(ctx context.Context, arg db.SearchDocumentsByTagsParams) ([]db.Document, error) {
	return r.queries.SearchDocumentsByTags(ctx, arg)
}
