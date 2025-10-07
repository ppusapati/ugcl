package models

import (
	"time"
)

// SearchResponse represents the response from a search query
type SearchResponse struct {
	Hits        []SearchHit           `json:"hits"`
	Total       int64                 `json:"total"`
	MaxScore    float64               `json:"max_score"`
	Took        int64                 `json:"took"`
	TimedOut    bool                  `json:"timed_out"`
	Facets      map[string]Facet      `json:"facets,omitempty"`
	Suggestions []Suggestion          `json:"suggestions,omitempty"`
	ScrollID    string                `json:"scroll_id,omitempty"`
}

// SearchHit represents a single search result
type SearchHit struct {
	ID          string                 `json:"id"`
	Index       string                 `json:"index"`
	Type        string                 `json:"type"`
	Score       float64                `json:"score"`
	Source      map[string]interface{} `json:"source"`
	Highlight   map[string][]string    `json:"highlight,omitempty"`
	Sort        []interface{}          `json:"sort,omitempty"`
}

// Facet represents aggregated search facets
type Facet struct {
	Field   string        `json:"field"`
	Buckets []FacetBucket `json:"buckets"`
}

// FacetBucket represents a facet bucket
type FacetBucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

// Suggestion represents auto-complete suggestions
type Suggestion struct {
	Text    string             `json:"text"`
	Options []SuggestionOption `json:"options"`
}

// SuggestionOption represents a single suggestion option
type SuggestionOption struct {
	Text   string                 `json:"text"`
	Score  float64                `json:"score"`
	Source map[string]interface{} `json:"source,omitempty"`
}

// IndexStats represents statistics for an index
type IndexStats struct {
	IndexName   string    `json:"index_name"`
	DocCount    int64     `json:"doc_count"`
	StoreSize   string    `json:"store_size"`
	LastUpdated time.Time `json:"last_updated"`
	Health      string    `json:"health"`
}

// SearchAnalytics represents search usage analytics
type SearchAnalytics struct {
	TotalSearches    int64             `json:"total_searches"`
	PopularQueries   []QueryFrequency  `json:"popular_queries"`
	SearchesByIndex  map[string]int64  `json:"searches_by_index"`
	AverageResponse  float64           `json:"average_response_time"`
	NoResultQueries  []string          `json:"no_result_queries"`
	TimeRange        TimeRange         `json:"time_range"`
}

// QueryFrequency represents query frequency statistics
type QueryFrequency struct {
	Query     string `json:"query"`
	Frequency int64  `json:"frequency"`
	Index     string `json:"index"`
}