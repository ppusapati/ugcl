package indexers

import (
	"context"
	"fmt"
	"time"

	"p9e.in/ugcl/metasearch/models"
	"p9e.in/ugcl/metasearch/services/interfaces"
)

// IIndexer defines the interface for all indexers
type IIndexer interface {
	GetIndexName() string
	GetIndexConfig() *models.IndexConfig
	IndexDocument(ctx context.Context, document interface{}) error
	UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteDocument(ctx context.Context, id string) error
	BulkIndex(ctx context.Context, documents []interface{}) error
	PrepareDocument(document interface{}) (map[string]interface{}, error)
}

// BaseIndexer provides common indexing functionality
type BaseIndexer struct {
	indexingService interfaces.IIndexingService
	indexName       string
	indexConfig     *models.IndexConfig
}

// NewBaseIndexer creates a new base indexer
func NewBaseIndexer(indexingService interfaces.IIndexingService, indexName string) *BaseIndexer {
	return &BaseIndexer{
		indexingService: indexingService,
		indexName:       indexName,
		indexConfig:     getDefaultIndexConfig(indexName),
	}
}

// GetIndexName returns the index name
func (b *BaseIndexer) GetIndexName() string {
	return b.indexName
}

// GetIndexConfig returns the index configuration
func (b *BaseIndexer) GetIndexConfig() *models.IndexConfig {
	return b.indexConfig
}

// IndexDocument indexes a single document
func (b *BaseIndexer) IndexDocument(ctx context.Context, document interface{}) error {
	doc, err := b.PrepareDocument(document)
	if err != nil {
		return fmt.Errorf("failed to prepare document: %w", err)
	}

	docID, ok := doc["id"].(string)
	if !ok {
		return fmt.Errorf("document must have an 'id' field")
	}

	return b.indexingService.IndexDocument(ctx, b.indexName, docID, doc)
}

// UpdateDocument updates a document
func (b *BaseIndexer) UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error {
	return b.indexingService.UpdateDocument(ctx, b.indexName, id, updates)
}

// DeleteDocument deletes a document
func (b *BaseIndexer) DeleteDocument(ctx context.Context, id string) error {
	return b.indexingService.DeleteDocument(ctx, b.indexName, id)
}

// BulkIndex indexes multiple documents
func (b *BaseIndexer) BulkIndex(ctx context.Context, documents []interface{}) error {
	var operations []interfaces.BulkOperation

	for _, doc := range documents {
		prepared, err := b.PrepareDocument(doc)
		if err != nil {
			continue // Skip invalid documents
		}

		docID, ok := prepared["id"].(string)
		if !ok {
			continue // Skip documents without ID
		}

		operations = append(operations, interfaces.BulkOperation{
			Action:   "index",
			Index:    b.indexName,
			ID:       docID,
			Document: prepared,
		})
	}

	return b.indexingService.BulkIndex(ctx, operations)
}

// PrepareDocument prepares a document for indexing (to be overridden by specific indexers)
func (b *BaseIndexer) PrepareDocument(document interface{}) (map[string]interface{}, error) {
	// Default implementation - just convert to map
	doc, ok := document.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("document must be a map[string]interface{}")
	}

	// Add common fields
	doc["indexed_at"] = time.Now()
	doc["index_type"] = b.indexName

	return doc, nil
}

// getDefaultIndexConfig returns default index configuration
func getDefaultIndexConfig(indexName string) *models.IndexConfig {
	return &models.IndexConfig{
		Name: indexName,
		Settings: models.IndexSettings{
			NumberOfShards:   1,
			NumberOfReplicas: 1,
			RefreshInterval:  "1s",
			MaxResultWindow:  10000,
			Analysis: models.AnalysisSettings{
				Analyzers: map[string]models.Analyzer{
					"standard_analyzer": {
						Type:      "standard",
						Tokenizer: "standard",
						Filters:   []string{"lowercase", "stop"},
					},
					"text_analyzer": {
						Type:      "custom",
						Tokenizer: "standard",
						Filters:   []string{"lowercase", "stop", "stemmer"},
					},
				},
			},
		},
		Mappings: models.IndexMappings{
			Properties: map[string]models.FieldMapping{
				"id": {
					Type:  "keyword",
					Index: true,
				},
				"title": {
					Type:     "text",
					Analyzer: "text_analyzer",
					Index:    true,
				},
				"content": {
					Type:     "text",
					Analyzer: "text_analyzer",
					Index:    true,
				},
				"created_at": {
					Type:  "date",
					Index: true,
				},
				"updated_at": {
					Type:  "date",
					Index: true,
				},
				"indexed_at": {
					Type:  "date",
					Index: true,
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
		IsActive:  true,
	}
}
