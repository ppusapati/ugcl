package helper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"p9e.in/ugcl/core/models"
)

// ParseQueryParams extracts and validates query parameters from HTTP request
func ParseQueryParams(r *http.Request) (*models.QueryParams, error) {
	params := &models.QueryParams{
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
