package handlers

import (
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	metasearchpb "p9e.in/ugcl/metasearch/api/v1/metasearch"
	"p9e.in/ugcl/metasearch/models"
)

// convertProtoToSearchRequest converts a proto SearchRequest to a service SearchRequest
func (h *SearchHandler) convertProtoToSearchRequest(proto *metasearchpb.SearchRequest) *models.SearchRequest {
	req := &models.SearchRequest{
		Query:     proto.Query,
		Filters:   proto.Filters,
		Facets:    proto.Facets,
		From:      int(proto.From),
		Size:      int(proto.Size),
		Highlight: proto.Highlight,
		Suggest:   proto.Suggest,
		Indices:   proto.Indices,
	}

	// Convert sort fields
	for _, sortField := range proto.Sort {
		req.Sort = append(req.Sort, models.SortField{
			Field: sortField.Field,
			Order: sortField.Order,
		})
	}

	// Convert time range
	if proto.TimeRange != nil {
		req.TimeRange = h.convertProtoTimeRange(proto.TimeRange)
	}

	return req
}

// convertSearchResponseToProto converts a service SearchResponse to a proto SearchResponse
func (h *SearchHandler) convertSearchResponseToProto(response *models.SearchResponse) *metasearchpb.SearchResponse {
	proto := &metasearchpb.SearchResponse{
		Total:    response.Total,
		MaxScore: response.MaxScore,
		Took:     response.Took,
		TimedOut: response.TimedOut,
		ScrollId: response.ScrollID,
	}

	// Convert hits
	for _, hit := range response.Hits {
		protoHit := &metasearchpb.SearchHit{
			Id:    hit.ID,
			Index: hit.Index,
			Type:  hit.Type,
			Score: hit.Score,
		}

		// Convert source to proto struct
		if hit.Source != nil {
			if source, err := structpb.NewStruct(hit.Source); err == nil {
				protoHit.Source = source
			}
		}

		// Convert highlight
		if hit.Highlight != nil {
			protoHit.Highlight = make(map[string]*metasearchpb.HighlightField)
			for field, fragments := range hit.Highlight {
				protoHit.Highlight[field] = &metasearchpb.HighlightField{
					Fragments: fragments,
				}
			}
		}

		// Convert sort values
		for _, sortValue := range hit.Sort {
			if value, err := structpb.NewValue(sortValue); err == nil {
				protoHit.Sort = append(protoHit.Sort, value)
			}
		}

		proto.Hits = append(proto.Hits, protoHit)
	}

	// Convert facets
	if response.Facets != nil {
		proto.Facets = make(map[string]*metasearchpb.Facet)
		for field, facet := range response.Facets {
			protoFacet := &metasearchpb.Facet{
				Field: facet.Field,
			}
			for _, bucket := range facet.Buckets {
				protoFacet.Buckets = append(protoFacet.Buckets, &metasearchpb.FacetBucket{
					Key:      bucket.Key,
					DocCount: bucket.DocCount,
				})
			}
			proto.Facets[field] = protoFacet
		}
	}

	// Convert suggestions
	proto.Suggestions = h.convertSuggestionsToProto(response.Suggestions)

	return proto
}

// convertSuggestionsToProto converts service suggestions to proto suggestions
func (h *SearchHandler) convertSuggestionsToProto(suggestions []models.Suggestion) []*metasearchpb.Suggestion {
	var protoSuggestions []*metasearchpb.Suggestion
	for _, suggestion := range suggestions {
		protoSugg := &metasearchpb.Suggestion{
			Text: suggestion.Text,
		}
		for _, option := range suggestion.Options {
			protoOption := &metasearchpb.SuggestionOption{
				Text:  option.Text,
				Score: option.Score,
			}
			if option.Source != nil {
				if source, err := structpb.NewStruct(option.Source); err == nil {
					protoOption.Source = source
				}
			}
			protoSugg.Options = append(protoSugg.Options, protoOption)
		}
		protoSuggestions = append(protoSuggestions, protoSugg)
	}
	return protoSuggestions
}

// convertProtoTimeRange converts a proto TimeRange to a service TimeRange
func (h *SearchHandler) convertProtoTimeRange(proto *metasearchpb.TimeRange) *models.TimeRange {
	if proto == nil {
		return nil
	}

	timeRange := &models.TimeRange{}
	if proto.From != nil {
		timeRange.From = proto.From.AsTime()
	}
	if proto.To != nil {
		timeRange.To = proto.To.AsTime()
	}
	return timeRange
}

// convertProtoStructToMap converts a proto Struct to a map[string]interface{}
func (h *SearchHandler) convertProtoStructToMap(proto *structpb.Struct) map[string]interface{} {
	if proto == nil {
		return nil
	}
	return proto.AsMap()
}

// convertProtoToIndexConfig converts a proto IndexConfig to a service IndexConfig
func (h *SearchHandler) convertProtoToIndexConfig(proto *metasearchpb.IndexConfig) *models.IndexConfig {
	config := &models.IndexConfig{
		Name:      proto.Name,
		Aliases:   proto.Aliases,
		Version:   int(proto.Version),
		IsActive:  proto.IsActive,
	}

	if proto.CreatedAt != nil {
		config.CreatedAt = proto.CreatedAt.AsTime()
	}
	if proto.UpdatedAt != nil {
		config.UpdatedAt = proto.UpdatedAt.AsTime()
	}

	// Convert settings
	if proto.Settings != nil {
		config.Settings = models.IndexSettings{
			NumberOfShards:   int(proto.Settings.NumberOfShards),
			NumberOfReplicas: int(proto.Settings.NumberOfReplicas),
			RefreshInterval:  proto.Settings.RefreshInterval,
			MaxResultWindow:  int(proto.Settings.MaxResultWindow),
		}
		if proto.Settings.Analysis != nil {
			// TODO: Convert proto.Settings.Analysis.AsMap() to proper AnalysisSettings
			// For now, use empty settings
			config.Settings.Analysis = models.AnalysisSettings{}
		}
	}

	// Convert mappings
	if proto.Mappings != nil {
		config.Mappings = models.IndexMappings{
			Properties: make(map[string]models.FieldMapping),
		}
		for field, mapping := range proto.Mappings.Properties {
			config.Mappings.Properties[field] = *h.convertProtoToFieldMapping(mapping)
		}
	}

	return config
}

// convertProtoToFieldMapping converts a proto FieldMapping to a service FieldMapping
func (h *SearchHandler) convertProtoToFieldMapping(proto *metasearchpb.FieldMapping) *models.FieldMapping {
	mapping := &models.FieldMapping{
		Type:     proto.Type,
		Analyzer: proto.Analyzer,
		Index:    proto.Index,
		Store:    proto.Store,
	}

	// Convert nested fields
	if len(proto.Fields) > 0 {
		mapping.Fields = make(map[string]models.FieldMapping)
		for field, fieldMapping := range proto.Fields {
			mapping.Fields[field] = *h.convertProtoToFieldMapping(fieldMapping)
		}
	}

	// Convert nested properties
	if len(proto.Properties) > 0 {
		mapping.Properties = make(map[string]models.FieldMapping)
		for prop, propMapping := range proto.Properties {
			mapping.Properties[prop] = *h.convertProtoToFieldMapping(propMapping)
		}
	}

	return mapping
}

// convertIndexStatsToProto converts a service IndexStats to a proto IndexStats
func (h *SearchHandler) convertIndexStatsToProto(stats *models.IndexStats) *metasearchpb.IndexStats {
	proto := &metasearchpb.IndexStats{
		IndexName: stats.IndexName,
		DocCount:  stats.DocCount,
		StoreSize: stats.StoreSize,
		Health:    stats.Health,
	}

	if !stats.LastUpdated.IsZero() {
		proto.LastUpdated = timestamppb.New(stats.LastUpdated)
	}

	return proto
}

// convertIndexingJobToProto converts a service IndexingJob to a proto IndexingJob
func (h *SearchHandler) convertIndexingJobToProto(job *models.IndexingJob) *metasearchpb.IndexingJob {
	proto := &metasearchpb.IndexingJob{
		Id:        job.ID,
		Type:      job.Type,
		Status:    job.Status,
		IndexName: job.IndexName,
		Error:     job.Error,
	}

	if job.StartedAt != nil && !job.StartedAt.IsZero() {
		proto.StartedAt = timestamppb.New(*job.StartedAt)
	}
	if job.CompletedAt != nil && !job.CompletedAt.IsZero() {
		proto.CompletedAt = timestamppb.New(*job.CompletedAt)
	}

	// Convert progress
	proto.Progress = &metasearchpb.IndexingProgress{
		TotalRecords:     job.Progress.TotalRecords,
		ProcessedRecords: job.Progress.ProcessedRecords,
		FailedRecords:    job.Progress.FailedRecords,
		Percentage:       job.Progress.Percentage,
	}

	// Convert config
	if job.Config != nil {
		if config, err := structpb.NewStruct(job.Config); err == nil {
			proto.Config = config
		}
	}

	return proto
}

// convertSearchAnalyticsToProto converts a service SearchAnalytics to a proto SearchAnalytics
func (h *SearchHandler) convertSearchAnalyticsToProto(analytics *models.SearchAnalytics) *metasearchpb.SearchAnalytics {
	proto := &metasearchpb.SearchAnalytics{
		TotalSearches:     analytics.TotalSearches,
		AverageResponse:   analytics.AverageResponse,
		NoResultQueries:   analytics.NoResultQueries,
		SearchesByIndex:   analytics.SearchesByIndex,
	}

	// Convert popular queries
	for _, query := range analytics.PopularQueries {
		proto.PopularQueries = append(proto.PopularQueries, &metasearchpb.QueryFrequency{
			Query:     query.Query,
			Frequency: query.Frequency,
			Index:     query.Index,
		})
	}

	// Convert time range
	proto.TimeRange = &metasearchpb.TimeRange{}
	if !analytics.TimeRange.From.IsZero() {
		proto.TimeRange.From = timestamppb.New(analytics.TimeRange.From)
	}
	if !analytics.TimeRange.To.IsZero() {
		proto.TimeRange.To = timestamppb.New(analytics.TimeRange.To)
	}

	return proto
}