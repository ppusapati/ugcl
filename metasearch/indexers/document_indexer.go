package indexers

import (
	"context"
	"fmt"
	"strings"

	"p9e.in/ugcl/metasearch/models"
	"p9e.in/ugcl/metasearch/services/interfaces"
)

// DocumentIndexer handles indexing of documents
type DocumentIndexer struct {
	*BaseIndexer
}

// DocumentData represents the structure of document data for indexing
type DocumentData struct {
	ID            string                 `json:"id"`
	FileName      string                 `json:"file_name"`
	Title         string                 `json:"title"`
	Content       string                 `json:"content"`
	ExtractedText string                 `json:"extracted_text"`
	MimeType      string                 `json:"mime_type"`
	SizeBytes     int64                  `json:"size_bytes"`
	PageCount     int                    `json:"page_count"`
	Author        string                 `json:"author"`
	UploadedBy    string                 `json:"uploaded_by"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	Tags          []string               `json:"tags"`
	Categories    []string               `json:"categories"`
	Permissions   []string               `json:"permissions"`
	Metadata      map[string]interface{} `json:"metadata"`
	Checksum      string                 `json:"checksum"`
	StoragePath   string                 `json:"storage_path"`
	IsDeleted     bool                   `json:"is_deleted"`
}

// NewDocumentIndexer creates a new document indexer
func NewDocumentIndexer(indexingService interfaces.IIndexingService) *DocumentIndexer {
	baseIndexer := NewBaseIndexer(indexingService, "documents")

	// Customize index config for documents
	baseIndexer.indexConfig = getDocumentIndexConfig()

	return &DocumentIndexer{
		BaseIndexer: baseIndexer,
	}
}

// PrepareDocument prepares a document for indexing
func (d *DocumentIndexer) PrepareDocument(document interface{}) (map[string]interface{}, error) {
	doc, ok := document.(*DocumentData)
	if !ok {
		return nil, fmt.Errorf("document must be of type *DocumentData")
	}

	prepared := map[string]interface{}{
		"id":             doc.ID,
		"file_name":      doc.FileName,
		"title":          doc.Title,
		"content":        doc.Content,
		"extracted_text": doc.ExtractedText,
		"mime_type":      doc.MimeType,
		"size_bytes":     doc.SizeBytes,
		"page_count":     doc.PageCount,
		"author":         doc.Author,
		"uploaded_by":    doc.UploadedBy,
		"created_at":     doc.CreatedAt,
		"updated_at":     doc.UpdatedAt,
		"tags":           doc.Tags,
		"categories":     doc.Categories,
		"permissions":    doc.Permissions,
		"metadata":       doc.Metadata,
		"checksum":       doc.Checksum,
		"storage_path":   doc.StoragePath,
		"is_deleted":     doc.IsDeleted,
	}

	// Create full-text search field combining all text content
	var fullTextParts []string
	if doc.Title != "" {
		fullTextParts = append(fullTextParts, doc.Title)
	}
	if doc.Content != "" {
		fullTextParts = append(fullTextParts, doc.Content)
	}
	if doc.ExtractedText != "" {
		fullTextParts = append(fullTextParts, doc.ExtractedText)
	}
	if doc.FileName != "" {
		fullTextParts = append(fullTextParts, doc.FileName)
	}
	prepared["full_text"] = strings.Join(fullTextParts, " ")

	// Add file extension for faceting
	if doc.FileName != "" {
		parts := strings.Split(doc.FileName, ".")
		if len(parts) > 1 {
			prepared["file_extension"] = strings.ToLower(parts[len(parts)-1])
		}
	}

	// Add size category for faceting
	prepared["size_category"] = getSizeCategory(doc.SizeBytes)

	// Call base preparation to add common fields
	return d.BaseIndexer.PrepareDocument(prepared)
}

// IndexDocumentFromFormAttachment indexes a document from form attachment
func (d *DocumentIndexer) IndexDocumentFromFormAttachment(ctx context.Context, attachment interface{}) error {
	// This would typically convert from your FormAttachment model to DocumentData
	// Implementation depends on your actual FormAttachment structure
	docData := &DocumentData{
		// Map fields from attachment to DocumentData
		// This is a placeholder - implement based on your actual model
	}

	return d.IndexDocument(ctx, docData)
}

// getSizeCategory categorizes file sizes
func getSizeCategory(sizeBytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case sizeBytes < KB:
		return "bytes"
	case sizeBytes < MB:
		return "kb"
	case sizeBytes < 10*MB:
		return "small_mb"
	case sizeBytes < 100*MB:
		return "medium_mb"
	case sizeBytes < GB:
		return "large_mb"
	default:
		return "gb"
	}
}

// getDocumentIndexConfig returns specific configuration for document index
func getDocumentIndexConfig() *models.IndexConfig {
	config := &models.IndexConfig{
		Name: "documents",
		Settings: models.IndexSettings{
			NumberOfShards:   2,
			NumberOfReplicas: 1,
			RefreshInterval:  "1s",
			MaxResultWindow:  50000,
			Analysis: models.AnalysisSettings{
				Analyzers: map[string]models.Analyzer{
					"document_analyzer": {
						Type:      "custom",
						Tokenizer: "standard",
						Filters:   []string{"lowercase", "stop", "stemmer"},
					},
					"filename_analyzer": {
						Type:      "custom",
						Tokenizer: "keyword",
						Filters:   []string{"lowercase"},
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
				"file_name": {
					Type:     "text",
					Analyzer: "filename_analyzer",
					Index:    true,
					Fields: map[string]models.FieldMapping{
						"keyword": {
							Type:  "keyword",
							Index: true,
						},
					},
				},
				"title": {
					Type:     "text",
					Analyzer: "document_analyzer",
					Index:    true,
				},
				"content": {
					Type:     "text",
					Analyzer: "document_analyzer",
					Index:    true,
				},
				"extracted_text": {
					Type:     "text",
					Analyzer: "document_analyzer",
					Index:    true,
				},
				"full_text": {
					Type:     "text",
					Analyzer: "document_analyzer",
					Index:    true,
				},
				"mime_type": {
					Type:  "keyword",
					Index: true,
				},
				"file_extension": {
					Type:  "keyword",
					Index: true,
				},
				"size_bytes": {
					Type:  "long",
					Index: true,
				},
				"size_category": {
					Type:  "keyword",
					Index: true,
				},
				"page_count": {
					Type:  "integer",
					Index: true,
				},
				"author": {
					Type:  "keyword",
					Index: true,
				},
				"uploaded_by": {
					Type:  "keyword",
					Index: true,
				},
				"tags": {
					Type:  "keyword",
					Index: true,
				},
				"categories": {
					Type:  "keyword",
					Index: true,
				},
				"permissions": {
					Type:  "keyword",
					Index: true,
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
				"is_deleted": {
					Type:  "boolean",
					Index: true,
				},
			},
		},
	}

	return config
}
