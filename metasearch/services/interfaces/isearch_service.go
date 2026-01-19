package interfaces

import (
	"context"

	"p9e.in/ugcl/metasearch/models"
)

// IQueryService handles search queries
type IQueryService interface {
	Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error)
	MultiSearch(ctx context.Context, requests []*models.SearchRequest) ([]*models.SearchResponse, error)
	ScrollSearch(ctx context.Context, req *models.SearchRequest, scrollID string) (*models.SearchResponse, error)
	Count(ctx context.Context, req *models.SearchRequest) (int64, error)
}

// ISuggestionService handles auto-complete and suggestions
type ISuggestionService interface {
	GetSuggestions(ctx context.Context, req *models.SuggestionRequest) (*models.SearchResponse, error)
	GetPopularQueries(ctx context.Context, limit int) ([]models.QueryFrequency, error)
	GetQueryHistory(ctx context.Context, userID string, limit int) ([]string, error)
}

// IIndexingService handles document indexing
type IIndexingService interface {
	IndexDocument(ctx context.Context, indexName string, docID string, document map[string]interface{}) error
	IndexDocuments(ctx context.Context, indexName string, documents []map[string]interface{}) error
	UpdateDocument(ctx context.Context, indexName string, docID string, updates map[string]interface{}) error
	DeleteDocument(ctx context.Context, indexName string, docID string) error
	BulkIndex(ctx context.Context, operations []BulkOperation) error
	RefreshIndex(ctx context.Context, indexName string) error
}

// IAdminService handles index administration
type IAdminService interface {
	CreateIndex(ctx context.Context, config *models.IndexConfig) error
	DeleteIndex(ctx context.Context, indexName string) error
	GetIndexStats(ctx context.Context, indexName string) (*models.IndexStats, error)
	GetAllIndices(ctx context.Context) ([]*models.IndexStats, error)
	ReindexData(ctx context.Context, sourceIndex, targetIndex string) (*models.IndexingJob, error)
	GetIndexingJobs(ctx context.Context, status string) ([]*models.IndexingJob, error)
	UpdateIndexSettings(ctx context.Context, indexName string, settings map[string]interface{}) error
}

// IAnalyticsService handles search analytics
type IAnalyticsService interface {
	RecordSearch(ctx context.Context, query string, indexName string, resultCount int64, responseTime int64) error
	GetSearchAnalytics(ctx context.Context, timeRange *models.TimeRange) (*models.SearchAnalytics, error)
	GetPopularSearches(ctx context.Context, limit int) ([]models.QueryFrequency, error)
	GetNoResultQueries(ctx context.Context, limit int) ([]string, error)
	GetSearchTrends(ctx context.Context, timeRange *models.TimeRange) (map[string]int64, error)
}

// BulkOperation represents a bulk indexing operation
type BulkOperation struct {
	Action   string                 `json:"action"` // "index", "update", "delete"
	Index    string                 `json:"index"`
	ID       string                 `json:"id"`
	Document map[string]interface{} `json:"document,omitempty"`
	Updates  map[string]interface{} `json:"updates,omitempty"`
}
