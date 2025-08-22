package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	formbuilderv2 "p9e.in/ugcl/formbuilder/api/v2/form_builder"
)

// FieldDataService handles API calls for field options
type FieldDataService struct {
	db         *sql.DB
	httpClient *http.Client
	cache      *FieldOptionsCache
}

func NewFieldDataService(db *sql.DB) *FieldDataService {
	return &FieldDataService{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cache:      NewFieldOptionsCache(db),
	}
}

// GetFieldOptions fetches options for a dropdown/select field
func (fds *FieldDataService) GetFieldOptions(
	ctx context.Context,
	field *formbuilderv2.FormField,
	context map[string]string,
) ([]*formbuilderv2.FieldOption, error) {

	// Check cache first if enabled
	if field.CacheConfig != nil && field.CacheConfig.Enabled && !field.CacheConfig.RealTime {
		cached, err := fds.cache.Get(field.Id, context)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// Build API request
	apiConfig := field.ApiConfig
	if apiConfig == nil {
		apiConfig = &formbuilderv2.ApiConfig{
			Method:         "GET",
			TimeoutSeconds: 30,
			RetryCount:     3,
		}
	}

	// Construct URL with parameters
	apiUrl := field.OptionsSource
	if len(context) > 0 {
		params := url.Values{}
		for k, v := range context {
			params.Add(k, v)
		}
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, params.Encode())
	}

	// Make API call with retries
	var options []*formbuilderv2.FieldOption
	var lastErr error

	for i := 0; i <= int(apiConfig.RetryCount); i++ {
		req, err := http.NewRequestWithContext(ctx, apiConfig.Method, apiUrl, nil)
		if err != nil {
			return nil, err
		}

		// Add headers
		for k, v := range apiConfig.Headers {
			req.Header.Set(k, v)
		}

		// Add auth
		if err := fds.addAuthentication(req, apiConfig); err != nil {
			return nil, err
		}

		resp, err := fds.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i) * time.Second) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Parse response
			options, err = fds.parseResponse(resp.Body, apiConfig.ResponseTransform)
			if err != nil {
				return nil, err
			}
			break
		}

		lastErr = fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	if options == nil {
		return nil, lastErr
	}

	// Cache if enabled
	if field.CacheConfig != nil && field.CacheConfig.Enabled {
		fds.cache.Set(field.Id, context, options, field.CacheConfig.TtlSeconds)
	}

	return options, nil
}

func (fds *FieldDataService) addAuthentication(req *http.Request, config *formbuilderv2.ApiConfig) error {
	switch config.AuthType {
	case "bearer":
		token := config.Headers["Authorization"]
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "basic":
		username := config.Headers["username"]
		password := config.Headers["password"]
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
	case "api_key":
		key := config.Headers["X-API-Key"]
		if key != "" {
			req.Header.Set("X-API-Key", key)
		}
	}
	return nil
}

func (fds *FieldDataService) parseResponse(body io.Reader, transform string) ([]*formbuilderv2.FieldOption, error) {
	var rawData interface{}
	if err := json.NewDecoder(body).Decode(&rawData); err != nil {
		return nil, err
	}

	// Simple parsing - assumes array of objects with value/label
	// In production, would use the transform expression
	var options []*formbuilderv2.FieldOption

	if arr, ok := rawData.([]interface{}); ok {
		for _, item := range arr {
			if obj, ok := item.(map[string]interface{}); ok {
				option := &formbuilderv2.FieldOption{
					Value: fmt.Sprintf("%v", obj["value"]),
					Label: fmt.Sprintf("%v", obj["label"]),
				}
				if disabled, ok := obj["disabled"].(bool); ok {
					option.Disabled = disabled
				}
				options = append(options, option)
			}
		}
	}

	return options, nil
}
