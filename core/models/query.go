package models

import "fmt"

// QueryParams holds all query parameters for any report request
type QueryParams struct {
	Page       int    `validate:"min=1"`
	Limit      int    `validate:"min=1,max=1000"`
	FromDate   string `validate:"omitempty,datetime"`
	ToDate     string `validate:"omitempty,datetime"`
	Fields     []string
	Filters    map[string]interface{} // Generic filters for any field
	DateColumn string                 // Configurable date column (default: "created_at")
}

// QueryResponse represents the API response structure
type QueryResponse struct {
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
	Data  []map[string]interface{} `json:"data"`
}

// Validate performs business logic validation on the parameters
func (p *QueryParams) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}

	if p.Limit < 1 {
		return fmt.Errorf("limit must be greater than 0")
	}

	if p.Limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}

	// Basic date format validation
	if p.FromDate != "" && len(p.FromDate) < 10 {
		return fmt.Errorf("fromDate must be in YYYY-MM-DD format")
	}

	if p.ToDate != "" && len(p.ToDate) < 10 {
		return fmt.Errorf("toDate must be in YYYY-MM-DD format")
	}

	if p.DateColumn == "" {
		p.DateColumn = "created_at" // Set default if empty
	}

	return nil
}

// GetOffset calculates the database offset based on page and limit
func (p *QueryParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// HasDateFilter returns true if any date filter is applied
func (p *QueryParams) HasDateFilter() bool {
	return p.FromDate != "" || p.ToDate != ""
}

// HasFilters returns true if any filters are applied
func (p *QueryParams) HasFilters() bool {
	return p.HasDateFilter() || len(p.Filters) > 0
}
