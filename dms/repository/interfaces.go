package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "p9e.in/ugcl/dms/db/generated"
)

// IDocumentRepository defines the interface for document data operations
type IDocumentRepository interface {
	Create(ctx context.Context, arg db.CreateDocumentParams) (*db.Document, error)
	GetByID(ctx context.Context, arg db.GetDocumentParams) (*db.Document, error)
	List(ctx context.Context, arg db.ListDocumentsParams) ([]db.Document, error)
	Count(ctx context.Context, arg db.CountDocumentsParams) (int64, error)
	Update(ctx context.Context, arg db.UpdateDocumentParams) (*db.Document, error)
	SoftDelete(ctx context.Context, arg db.SoftDeleteDocumentParams) error
	GetExpired(ctx context.Context, tenantID uuid.UUID) ([]db.Document, error)
	SearchByTags(ctx context.Context, arg db.SearchDocumentsByTagsParams) ([]db.Document, error)
}

// IDocumentShareRepository defines the interface for document share data operations
type IDocumentShareRepository interface {
	Create(ctx context.Context, arg db.CreateDocumentShareParams) (*db.DocumentShare, error)
	GetByID(ctx context.Context, arg db.GetDocumentShareParams) (*db.DocumentShare, error)
	List(ctx context.Context, arg db.ListDocumentSharesParams) ([]db.DocumentShare, error)
	GetByLink(ctx context.Context, shareLink pgtype.Text) (*db.DocumentShare, error)
	UpdateAccessCount(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, arg db.DeleteDocumentShareParams) error
	GetUserShares(ctx context.Context, arg db.GetUserDocumentSharesParams) ([]db.GetUserDocumentSharesRow, error)
}
