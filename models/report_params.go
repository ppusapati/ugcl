package models

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// ReportParams holds all query parameters for any report request
type ReportParams struct {
	Page       int    `validate:"min=1"`
	Limit      int    `validate:"min=1,max=1000"`
	FromDate   string `validate:"omitempty,datetime"`
	ToDate     string `validate:"omitempty,datetime"`
	Fields     []string
	Filters    map[string]interface{} // Generic filters for any field
	DateColumn string                 // Configurable date column (default: "created_at")
}

// ReportResponse represents the API response structure
type ReportResponse struct {
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
	Data  []map[string]interface{} `json:"data"`
}

// ReportService provides generic reporting functionality for any GORM model
type ReportService[T any] struct {
	db    *gorm.DB
	model T
}

// NewReportService creates a new generic report service
func NewReportService[T any](db *gorm.DB, model T) *ReportService[T] {
	return &ReportService[T]{
		db:    db,
		model: model,
	}
}

// ParseReportParams extracts and validates query parameters from HTTP request
func ParseReportParams(r *http.Request) (*ReportParams, error) {
	params := &ReportParams{
		Page:       1,
		Limit:      10,
		Filters:    make(map[string]interface{}),
		DateColumn: "created_at", // default date column
	}

	query := r.URL.Query()

	// Parse page parameter
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err != nil {
			return nil, fmt.Errorf("invalid page parameter: %s (must be a number)", pageStr)
		} else if p < 1 {
			return nil, fmt.Errorf("invalid page parameter: %d (must be greater than 0)", p)
		} else {
			params.Page = p
		}
	}

	// Parse limit parameter
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err != nil {
			return nil, fmt.Errorf("invalid limit parameter: %s (must be a number)", limitStr)
		} else if l < 1 {
			return nil, fmt.Errorf("invalid limit parameter: %d (must be greater than 0)", l)
		} else {
			if l > 1000 {
				l = 1000 // Cap at 1000
			}
			params.Limit = l
		}
	}

	// Parse fields parameter
	if fieldsStr := query.Get("fields"); fieldsStr != "" {
		fields := strings.Split(fieldsStr, ",")
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}
		params.Fields = fields
	}

	// Parse date parameters
	params.FromDate = strings.TrimSpace(query.Get("fromDate"))
	params.ToDate = strings.TrimSpace(query.Get("toDate"))

	// Parse custom date column if provided
	if dateCol := query.Get("dateColumn"); dateCol != "" {
		params.DateColumn = strings.TrimSpace(dateCol)
	}

	// Parse generic filters (any other query parameters)
	reservedParams := map[string]bool{
		"page": true, "limit": true, "fields": true,
		"fromDate": true, "toDate": true, "dateColumn": true,
	}

	for key, values := range query {
		if !reservedParams[key] && len(values) > 0 {
			value := strings.TrimSpace(values[0])
			if value != "" {
				params.Filters[key] = value
			}
		}
	}

	return params, nil
}

// Validate performs business logic validation on the parameters
func (p *ReportParams) Validate() error {
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
func (p *ReportParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// HasDateFilter returns true if any date filter is applied
func (p *ReportParams) HasDateFilter() bool {
	return p.FromDate != "" || p.ToDate != ""
}

// HasFilters returns true if any filters are applied
func (p *ReportParams) HasFilters() bool {
	return p.HasDateFilter() || len(p.Filters) > 0
}

// GetReport fetches report data using the provided parameters
func (s *ReportService[T]) GetReport(params *ReportParams) (*ReportResponse, error) {
	// Get JSON to DB column mapping using your existing function
	jsonToDB, err := BuildJSONtoDBColumnMap(s.db, s.model)
	if err != nil {
		return nil, fmt.Errorf("failed to get column mapping: %w", err)
	}

	// Build base query
	query := s.db.Model(&s.model)

	// Apply field selection
	if len(params.Fields) > 0 {
		dbFields := s.getDBFields(params.Fields, jsonToDB)
		if len(dbFields) > 0 {
			query = query.Select(dbFields)
		}
	}

	// Apply filters
	query = s.applyFilters(query, params, jsonToDB)

	// Execute main query
	rows, err := query.Limit(params.Limit).Offset(params.GetOffset()).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Process results
	results, err := s.processRows(rows, jsonToDB)
	if err != nil {
		return nil, fmt.Errorf("failed to process rows: %w", err)
	}

	// Get total count with same filters
	total, err := s.getTotalCount(params, jsonToDB)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	return &ReportResponse{
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
		Data:  results,
	}, nil
}

// getDBFields converts JSON field names to DB column names
func (s *ReportService[T]) getDBFields(fields []string, jsonToDB map[string]string) []string {
	var dbFields []string
	for _, field := range fields {
		if dbCol, ok := jsonToDB[field]; ok {
			dbFields = append(dbFields, dbCol)
		}
	}
	return dbFields
}

// applyFilters applies all filters to the database query
func (s *ReportService[T]) applyFilters(query *gorm.DB, params *ReportParams, jsonToDB map[string]string) *gorm.DB {
	// Apply date filters
	if params.HasDateFilter() {
		if params.FromDate != "" && params.ToDate != "" {
			query = query.Where(params.DateColumn+" BETWEEN ? AND ?", params.FromDate, params.ToDate)
		} else if params.FromDate != "" {
			query = query.Where(params.DateColumn+" >= ?", params.FromDate)
		} else if params.ToDate != "" {
			query = query.Where(params.DateColumn+" <= ?", params.ToDate)
		}
	}

	// Apply generic filters
	for jsonField, value := range params.Filters {
		if dbCol, ok := jsonToDB[jsonField]; ok {
			// Use the DB column name for the query
			query = query.Where(dbCol+" = ?", value)
		} else {
			// If not in mapping, try using the field name directly
			query = query.Where(jsonField+" = ?", value)
		}
	}

	return query
}

// processRows converts database rows to map slice
func (s *ReportService[T]) processRows(rows *sql.Rows, jsonToDB map[string]string) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create slice for scanning
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to map with JSON field names
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			// Find JSON field name for this DB column
			jsonField := s.findJSONField(col, jsonToDB)
			rowMap[jsonField] = values[i]
		}
		results = append(results, rowMap)
	}

	return results, nil
}

// findJSONField finds the JSON field name for a given DB column
func (s *ReportService[T]) findJSONField(dbCol string, jsonToDB map[string]string) string {
	for jsonName, dbName := range jsonToDB {
		if dbName == dbCol {
			return jsonName
		}
	}
	return dbCol // fallback to DB column name
}

// getTotalCount gets the total count with the same filters applied
func (s *ReportService[T]) getTotalCount(params *ReportParams, jsonToDB map[string]string) (int64, error) {
	countQuery := s.db.Model(&s.model)
	countQuery = s.applyFilters(countQuery, params, jsonToDB)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return total, nil
}

// BuildJSONtoDBColumnMap builds mapping between JSON field names and DB column names
// This uses your existing implementation which is more robust
func BuildJSONtoDBColumnMap(db *gorm.DB, model interface{}) (map[string]string, error) {
	// Get schema
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, field := range stmt.Schema.Fields {
		jsonName := field.Tag.Get("json")
		if idx := strings.Index(jsonName, ","); idx != -1 {
			jsonName = jsonName[:idx]
		}
		dbCol := field.DBName
		if jsonName != "" && dbCol != "" && jsonName != "-" {
			out[jsonName] = dbCol
		}
	}

	return out, nil
}
