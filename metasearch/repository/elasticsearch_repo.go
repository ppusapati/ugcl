package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"

	"p9e.in/ugcl/metasearch/config"
	"p9e.in/ugcl/metasearch/models"
)

// ElasticsearchRepository handles all Elasticsearch operations
type ElasticsearchRepository struct {
	client *elasticsearch.Client
	logger *zap.Logger
	config *config.ElasticsearchConfig
}

// NewElasticsearchRepository creates a new Elasticsearch repository
func NewElasticsearchRepository(cfg *config.ElasticsearchConfig, logger *zap.Logger) (*ElasticsearchRepository, error) {
	// Create Elasticsearch client configuration
	esConfig := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * 30,
			TLSHandshakeTimeout:   time.Second * 10,
		},
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			return time.Duration(i) * 100 * time.Millisecond
		},
		MaxRetries: 5,
	}

	// Create client
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch returned error: %s", res.String())
	}

	logger.Info("Connected to Elasticsearch successfully")

	return &ElasticsearchRepository{
		client: client,
		logger: logger,
		config: cfg,
	}, nil
}

// CreateIndex creates a new index with the specified configuration
func (r *ElasticsearchRepository) CreateIndex(ctx context.Context, config *models.IndexConfig) error {
	// Check if index already exists
	exists, err := r.IndexExists(ctx, config.Name)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if exists {
		return fmt.Errorf("index %s already exists", config.Name)
	}

	// Prepare index body
	indexBody := map[string]interface{}{
		"settings": config.Settings,
		"mappings": config.Mappings,
	}

	if len(config.Aliases) > 0 {
		aliases := make(map[string]interface{})
		for _, alias := range config.Aliases {
			aliases[alias] = map[string]interface{}{}
		}
		indexBody["aliases"] = aliases
	}

	// Convert to JSON
	body, err := json.Marshal(indexBody)
	if err != nil {
		return fmt.Errorf("failed to marshal index configuration: %w", err)
	}

	// Create index
	req := esapi.IndicesCreateRequest{
		Index: config.Name,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create index %s: %s", config.Name, string(body))
	}

	r.logger.Info("Index created successfully", zap.String("index", config.Name))
	return nil
}

// DeleteIndex deletes an index
func (r *ElasticsearchRepository) DeleteIndex(ctx context.Context, indexName string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to delete index %s: %s", indexName, string(body))
	}

	r.logger.Info("Index deleted successfully", zap.String("index", indexName))
	return nil
}

// IndexExists checks if an index exists
func (r *ElasticsearchRepository) IndexExists(ctx context.Context, indexName string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// IndexDocument indexes a single document
func (r *ElasticsearchRepository) IndexDocument(ctx context.Context, indexName, docID string, document map[string]interface{}) error {
	// Add timestamp
	document["@timestamp"] = time.Now()

	// Convert to JSON
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// Index document
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(body),
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to index document %s: %s", docID, string(body))
	}

	return nil
}

// UpdateDocument updates a document
func (r *ElasticsearchRepository) UpdateDocument(ctx context.Context, indexName, docID string, updates map[string]interface{}) error {
	// Add timestamp
	updates["@timestamp"] = time.Now()

	// Prepare update body
	updateBody := map[string]interface{}{
		"doc": updates,
	}

	body, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	// Update document
	req := esapi.UpdateRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(body),
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to update document %s: %s", docID, string(body))
	}

	return nil
}

// DeleteDocument deletes a document
func (r *ElasticsearchRepository) DeleteDocument(ctx context.Context, indexName, docID string) error {
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: docID,
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to delete document %s: %s", docID, string(body))
	}

	return nil
}

// BulkIndex performs bulk indexing operations
func (r *ElasticsearchRepository) BulkIndex(ctx context.Context, operations []BulkOperation) error {
	if len(operations) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, op := range operations {
		switch op.Action {
		case "index":
			// Index action header
			meta := map[string]interface{}{
				"index": map[string]interface{}{
					"_index": op.Index,
					"_id":    op.ID,
				},
			}
			metaJSON, _ := json.Marshal(meta)
			buf.Write(metaJSON)
			buf.WriteByte('\n')

			// Document body
			if op.Document != nil {
				op.Document["@timestamp"] = time.Now()
				docJSON, _ := json.Marshal(op.Document)
				buf.Write(docJSON)
				buf.WriteByte('\n')
			}

		case "update":
			// Update action header
			meta := map[string]interface{}{
				"update": map[string]interface{}{
					"_index": op.Index,
					"_id":    op.ID,
				},
			}
			metaJSON, _ := json.Marshal(meta)
			buf.Write(metaJSON)
			buf.WriteByte('\n')

			// Update body
			if op.Updates != nil {
				op.Updates["@timestamp"] = time.Now()
				updateBody := map[string]interface{}{
					"doc": op.Updates,
				}
				updateJSON, _ := json.Marshal(updateBody)
				buf.Write(updateJSON)
				buf.WriteByte('\n')
			}

		case "delete":
			// Delete action header
			meta := map[string]interface{}{
				"delete": map[string]interface{}{
					"_index": op.Index,
					"_id":    op.ID,
				},
			}
			metaJSON, _ := json.Marshal(meta)
			buf.Write(metaJSON)
			buf.WriteByte('\n')
		}
	}

	// Perform bulk request
	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "wait_for",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to perform bulk operation: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk operation failed: %s", string(body))
	}

	// Parse response to check for errors
	var bulkResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&bulkResponse); err != nil {
		return fmt.Errorf("failed to parse bulk response: %w", err)
	}

	if errors, hasErrors := bulkResponse["errors"].(bool); hasErrors && errors {
		r.logger.Warn("Some bulk operations failed",
			zap.Int("total_operations", len(operations)))
	}

	return nil
}

// Search performs a search query
func (r *ElasticsearchRepository) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	// Build search query
	query := r.buildSearchQuery(req)

	// Convert to JSON
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	// Determine indices
	indices := req.Indices
	if len(indices) == 0 {
		indices = []string{"_all"}
	}

	// Perform search
	searchReq := esapi.SearchRequest{
		Index: indices,
		Body:  bytes.NewReader(body),
	}

	res, err := searchReq.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search failed: %s", string(body))
	}

	// Parse response
	var searchResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return r.parseSearchResponse(searchResponse), nil
}

// RefreshIndex refreshes an index
func (r *ElasticsearchRepository) RefreshIndex(ctx context.Context, indexName string) error {
	req := esapi.IndicesRefreshRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to refresh index %s: %s", indexName, string(body))
	}

	return nil
}

// GetIndexStats gets statistics for an index
func (r *ElasticsearchRepository) GetIndexStats(ctx context.Context, indexName string) (*models.IndexStats, error) {
	req := esapi.IndicesStatsRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get index stats: %s", string(body))
	}

	var statsResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&statsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse stats response: %w", err)
	}

	return r.parseIndexStats(indexName, statsResponse), nil
}

// Helper functions

func (r *ElasticsearchRepository) buildSearchQuery(req *models.SearchRequest) map[string]interface{} {
	query := map[string]interface{}{
		"from": req.From,
		"size": req.Size,
	}

	// Build main query
	if req.Query != "" || len(req.Filters) > 0 {
		boolQuery := map[string]interface{}{}

		// Main text query
		if req.Query != "" {
			boolQuery["must"] = []interface{}{
				map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":     req.Query,
						"fields":    []string{"title^2", "content", "full_text", "file_name"},
						"type":      "best_fields",
						"fuzziness": "AUTO",
					},
				},
			}
		}

		// Filters
		if len(req.Filters) > 0 {
			filters := make([]interface{}, 0, len(req.Filters))
			for field, value := range req.Filters {
				filters = append(filters, map[string]interface{}{
					"term": map[string]interface{}{
						field: value,
					},
				})
			}
			boolQuery["filter"] = filters
		}

		// Time range filter
		if req.TimeRange != nil {
			timeFilter := map[string]interface{}{
				"range": map[string]interface{}{
					"@timestamp": map[string]interface{}{
						"gte": req.TimeRange.From,
						"lte": req.TimeRange.To,
					},
				},
			}
			if boolQuery["filter"] == nil {
				boolQuery["filter"] = []interface{}{timeFilter}
			} else {
				boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), timeFilter)
			}
		}

		query["query"] = map[string]interface{}{
			"bool": boolQuery,
		}
	}

	// Sorting
	if len(req.Sort) > 0 {
		sorts := make([]interface{}, 0, len(req.Sort))
		for _, sort := range req.Sort {
			sorts = append(sorts, map[string]interface{}{
				sort.Field: map[string]interface{}{
					"order": sort.Order,
				},
			})
		}
		query["sort"] = sorts
	}

	// Highlighting
	if req.Highlight {
		query["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"title":     map[string]interface{}{},
				"content":   map[string]interface{}{},
				"full_text": map[string]interface{}{},
				"file_name": map[string]interface{}{},
			},
			"pre_tags":  []string{"<mark>"},
			"post_tags": []string{"</mark>"},
		}
	}

	// Aggregations for facets
	if len(req.Facets) > 0 {
		aggs := make(map[string]interface{})
		for _, facet := range req.Facets {
			aggs[facet] = map[string]interface{}{
				"terms": map[string]interface{}{
					"field": facet,
					"size":  20,
				},
			}
		}
		query["aggs"] = aggs
	}

	return query
}

func (r *ElasticsearchRepository) parseSearchResponse(response map[string]interface{}) *models.SearchResponse {
	result := &models.SearchResponse{
		Hits:   []models.SearchHit{},
		Facets: make(map[string]models.Facet),
	}

	// Parse hits
	if hits, ok := response["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				result.Total = int64(value)
			}
		}

		if maxScore, ok := hits["max_score"].(float64); ok {
			result.MaxScore = maxScore
		}

		if hitsList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsList {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					searchHit := models.SearchHit{
						ID:    getString(hitMap, "_id"),
						Index: getString(hitMap, "_index"),
						Type:  getString(hitMap, "_type"),
					}

					if score, ok := hitMap["_score"].(float64); ok {
						searchHit.Score = score
					}

					if source, ok := hitMap["_source"].(map[string]interface{}); ok {
						searchHit.Source = source
					}

					if highlight, ok := hitMap["highlight"].(map[string]interface{}); ok {
						searchHit.Highlight = make(map[string][]string)
						for field, highlights := range highlight {
							if highlightList, ok := highlights.([]interface{}); ok {
								highlightStrings := make([]string, 0, len(highlightList))
								for _, hl := range highlightList {
									if hlStr, ok := hl.(string); ok {
										highlightStrings = append(highlightStrings, hlStr)
									}
								}
								searchHit.Highlight[field] = highlightStrings
							}
						}
					}

					result.Hits = append(result.Hits, searchHit)
				}
			}
		}
	}

	// Parse aggregations
	if aggs, ok := response["aggregations"].(map[string]interface{}); ok {
		for field, agg := range aggs {
			if aggMap, ok := agg.(map[string]interface{}); ok {
				if buckets, ok := aggMap["buckets"].([]interface{}); ok {
					facet := models.Facet{
						Field:   field,
						Buckets: make([]models.FacetBucket, 0, len(buckets)),
					}

					for _, bucket := range buckets {
						if bucketMap, ok := bucket.(map[string]interface{}); ok {
							facetBucket := models.FacetBucket{
								Key: getString(bucketMap, "key"),
							}

							if docCount, ok := bucketMap["doc_count"].(float64); ok {
								facetBucket.DocCount = int64(docCount)
							}

							facet.Buckets = append(facet.Buckets, facetBucket)
						}
					}

					result.Facets[field] = facet
				}
			}
		}
	}

	// Parse timing
	if took, ok := response["took"].(float64); ok {
		result.Took = int64(took)
	}

	if timedOut, ok := response["timed_out"].(bool); ok {
		result.TimedOut = timedOut
	}

	return result
}

func (r *ElasticsearchRepository) parseIndexStats(indexName string, response map[string]interface{}) *models.IndexStats {
	stats := &models.IndexStats{
		IndexName: indexName,
	}

	if indices, ok := response["indices"].(map[string]interface{}); ok {
		if indexStats, ok := indices[indexName].(map[string]interface{}); ok {
			if primaries, ok := indexStats["primaries"].(map[string]interface{}); ok {
				if docs, ok := primaries["docs"].(map[string]interface{}); ok {
					if count, ok := docs["count"].(float64); ok {
						stats.DocCount = int64(count)
					}
				}

				if store, ok := primaries["store"].(map[string]interface{}); ok {
					if size, ok := store["size_in_bytes"].(float64); ok {
						stats.StoreSize = fmt.Sprintf("%.2f MB", size/1024/1024)
					}
				}
			}
		}
	}

	stats.LastUpdated = time.Now()
	stats.Health = "green" // Default, would need cluster health API for actual status

	return stats
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// BulkOperation represents a bulk operation
type BulkOperation struct {
	Action   string                 `json:"action"`
	Index    string                 `json:"index"`
	ID       string                 `json:"id"`
	Document map[string]interface{} `json:"document,omitempty"`
	Updates  map[string]interface{} `json:"updates,omitempty"`
}
