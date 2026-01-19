package models

import (
	"time"
)

// IndexConfig represents the configuration for an Elasticsearch index
type IndexConfig struct {
	Name        string                 `json:"name"`
	Settings    IndexSettings          `json:"settings"`
	Mappings    IndexMappings          `json:"mappings"`
	Aliases     []string               `json:"aliases,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     int                    `json:"version"`
	IsActive    bool                   `json:"is_active"`
}

// IndexSettings represents Elasticsearch index settings
type IndexSettings struct {
	NumberOfShards   int               `json:"number_of_shards"`
	NumberOfReplicas int               `json:"number_of_replicas"`
	RefreshInterval  string            `json:"refresh_interval"`
	MaxResultWindow  int               `json:"max_result_window"`
	Analysis         AnalysisSettings  `json:"analysis,omitempty"`
}

// AnalysisSettings represents text analysis settings
type AnalysisSettings struct {
	Analyzers  map[string]Analyzer  `json:"analyzers,omitempty"`
	Tokenizers map[string]Tokenizer `json:"tokenizers,omitempty"`
	Filters    map[string]Filter    `json:"filters,omitempty"`
}

// Analyzer represents a text analyzer configuration
type Analyzer struct {
	Type      string   `json:"type"`
	Tokenizer string   `json:"tokenizer,omitempty"`
	Filters   []string `json:"filters,omitempty"`
}

// Tokenizer represents a tokenizer configuration
type Tokenizer struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern,omitempty"`
}

// Filter represents a token filter configuration
type Filter struct {
	Type     string   `json:"type"`
	Synonyms []string `json:"synonyms,omitempty"`
}

// IndexMappings represents Elasticsearch field mappings
type IndexMappings struct {
	Properties map[string]FieldMapping `json:"properties"`
}

// FieldMapping represents a field mapping configuration
type FieldMapping struct {
	Type       string                     `json:"type"`
	Analyzer   string                     `json:"analyzer,omitempty"`
	Index      bool                       `json:"index"`
	Store      bool                       `json:"store,omitempty"`
	Fields     map[string]FieldMapping    `json:"fields,omitempty"`
	Properties map[string]FieldMapping    `json:"properties,omitempty"`
}

// IndexTemplate represents an index template for dynamic index creation
type IndexTemplate struct {
	Name         string        `json:"name"`
	IndexPattern string        `json:"index_pattern"`
	Settings     IndexSettings `json:"settings"`
	Mappings     IndexMappings `json:"mappings"`
	Priority     int           `json:"priority"`
	Version      int           `json:"version"`
}

// IndexingJob represents a background indexing job
type IndexingJob struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "full", "incremental", "document"
	Status      string                 `json:"status"` // "pending", "running", "completed", "failed"
	IndexName   string                 `json:"index_name"`
	Progress    IndexingProgress       `json:"progress"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// IndexingProgress represents the progress of an indexing job
type IndexingProgress struct {
	TotalRecords     int64 `json:"total_records"`
	ProcessedRecords int64 `json:"processed_records"`
	FailedRecords    int64 `json:"failed_records"`
	Percentage       float64 `json:"percentage"`
}