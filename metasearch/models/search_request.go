package models

import (
	"time"
)

// SearchRequest represents a search query request
type SearchRequest struct {
	Query      string            `json:"query"`
	Filters    map[string]string `json:"filters,omitempty"`
	Facets     []string          `json:"facets,omitempty"`
	Sort       []SortField       `json:"sort,omitempty"`
	From       int               `json:"from"`
	Size       int               `json:"size"`
	Highlight  bool              `json:"highlight"`
	Suggest    bool              `json:"suggest"`
	Indices    []string          `json:"indices,omitempty"`
	TimeRange  *TimeRange        `json:"time_range,omitempty"`
}

// SortField represents sorting criteria
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// TimeRange represents a time-based filter
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// SuggestionRequest represents an auto-complete request
type SuggestionRequest struct {
	Query   string   `json:"query"`
	Indices []string `json:"indices,omitempty"`
	Size    int      `json:"size"`
}

// AdminRequest represents index administration request
type AdminRequest struct {
	Action    string                 `json:"action"` // "create", "delete", "reindex", "refresh"
	Index     string                 `json:"index"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	Mappings  map[string]interface{} `json:"mappings,omitempty"`
	ForceSync bool                   `json:"force_sync"`
}