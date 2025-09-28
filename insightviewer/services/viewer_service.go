package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	pb "p9e.in/ugcl/insightviewer/api/proto"
	"p9e.in/ugcl/insightviewer/mappers"
	"p9e.in/ugcl/insightviewer/models"
	"p9e.in/ugcl/insightviewer/repository"
)

// ViewerService implements the viewer service using clean architecture
type ViewerService struct {
	repoManager *repository.RepositoryManager
	protoMapper *mappers.ViewerMapper
}

// NewViewerService creates a new viewer service instance
func NewViewerService(repoManager *repository.RepositoryManager) *ViewerService {
	return &ViewerService{
		repoManager: repoManager,
		protoMapper: mappers.NewViewerMapper(),
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

	var parameters map[string]interface{}
	if req.Parameters != nil {
		parameters = req.Parameters.AsMap()
	}

	// Create a new report run
	run := &models.ReportRun{
		ID:         uuid.New(),
		ReportID:   reportID,
		RunBy:      req.RunBy,
		Status:     "running",
		Parameters: parameters,
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	createdRun, err := s.repoManager.Execution.CreateReportRun(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("failed to create report run: %w", err)
	}

	// TODO: Implement actual report execution logic
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

	// Simulate execution time
	time.Sleep(100 * time.Millisecond)

	// Complete the run
	durationMs := int32(100)
	completedRun, err := s.repoManager.Execution.CompleteReportRun(ctx, createdRun.ID, "completed", &durationMs, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete report run: %w", err)
	}

	// Store the results
	result := &models.ReportResult{
		ID:         uuid.New(),
		RunID:      completedRun.ID,
		ResultData: mockResults,
		RowCount:   2,
		CreatedAt:  time.Now(),
	}

	_, err = s.repoManager.Execution.CreateReportResult(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to store report result: %w", err)
	}

	return &pb.ExecuteReportResponse{
		RunId:  completedRun.ID.String(),
		Status: completedRun.Status,
	}, nil
}

func (s *ViewerService) GetReportResult(ctx context.Context, resultID uuid.UUID) (*pb.ReportResult, error) {
	result, err := s.repoManager.Execution.GetReportResultByID(ctx, resultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report result: %w", err)
	}

	return s.protoMapper.DomainReportResultToProto(result), nil
}

func (s *ViewerService) GetReportRuns(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*pb.ReportRun, int32, error) {
	runs, totalCount, err := s.repoManager.Execution.GetReportRuns(ctx, reportID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get report runs: %w", err)
	}

	var result []*pb.ReportRun
	for _, run := range runs {
		result = append(result, s.protoMapper.DomainReportRunToProto(run))
	}

	return result, totalCount, nil
}

// =============================================================================
// Report Caching
// =============================================================================

func (s *ViewerService) CacheReportResult(ctx context.Context, req *pb.CacheReportResultRequest) (*pb.ReportCache, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	var resultData map[string]interface{}
	if req.ResultData != nil {
		resultData = req.ResultData.AsMap()
	}

	cache := &models.ReportCache{
		ID:           uuid.New(),
		ReportID:     reportID,
		CacheKey:     req.CacheKey,
		ResultData:   resultData,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(req.TtlSeconds) * time.Second),
		HitCount:     0,
		LastAccessed: time.Now(),
	}

	createdCache, err := s.repoManager.Execution.CreateReportCache(ctx, cache)
	if err != nil {
		return nil, fmt.Errorf("failed to cache report result: %w", err)
	}

	return s.protoMapper.DomainReportCacheToProto(createdCache), nil
}

func (s *ViewerService) GetCachedResult(ctx context.Context, cacheKey string) (*pb.ReportCache, error) {
	cache, err := s.repoManager.Execution.GetCachedResult(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached result: %w", err)
	}

	// Update access time
	_ = s.repoManager.Execution.UpdateCacheAccess(ctx, cacheKey)

	return s.protoMapper.DomainReportCacheToProto(cache), nil
}

func (s *ViewerService) InvalidateCache(ctx context.Context, reportID uuid.UUID) error {
	err := s.repoManager.Execution.InvalidateReportCache(ctx, reportID)
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

	export := &models.ReportExport{
		ID:        uuid.New(),
		RunID:     runID,
		Format:    req.Format,
		Status:    "processing",
		CreatedBy: req.CreatedBy,
		CreatedAt: time.Now(),
	}

	createdExport, err := s.repoManager.Export.CreateReportExport(ctx, export)
	if err != nil {
		return nil, fmt.Errorf("failed to create report export: %w", err)
	}

	// TODO: Implement actual export processing
	// For now, simulate export completion
	filePath := fmt.Sprintf("/exports/report_%s.%s", createdExport.ID.String(), req.Format)
	fileSize := int64(1024) // 1KB mock size
	completedExport, err := s.repoManager.Export.CompleteReportExport(ctx, createdExport.ID, "completed", &filePath, &fileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to complete export: %w", err)
	}

	return s.protoMapper.DomainReportExportToProto(completedExport), nil
}

func (s *ViewerService) GetReportExport(ctx context.Context, exportID uuid.UUID) (*pb.ReportExport, error) {
	export, err := s.repoManager.Export.GetReportExportByID(ctx, exportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report export: %w", err)
	}

	return s.protoMapper.DomainReportExportToProto(export), nil
}

func (s *ViewerService) ListReportExports(ctx context.Context, reportID uuid.UUID, limit, offset int32) ([]*pb.ReportExport, int32, error) {
	exports, totalCount, err := s.repoManager.Export.ListReportExports(ctx, reportID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list report exports: %w", err)
	}

	var result []*pb.ReportExport
	for _, export := range exports {
		result = append(result, s.protoMapper.DomainReportExportToProto(export))
	}

	return result, totalCount, nil
}

// =============================================================================
// Report Scheduling
// =============================================================================

func (s *ViewerService) CreateReportSchedule(ctx context.Context, req *pb.CreateReportScheduleRequest) (*pb.ReportSchedule, error) {
	reportID, err := uuid.Parse(req.ReportId)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	var parameters map[string]interface{}
	if req.Parameters != nil {
		parameters = req.Parameters.AsMap()
	}

	schedule := &models.ReportSchedule{
		ID:             uuid.New(),
		ReportID:       reportID,
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		IsActive:       req.IsActive,
		Parameters:     parameters,
		CreatedBy:      req.CreatedBy,
		CreatedAt:      time.Now(),
		UpdatedBy:      req.CreatedBy,
		UpdatedAt:      time.Now(),
		// TODO: Calculate NextRunAt from cron expression
	}

	createdSchedule, err := s.repoManager.Scheduling.CreateReportSchedule(ctx, schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to create report schedule: %w", err)
	}

	return s.protoMapper.DomainReportScheduleToProto(createdSchedule), nil
}

// Add placeholder methods for other operations to keep the interface complete
func (s *ViewerService) UpdateReportSchedule(ctx context.Context, req *pb.UpdateReportScheduleRequest) (*pb.ReportSchedule, error) {
	// TODO: Implement using repository pattern
	return nil, fmt.Errorf("not implemented")
}

func (s *ViewerService) DeleteReportSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	return s.repoManager.Scheduling.DeleteReportSchedule(ctx, scheduleID)
}

func (s *ViewerService) GetReportSchedules(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportSchedule, error) {
	schedules, err := s.repoManager.Scheduling.GetReportSchedules(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report schedules: %w", err)
	}

	var result []*pb.ReportSchedule
	for _, schedule := range schedules {
		result = append(result, s.protoMapper.DomainReportScheduleToProto(schedule))
	}

	return result, nil
}

func (s *ViewerService) GetActiveSchedules(ctx context.Context) ([]*pb.ReportSchedule, error) {
	schedules, err := s.repoManager.Scheduling.GetActiveSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active schedules: %w", err)
	}

	var result []*pb.ReportSchedule
	for _, schedule := range schedules {
		result = append(result, s.protoMapper.DomainReportScheduleToProto(schedule))
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

	var alertRules map[string]interface{}
	if req.AlertRules != nil {
		alertRules = req.AlertRules.AsMap()
	}

	alert := &models.ReportAlert{
		ID:                   uuid.New(),
		ReportID:             reportID,
		Name:                 req.Name,
		Description:          req.Description,
		AlertType:            req.AlertType,
		AlertRules:           alertRules,
		IsActive:             req.IsActive,
		NotificationChannels: req.NotificationChannels,
		CreatedBy:            req.CreatedBy,
		CreatedAt:            time.Now(),
		UpdatedBy:            req.CreatedBy,
		UpdatedAt:            time.Now(),
		TriggerCount:         0,
	}

	createdAlert, err := s.repoManager.Scheduling.CreateReportAlert(ctx, alert)
	if err != nil {
		return nil, fmt.Errorf("failed to create report alert: %w", err)
	}

	return s.protoMapper.DomainReportAlertToProto(createdAlert), nil
}

// Add other methods as placeholders
func (s *ViewerService) UpdateReportAlert(ctx context.Context, req *pb.UpdateReportAlertRequest) (*pb.ReportAlert, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ViewerService) DeleteReportAlert(ctx context.Context, alertID uuid.UUID) error {
	return s.repoManager.Scheduling.DeleteReportAlert(ctx, alertID)
}

func (s *ViewerService) GetReportAlerts(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportAlert, error) {
	alerts, err := s.repoManager.Scheduling.GetReportAlerts(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report alerts: %w", err)
	}

	var result []*pb.ReportAlert
	for _, alert := range alerts {
		result = append(result, s.protoMapper.DomainReportAlertToProto(alert))
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

	var settings map[string]interface{}
	if req.Settings != nil {
		settings = req.Settings.AsMap()
	}

	subscription := &models.ReportSubscription{
		ID:               uuid.New(),
		ReportID:         reportID,
		UserID:           req.UserId,
		SubscriptionType: req.SubscriptionType,
		Frequency:        req.Frequency,
		DeliveryChannels: req.DeliveryChannels,
		IsActive:         req.IsActive,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        time.Now(),
		UpdatedBy:        req.CreatedBy,
		UpdatedAt:        time.Now(),
		Settings:         settings,
	}

	createdSubscription, err := s.repoManager.Subscription.CreateReportSubscription(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create report subscription: %w", err)
	}

	return s.protoMapper.DomainReportSubscriptionToProto(createdSubscription), nil
}

// Add other subscription methods as placeholders
func (s *ViewerService) UpdateReportSubscription(ctx context.Context, req *pb.UpdateReportSubscriptionRequest) (*pb.ReportSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ViewerService) DeleteReportSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	return s.repoManager.Subscription.DeleteReportSubscription(ctx, subscriptionID)
}

func (s *ViewerService) GetReportSubscriptions(ctx context.Context, reportID uuid.UUID) ([]*pb.ReportSubscription, error) {
	subscriptions, err := s.repoManager.Subscription.GetReportSubscriptions(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report subscriptions: %w", err)
	}

	var result []*pb.ReportSubscription
	for _, subscription := range subscriptions {
		result = append(result, s.protoMapper.DomainReportSubscriptionToProto(subscription))
	}

	return result, nil
}

func (s *ViewerService) GetUserSubscriptions(ctx context.Context, userID string) ([]*pb.ReportSubscription, error) {
	subscriptions, err := s.repoManager.Subscription.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	var result []*pb.ReportSubscription
	for _, subscription := range subscriptions {
		result = append(result, s.protoMapper.DomainReportSubscriptionToProto(subscription))
	}

	return result, nil
}