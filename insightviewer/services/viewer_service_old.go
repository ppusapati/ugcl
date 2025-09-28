package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/insightviewer/api/proto"
	"p9e.in/ugcl/insightviewer/db/generated"
	"p9e.in/ugcl/insightviewer/mappers"
)

// ViewerService implements the IViewerService interface
type ViewerService struct {
	queries *generated.Queries
	mapper  *mappers.ViewerMapper
}

// NewViewerService creates a new viewer service instance
func NewViewerService(db *sql.DB) *ViewerService {
	return &ViewerService{
		queries: generated.New(db),
		mapper:  mappers.NewViewerMapper(),
	}
}

// =============================================================================
// Report Execution
// =============================================================================

func (s *ViewerService) ExecuteReport(ctx context.Context, req *pb.ExecuteReportRequest) (*pb.ExecuteReportResponse, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	// Create a new report run
	run, err := s.queries.CreateReportRun(ctx, generated.CreateReportRunParams{
		ReportID:   reportID,
		RunBy:      req.RunBy,
		Status:     "running",
		Parameters: []byte("{}"), // TODO: Handle parameters properly
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create report run: %w", err)
	}

	// TODO: Implement actual report execution logic
	// This would involve:
	// 1. Loading report definition from InsightHub
	// 2. Building SQL query from report definition
	// 3. Executing query against data sources
	// 4. Processing and formatting results
	// 5. Storing results in report_results table

	// For now, simulate execution with mock data
	mockResults := map[string]interface{}{
		"columns": []map[string]interface{}{
			{"name": "id", "type": "integer"},
			{"name": "name", "type": "string"},
			{"name": "value", "type": "decimal"},
		},
		"rows": []map[string]interface{}{
			{"id": 1, "name": "Sample 1", "value": 100.50},
			{"id": 2, "name": "Sample 2", "value": 250.75},
		},
	}

	resultsJSON, _ := json.Marshal(mockResults)

	// Update run status and store results
	completedRun, err := s.queries.CompleteReportRun(ctx, generated.CompleteReportRunParams{
		ID:         run.ID,
		Status:     "completed",
		DurationMs: sql.NullInt32{Int32: 1500, Valid: true}, // 1.5 seconds
		ErrorMsg:   sql.NullString{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to complete report run: %w", err)
	}

	// Store the results
	_, err = s.queries.CreateReportResult(ctx, generated.CreateReportResultParams{
		RunID:      run.ID,
		ResultData: resultsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store report result: %w", err)
	}

	return &pb.ExecuteReportResponse{
		RunId:  completedRun.ID.String(),
		Status: completedRun.Status,
	}, nil
}

func (s *ViewerService) GetReportResult(ctx context.Context, resultID uuid.UUID) (*pb.ReportResult, error) {
	result, err := s.queries.GetReportResult(ctx, resultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report result: %w", err)
	}

	return s.mapper.ToProtoReportResult(result), nil
}

func (s *ViewerService) GetReportRuns(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*pb.ReportRun, int32, error) {
	runs, err := s.queries.GetReportRuns(ctx, generated.GetReportRunsParams{
		ReportID: reportID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get report runs: %w", err)
	}

	var result []*pb.ReportRun
	for _, run := range runs {
		result = append(result, s.mapper.ToProtoReportRun(run))
	}

	// Get total count
	totalCount, err := s.queries.CountReportRuns(ctx, reportID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count report runs: %w", err)
	}

	return result, int32(totalCount), nil
}

// =============================================================================
// Report Caching
// =============================================================================

func (s *ViewerService) CacheReportResult(ctx context.Context, req *pb.CacheReportResultRequest) (*pb.ReportCache, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	var resultData []byte
	if req.ResultData != nil {
		resultData, _ = json.Marshal(req.ResultData.AsMap())
	}

	cache, err := s.queries.CreateReportCache(ctx, generated.CreateReportCacheParams{
		ReportID:   reportID,
		CacheKey:   req.CacheKey,
		ResultData: resultData,
		ExpiresAt:  time.Now().Add(time.Duration(req.TtlSeconds) * time.Second),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cache report result: %w", err)
	}

	return s.mapper.ToProtoReportCache(cache), nil
}

func (s *ViewerService) GetCachedResult(ctx context.Context, cacheKey string) (*pb.ReportCache, error) {
	cache, err := s.queries.GetCachedResult(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached result: %w", err)
	}

	return s.mapper.ToProtoReportCache(cache), nil
}

func (s *ViewerService) InvalidateCache(ctx context.Context, reportID uuid.UUID) error {
	err := s.queries.InvalidateReportCache(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	return nil
}

// =============================================================================
// Report Exports
// =============================================================================

func (s *ViewerService) ExportReport(ctx context.Context, req *pb.ExportReportRequest) (*pb.ReportExport, error) {
	runID, err := uuid.Parse(req.RunId)
	if err != nil {
		return nil, fmt.Errorf("invalid run ID: %w", err)
	}

	export, err := s.queries.CreateReportExport(ctx, generated.CreateReportExportParams{
		RunID:     runID,
		Format:    req.Format,
		Status:    "processing",
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create report export: %w", err)
	}

	// TODO: Implement actual export processing
	// This would involve:
	// 1. Getting the report results
	// 2. Converting to requested format (CSV, Excel, PDF)
	// 3. Storing file and updating export record

	// For now, simulate export completion
	filePath := fmt.Sprintf("/exports/report_%s.%s", export.ID.String(), req.Format)
	completedExport, err := s.queries.CompleteReportExport(ctx, generated.CompleteReportExportParams{
		ID:       export.ID,
		Status:   "completed",
		FilePath: sql.NullString{String: filePath, Valid: true},
		FileSize: sql.NullInt64{Int64: 1024, Valid: true}, // 1KB mock size
	})
	if err != nil {
		return nil, fmt.Errorf("failed to complete export: %w", err)
	}

	return s.mapper.ToProtoReportExport(completedExport), nil
}

func (s *ViewerService) GetReportExport(ctx context.Context, exportID uuid.UUID) (*pb.ReportExport, error) {
	export, err := s.queries.GetReportExport(ctx, exportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report export: %w", err)
	}

	return s.mapper.ToProtoReportExport(export), nil
}

func (s *ViewerService) ListReportExports(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*pb.ReportExport, int32, error) {
	exports, err := s.queries.ListReportExports(ctx, generated.ListReportExportsParams{
		ReportID: reportID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list report exports: %w", err)
	}

	var result []*pb.ReportExport
	for _, export := range exports {
		result = append(result, s.mapper.ToProtoReportExport(export))
	}

	// Get total count
	totalCount, err := s.queries.CountReportExports(ctx, reportID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count report exports: %w", err)
	}

	return result, int32(totalCount), nil
}

// =============================================================================
// Report Scheduling
// =============================================================================

func (s *ViewerService) CreateReportSchedule(ctx context.Context, req *pb.CreateReportScheduleRequest) (*pb.ReportSchedule, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	schedule, err := s.queries.CreateReportSchedule(ctx, generated.CreateReportScheduleParams{
		ReportID:       reportID,
		Name:           req.Name,
		CronExpression: req.CronExpression,
		IsActive:       req.IsActive,
		CreatedBy:      req.CreatedBy,
		NextRunAt:      sql.NullTime{}, // TODO: Calculate from cron expression
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create report schedule: %w", err)
	}

	return s.mapper.ToProtoReportSchedule(schedule), nil
}

func (s *ViewerService) UpdateReportSchedule(ctx context.Context, req *pb.UpdateReportScheduleRequest) (*pb.ReportSchedule, error) {
	scheduleID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID: %w", err)
	}

	schedule, err := s.queries.UpdateReportSchedule(ctx, generated.UpdateReportScheduleParams{
		ID:             scheduleID,
		Name:           req.Name,
		CronExpression: req.CronExpression,
		IsActive:       req.IsActive,
		UpdatedBy:      sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report schedule: %w", err)
	}

	return s.mapper.ToProtoReportSchedule(schedule), nil
}

func (s *ViewerService) DeleteReportSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	err := s.queries.DeleteReportSchedule(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete report schedule: %w", err)
	}

	return nil
}

func (s *ViewerService) GetReportSchedules(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportSchedule, error) {
	schedules, err := s.queries.GetReportSchedules(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report schedules: %w", err)
	}

	var result []*pb.ReportSchedule
	for _, schedule := range schedules {
		result = append(result, s.mapper.ToProtoReportSchedule(schedule))
	}

	return result, nil
}

func (s *ViewerService) GetActiveSchedules(ctx context.Context) ([]*pb.ReportSchedule, error) {
	schedules, err := s.queries.GetActiveSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active schedules: %w", err)
	}

	var result []*pb.ReportSchedule
	for _, schedule := range schedules {
		result = append(result, s.mapper.ToProtoReportSchedule(schedule))
	}

	return result, nil
}

// =============================================================================
// Report Alerts
// =============================================================================

func (s *ViewerService) CreateReportAlert(ctx context.Context, req *pb.CreateReportAlertRequest) (*pb.ReportAlert, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	var alertRules []byte
	if req.AlertRules != nil {
		alertRules, _ = json.Marshal(req.AlertRules.AsMap())
	}

	alert, err := s.queries.CreateReportAlert(ctx, generated.CreateReportAlertParams{
		ReportID:       reportID,
		Name:           req.Name,
		AlertType:      req.AlertType,
		AlertRules:     alertRules,
		IsActive:       req.IsActive,
		CreatedBy:      req.CreatedBy,
		NotificationChannels: req.NotificationChannels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create report alert: %w", err)
	}

	return s.mapper.ToProtoReportAlert(alert), nil
}

func (s *ViewerService) UpdateReportAlert(ctx context.Context, req *pb.UpdateReportAlertRequest) (*pb.ReportAlert, error) {
	alertID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid alert ID: %w", err)
	}

	var alertRules []byte
	if req.AlertRules != nil {
		alertRules, _ = json.Marshal(req.AlertRules.AsMap())
	}

	alert, err := s.queries.UpdateReportAlert(ctx, generated.UpdateReportAlertParams{
		ID:                   alertID,
		Name:                 req.Name,
		AlertType:            req.AlertType,
		AlertRules:           alertRules,
		IsActive:             req.IsActive,
		UpdatedBy:            sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
		NotificationChannels: req.NotificationChannels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report alert: %w", err)
	}

	return s.mapper.ToProtoReportAlert(alert), nil
}

func (s *ViewerService) DeleteReportAlert(ctx context.Context, alertID uuid.UUID) error {
	err := s.queries.DeleteReportAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("failed to delete report alert: %w", err)
	}

	return nil
}

func (s *ViewerService) GetReportAlerts(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportAlert, error) {
	alerts, err := s.queries.GetReportAlerts(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report alerts: %w", err)
	}

	var result []*pb.ReportAlert
	for _, alert := range alerts {
		result = append(result, s.mapper.ToProtoReportAlert(alert))
	}

	return result, nil
}

// =============================================================================
// Report Subscriptions
// =============================================================================

func (s *ViewerService) CreateReportSubscription(ctx context.Context, req *pb.CreateReportSubscriptionRequest) (*pb.ReportSubscription, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	subscription, err := s.queries.CreateReportSubscription(ctx, generated.CreateReportSubscriptionParams{
		ReportID:         reportID,
		UserID:           req.UserId,
		SubscriptionType: req.SubscriptionType,
		Frequency:        req.Frequency,
		IsActive:         req.IsActive,
		DeliveryChannels: req.DeliveryChannels,
		CreatedBy:        req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create report subscription: %w", err)
	}

	return s.mapper.ToProtoReportSubscription(subscription), nil
}

func (s *ViewerService) UpdateReportSubscription(ctx context.Context, req *pb.UpdateReportSubscriptionRequest) (*pb.ReportSubscription, error) {
	subscriptionID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription ID: %w", err)
	}

	subscription, err := s.queries.UpdateReportSubscription(ctx, generated.UpdateReportSubscriptionParams{
		ID:               subscriptionID,
		SubscriptionType: req.SubscriptionType,
		Frequency:        req.Frequency,
		IsActive:         req.IsActive,
		DeliveryChannels: req.DeliveryChannels,
		UpdatedBy:        sql.NullString{String: req.UpdatedBy, Valid: req.UpdatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update report subscription: %w", err)
	}

	return s.mapper.ToProtoReportSubscription(subscription), nil
}

func (s *ViewerService) DeleteReportSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	err := s.queries.DeleteReportSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to delete report subscription: %w", err)
	}

	return nil
}

func (s *ViewerService) GetReportSubscriptions(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportSubscription, error) {
	subscriptions, err := s.queries.GetReportSubscriptions(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report subscriptions: %w", err)
	}

	var result []*pb.ReportSubscription
	for _, subscription := range subscriptions {
		result = append(result, s.mapper.ToProtoReportSubscription(subscription))
	}

	return result, nil
}

func (s *ViewerService) GetUserSubscriptions(ctx context.Context, userID string) ([]*pb.ReportSubscription, error) {
	subscriptions, err := s.queries.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	var result []*pb.ReportSubscription
	for _, subscription := range subscriptions {
		result = append(result, s.mapper.ToProtoReportSubscription(subscription))
	}

	return result, nil
}