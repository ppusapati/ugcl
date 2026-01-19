package services

import (
	"database/sql"
	"fmt"

	"p9e.in/ugcl/core/helper"
	"p9e.in/ugcl/core/models"

	"gorm.io/gorm"
)

// QueryService provides generic Querying functionality for any GORM model
type QueryService[T any] struct {
	db    *gorm.DB
	model T
}

// NewQueryService creates a new generic Query service
func NewQueryService[T any](db *gorm.DB, model T) *QueryService[T] {
	return &QueryService[T]{
		db:    db,
		model: model,
	}
}

// GetQuery fetches Query data using the provided parameters
func (s *QueryService[T]) GetQuery(params *models.QueryParams) (*models.QueryResponse, error) {
	// Get JSON to DB column mapping using your existing function
	fmt.Println("params", params)
	jsonToDB, err := helper.BuildJSONtoDBColumnMap(s.db, s.model)
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

	return &models.QueryResponse{
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
		Data:  results,
	}, nil
}

// getDBFields converts JSON field names to DB column names
func (s *QueryService[T]) getDBFields(fields []string, jsonToDB map[string]string) []string {
	var dbFields []string
	for _, field := range fields {
		if dbCol, ok := jsonToDB[field]; ok {
			dbFields = append(dbFields, dbCol)
		}
	}
	return dbFields
}

// applyFilters applies all filters to the database query
func (s *QueryService[T]) applyFilters(query *gorm.DB, params *models.QueryParams, jsonToDB map[string]string) *gorm.DB {
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
func (s *QueryService[T]) processRows(rows *sql.Rows, jsonToDB map[string]string) ([]map[string]interface{}, error) {
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
func (s *QueryService[T]) findJSONField(dbCol string, jsonToDB map[string]string) string {
	for jsonName, dbName := range jsonToDB {
		if dbName == dbCol {
			return jsonName
		}
	}
	return dbCol // fallback to DB column name
}

// getTotalCount gets the total count with the same filters applied
func (s *QueryService[T]) getTotalCount(params *models.QueryParams, jsonToDB map[string]string) (int64, error) {
	countQuery := s.db.Model(&s.model)
	countQuery = s.applyFilters(countQuery, params, jsonToDB)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return total, nil
}
