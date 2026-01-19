package handlers

import (
	"context"

	"connectrpc.com/connect"
	metasearchpb "p9e.in/ugcl/metasearch/api/v1/metasearch"
	"p9e.in/ugcl/metasearch/api/v1/metasearch/metasearchconnect"
	"p9e.in/ugcl/metasearch/models"
	"p9e.in/ugcl/metasearch/services/interfaces"
)

// SearchHandler implements the metasearch SearchServiceHandler interface
type SearchHandler struct {
	queryService      interfaces.IQueryService
	suggestionService interfaces.ISuggestionService
	indexingService   interfaces.IIndexingService
	adminService      interfaces.IAdminService
	analyticsService  interfaces.IAnalyticsService
}

// NewSearchHandler creates a new search handler with the required services
func NewSearchHandler(
	queryService interfaces.IQueryService,
	suggestionService interfaces.ISuggestionService,
	indexingService interfaces.IIndexingService,
	adminService interfaces.IAdminService,
	analyticsService interfaces.IAnalyticsService,
) metasearchconnect.SearchServiceHandler {
	return &SearchHandler{
		queryService:      queryService,
		suggestionService: suggestionService,
		indexingService:   indexingService,
		adminService:      adminService,
		analyticsService:  analyticsService,
	}
}

// Query operations

func (h *SearchHandler) Search(ctx context.Context, req *connect.Request[metasearchpb.SearchRequest]) (*connect.Response[metasearchpb.SearchResponse], error) {
	// Convert proto request to service model
	searchReq := h.convertProtoToSearchRequest(req.Msg)

	// Call service
	result, err := h.queryService.Search(ctx, searchReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service model to proto response
	response := h.convertSearchResponseToProto(result)
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) MultiSearch(ctx context.Context, req *connect.Request[metasearchpb.MultiSearchRequest]) (*connect.Response[metasearchpb.MultiSearchResponse], error) {
	// Convert proto requests to service models
	var searchRequests []*models.SearchRequest
	for _, protoReq := range req.Msg.Requests {
		searchRequests = append(searchRequests, h.convertProtoToSearchRequest(protoReq))
	}

	// Call service
	results, err := h.queryService.MultiSearch(ctx, searchRequests)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service models to proto response
	var responses []*metasearchpb.SearchResponse
	for _, result := range results {
		responses = append(responses, h.convertSearchResponseToProto(result))
	}

	response := &metasearchpb.MultiSearchResponse{
		Responses: responses,
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) ScrollSearch(ctx context.Context, req *connect.Request[metasearchpb.ScrollSearchRequest]) (*connect.Response[metasearchpb.SearchResponse], error) {
	// Convert proto request to service model
	searchReq := h.convertProtoToSearchRequest(req.Msg.Request)

	// Call service
	result, err := h.queryService.ScrollSearch(ctx, searchReq, req.Msg.ScrollId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service model to proto response
	response := h.convertSearchResponseToProto(result)
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) Count(ctx context.Context, req *connect.Request[metasearchpb.CountRequest]) (*connect.Response[metasearchpb.CountResponse], error) {
	// Convert proto request to service model (simplified for count)
	searchReq := &models.SearchRequest{
		Query:     req.Msg.Query,
		Filters:   req.Msg.Filters,
		Indices:   req.Msg.Indices,
		TimeRange: h.convertProtoTimeRange(req.Msg.TimeRange),
	}

	// Call service
	count, err := h.queryService.Count(ctx, searchReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &metasearchpb.CountResponse{
		Count: count,
	}
	return connect.NewResponse(response), nil
}

// Suggestion operations

func (h *SearchHandler) GetSuggestions(ctx context.Context, req *connect.Request[metasearchpb.SuggestionRequest]) (*connect.Response[metasearchpb.SuggestionResponse], error) {
	// Convert proto request to service model
	suggestionReq := &models.SuggestionRequest{
		Query:   req.Msg.Query,
		Indices: req.Msg.Indices,
		Size:    int(req.Msg.Size),
	}

	// Call service
	result, err := h.suggestionService.GetSuggestions(ctx, suggestionReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service model to proto response
	response := &metasearchpb.SuggestionResponse{
		Suggestions: h.convertSuggestionsToProto(result.Suggestions),
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) GetPopularQueries(ctx context.Context, req *connect.Request[metasearchpb.PopularQueriesRequest]) (*connect.Response[metasearchpb.PopularQueriesResponse], error) {
	// Call service
	queries, err := h.suggestionService.GetPopularQueries(ctx, int(req.Msg.Limit))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service model to proto response
	var protoQueries []*metasearchpb.QueryFrequency
	for _, query := range queries {
		protoQueries = append(protoQueries, &metasearchpb.QueryFrequency{
			Query:     query.Query,
			Frequency: query.Frequency,
			Index:     query.Index,
		})
	}

	response := &metasearchpb.PopularQueriesResponse{
		Queries: protoQueries,
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) GetQueryHistory(ctx context.Context, req *connect.Request[metasearchpb.QueryHistoryRequest]) (*connect.Response[metasearchpb.QueryHistoryResponse], error) {
	// Call service
	queries, err := h.suggestionService.GetQueryHistory(ctx, req.Msg.UserId, int(req.Msg.Limit))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &metasearchpb.QueryHistoryResponse{
		Queries: queries,
	}
	return connect.NewResponse(response), nil
}

// Indexing operations

func (h *SearchHandler) IndexDocument(ctx context.Context, req *connect.Request[metasearchpb.IndexDocumentRequest]) (*connect.Response[metasearchpb.IndexDocumentResponse], error) {
	// Convert proto struct to map
	document := h.convertProtoStructToMap(req.Msg.Document)

	// Call service
	err := h.indexingService.IndexDocument(ctx, req.Msg.IndexName, req.Msg.DocId, document)
	if err != nil {
		return connect.NewResponse(&metasearchpb.IndexDocumentResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	response := &metasearchpb.IndexDocumentResponse{
		Success: true,
		Message: "Document indexed successfully",
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) IndexDocuments(ctx context.Context, req *connect.Request[metasearchpb.IndexDocumentsRequest]) (*connect.Response[metasearchpb.IndexDocumentsResponse], error) {
	// Convert proto structs to maps
	var documents []map[string]interface{}
	for _, doc := range req.Msg.Documents {
		documents = append(documents, h.convertProtoStructToMap(doc))
	}

	// Call service
	err := h.indexingService.IndexDocuments(ctx, req.Msg.IndexName, documents)
	if err != nil {
		return connect.NewResponse(&metasearchpb.IndexDocumentsResponse{
			Success:        false,
			Message:        err.Error(),
			ProcessedCount: 0,
			FailedCount:    int32(len(documents)),
		}), nil
	}

	response := &metasearchpb.IndexDocumentsResponse{
		Success:        true,
		Message:        "Documents indexed successfully",
		ProcessedCount: int32(len(documents)),
		FailedCount:    0,
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) UpdateDocument(ctx context.Context, req *connect.Request[metasearchpb.UpdateDocumentRequest]) (*connect.Response[metasearchpb.UpdateDocumentResponse], error) {
	// Convert proto struct to map
	updates := h.convertProtoStructToMap(req.Msg.Updates)

	// Call service
	err := h.indexingService.UpdateDocument(ctx, req.Msg.IndexName, req.Msg.DocId, updates)
	if err != nil {
		return connect.NewResponse(&metasearchpb.UpdateDocumentResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	response := &metasearchpb.UpdateDocumentResponse{
		Success: true,
		Message: "Document updated successfully",
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) DeleteDocument(ctx context.Context, req *connect.Request[metasearchpb.DeleteDocumentRequest]) (*connect.Response[metasearchpb.DeleteDocumentResponse], error) {
	// Call service
	err := h.indexingService.DeleteDocument(ctx, req.Msg.IndexName, req.Msg.DocId)
	if err != nil {
		return connect.NewResponse(&metasearchpb.DeleteDocumentResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	response := &metasearchpb.DeleteDocumentResponse{
		Success: true,
		Message: "Document deleted successfully",
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) BulkIndex(ctx context.Context, req *connect.Request[metasearchpb.BulkIndexRequest]) (*connect.Response[metasearchpb.BulkIndexResponse], error) {
	// Convert proto operations to service model
	var operations []interfaces.BulkOperation
	for _, op := range req.Msg.Operations {
		operation := interfaces.BulkOperation{
			Action: op.Action,
			Index:  op.Index,
			ID:     op.Id,
		}
		if op.Document != nil {
			operation.Document = h.convertProtoStructToMap(op.Document)
		}
		if op.Updates != nil {
			operation.Updates = h.convertProtoStructToMap(op.Updates)
		}
		operations = append(operations, operation)
	}

	// Call service
	err := h.indexingService.BulkIndex(ctx, operations)
	if err != nil {
		return connect.NewResponse(&metasearchpb.BulkIndexResponse{
			Success:        false,
			Message:        err.Error(),
			ProcessedCount: 0,
			FailedCount:    int32(len(operations)),
			Errors:         []string{err.Error()},
		}), nil
	}

	response := &metasearchpb.BulkIndexResponse{
		Success:        true,
		Message:        "Bulk operations completed successfully",
		ProcessedCount: int32(len(operations)),
		FailedCount:    0,
		Errors:         []string{},
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) RefreshIndex(ctx context.Context, req *connect.Request[metasearchpb.RefreshIndexRequest]) (*connect.Response[metasearchpb.RefreshIndexResponse], error) {
	// Call service
	err := h.indexingService.RefreshIndex(ctx, req.Msg.IndexName)
	if err != nil {
		return connect.NewResponse(&metasearchpb.RefreshIndexResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	response := &metasearchpb.RefreshIndexResponse{
		Success: true,
		Message: "Index refreshed successfully",
	}
	return connect.NewResponse(response), nil
}

// Admin operations

func (h *SearchHandler) CreateIndex(ctx context.Context, req *connect.Request[metasearchpb.CreateIndexRequest]) (*connect.Response[metasearchpb.CreateIndexResponse], error) {
	// Convert proto config to service model
	config := h.convertProtoToIndexConfig(req.Msg.Config)

	// Call service
	err := h.adminService.CreateIndex(ctx, config)
	if err != nil {
		return connect.NewResponse(&metasearchpb.CreateIndexResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	response := &metasearchpb.CreateIndexResponse{
		Success: true,
		Message: "Index created successfully",
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) DeleteIndex(ctx context.Context, req *connect.Request[metasearchpb.DeleteIndexRequest]) (*connect.Response[metasearchpb.DeleteIndexResponse], error) {
	// Call service
	err := h.adminService.DeleteIndex(ctx, req.Msg.IndexName)
	if err != nil {
		return connect.NewResponse(&metasearchpb.DeleteIndexResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	response := &metasearchpb.DeleteIndexResponse{
		Success: true,
		Message: "Index deleted successfully",
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) GetIndexStats(ctx context.Context, req *connect.Request[metasearchpb.GetIndexStatsRequest]) (*connect.Response[metasearchpb.GetIndexStatsResponse], error) {
	// Call service
	stats, err := h.adminService.GetIndexStats(ctx, req.Msg.IndexName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service model to proto response
	response := &metasearchpb.GetIndexStatsResponse{
		Stats: h.convertIndexStatsToProto(stats),
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) GetAllIndices(ctx context.Context, req *connect.Request[metasearchpb.GetAllIndicesRequest]) (*connect.Response[metasearchpb.GetAllIndicesResponse], error) {
	// Call service
	indices, err := h.adminService.GetAllIndices(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service models to proto response
	var protoIndices []*metasearchpb.IndexStats
	for _, index := range indices {
		protoIndices = append(protoIndices, h.convertIndexStatsToProto(index))
	}

	response := &metasearchpb.GetAllIndicesResponse{
		Indices: protoIndices,
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) ReindexData(ctx context.Context, req *connect.Request[metasearchpb.ReindexDataRequest]) (*connect.Response[metasearchpb.ReindexDataResponse], error) {
	// Call service
	job, err := h.adminService.ReindexData(ctx, req.Msg.SourceIndex, req.Msg.TargetIndex)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service model to proto response
	response := &metasearchpb.ReindexDataResponse{
		Job: h.convertIndexingJobToProto(job),
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) GetIndexingJobs(ctx context.Context, req *connect.Request[metasearchpb.GetIndexingJobsRequest]) (*connect.Response[metasearchpb.GetIndexingJobsResponse], error) {
	// Call service
	jobs, err := h.adminService.GetIndexingJobs(ctx, req.Msg.Status)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service models to proto response
	var protoJobs []*metasearchpb.IndexingJob
	for _, job := range jobs {
		protoJobs = append(protoJobs, h.convertIndexingJobToProto(job))
	}

	response := &metasearchpb.GetIndexingJobsResponse{
		Jobs: protoJobs,
	}
	return connect.NewResponse(response), nil
}

// Analytics operations

func (h *SearchHandler) RecordSearch(ctx context.Context, req *connect.Request[metasearchpb.RecordSearchRequest]) (*connect.Response[metasearchpb.RecordSearchResponse], error) {
	// Call service
	err := h.analyticsService.RecordSearch(ctx, req.Msg.Query, req.Msg.IndexName, req.Msg.ResultCount, req.Msg.ResponseTime)
	if err != nil {
		return connect.NewResponse(&metasearchpb.RecordSearchResponse{
			Success: false,
		}), nil
	}

	response := &metasearchpb.RecordSearchResponse{
		Success: true,
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) GetSearchAnalytics(ctx context.Context, req *connect.Request[metasearchpb.GetSearchAnalyticsRequest]) (*connect.Response[metasearchpb.GetSearchAnalyticsResponse], error) {
	// Convert proto time range to service model
	timeRange := h.convertProtoTimeRange(req.Msg.TimeRange)

	// Call service
	analytics, err := h.analyticsService.GetSearchAnalytics(ctx, timeRange)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service model to proto response
	response := &metasearchpb.GetSearchAnalyticsResponse{
		Analytics: h.convertSearchAnalyticsToProto(analytics),
	}
	return connect.NewResponse(response), nil
}

func (h *SearchHandler) GetSearchTrends(ctx context.Context, req *connect.Request[metasearchpb.GetSearchTrendsRequest]) (*connect.Response[metasearchpb.GetSearchTrendsResponse], error) {
	// Convert proto time range to service model
	timeRange := h.convertProtoTimeRange(req.Msg.TimeRange)

	// Call service
	trends, err := h.analyticsService.GetSearchTrends(ctx, timeRange)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &metasearchpb.GetSearchTrendsResponse{
		Trends: trends,
	}
	return connect.NewResponse(response), nil
}