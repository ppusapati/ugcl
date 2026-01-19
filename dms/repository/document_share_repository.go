package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "p9e.in/ugcl/dms/db/generated"
)

// DocumentShareRepository implements IDocumentShareRepository
type DocumentShareRepository struct {
	queries *db.Queries
}

// NewDocumentShareRepository creates a new instance of DocumentShareRepository
func NewDocumentShareRepository(queries *db.Queries) IDocumentShareRepository {
	return &DocumentShareRepository{
		queries: queries,
	}
}

// Create creates a new document share
func (r *DocumentShareRepository) Create(ctx context.Context, arg db.CreateDocumentShareParams) (*db.DocumentShare, error) {
	share, err := r.queries.CreateDocumentShare(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetByID retrieves a document share by ID
func (r *DocumentShareRepository) GetByID(ctx context.Context, arg db.GetDocumentShareParams) (*db.DocumentShare, error) {
	share, err := r.queries.GetDocumentShare(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// List retrieves all shares for a document
func (r *DocumentShareRepository) List(ctx context.Context, arg db.ListDocumentSharesParams) ([]db.DocumentShare, error) {
	return r.queries.ListDocumentShares(ctx, arg)
}

// GetByLink retrieves a document share by share link
func (r *DocumentShareRepository) GetByLink(ctx context.Context, shareLink pgtype.Text) (*db.DocumentShare, error) {
	share, err := r.queries.GetShareByLink(ctx, shareLink)
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// UpdateAccessCount increments the access count for a share
func (r *DocumentShareRepository) UpdateAccessCount(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateShareAccessCount(ctx, id)
}

// Delete deletes a document share
func (r *DocumentShareRepository) Delete(ctx context.Context, arg db.DeleteDocumentShareParams) error {
	return r.queries.DeleteDocumentShare(ctx, arg)
}

// GetUserShares retrieves all shares for a user
func (r *DocumentShareRepository) GetUserShares(ctx context.Context, arg db.GetUserDocumentSharesParams) ([]db.GetUserDocumentSharesRow, error) {
	return r.queries.GetUserDocumentShares(ctx, arg)
}
