package mappers

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/insightviewer/api/proto"
	"p9e.in/ugcl/insightviewer/db/generated"
)

// ViewerMapper handles conversions between database models and protobuf models
type ViewerMapper struct{}

// NewViewerMapper creates a new viewer mapper instance
func NewViewerMapper() *ViewerMapper {
	return &ViewerMapper{}
}

// =============================================================================
// Report Run Mappings
// =============================================================================

func (m *ViewerMapper) ToProtoReportRun(run generated.ReportRun) *pb.ReportRun {
	return &pb.ReportRun{
		Id:         run.ID.String(),
		ReportId:   run.ReportID.String(),
		RunBy:      run.RunBy,
		RunAt:      timestamppb.New(run.RunAt),
		Status:     run.Status,
		DurationMs: run.DurationMs.Int32,
		ErrorMsg:   run.ErrorMsg.String,
	}
}

// =============================================================================
// Report Result Mappings
// =============================================================================

func (m *ViewerMapper) ToProtoReportResult(result generated.ReportResult) *pb.ReportResult {
	var resultData *structpb.Struct
	if result.ResultData != nil {
		var dataMap map[string]interface{}
		if err := json.Unmarshal(result.ResultData, &dataMap); err == nil {
			resultData, _ = structpb.NewStruct(dataMap)
		}
	}

	return &pb.ReportResult{
		Id:         result.ID.String(),
		RunId:      result.RunID.String(),
		ResultData: resultData,
		CreatedAt:  timestamppb.New(result.CreatedAt),
	}
}

// =============================================================================
// Report Cache Mappings
// =============================================================================

func (m *ViewerMapper) ToProtoReportCache(cache generated.ReportCache) *pb.ReportCache {
	var resultData *structpb.Struct
	if cache.ResultData != nil {
		var dataMap map[string]interface{}
		if err := json.Unmarshal(cache.ResultData, &dataMap); err == nil {
			resultData, _ = structpb.NewStruct(dataMap)
		}
	}

	return &pb.ReportCache{
		Id:         cache.ID.String(),
		ReportId:   cache.ReportID.String(),
		CacheKey:   cache.CacheKey,
		ResultData: resultData,
		CreatedAt:  timestamppb.New(cache.CreatedAt),
		ExpiresAt:  timestamppb.New(cache.ExpiresAt),
	}
}

// =============================================================================
// Report Export Mappings
// =============================================================================

func (m *ViewerMapper) ToProtoReportExport(export generated.ReportExport) *pb.ReportExport {
	return &pb.ReportExport{
		Id:          export.ID.String(),
		RunId:       export.RunID.String(),
		Format:      export.Format,
		Status:      export.Status,
		FilePath:    export.FilePath.String,
		FileSize:    export.FileSize.Int64,
		CreatedBy:   export.CreatedBy,
		CreatedAt:   timestamppb.New(export.CreatedAt),
		CompletedAt: timestamppb.New(export.CompletedAt.Time),
	}
}

// =============================================================================
// Report Schedule Mappings
// =============================================================================

func (m *ViewerMapper) ToProtoReportSchedule(schedule generated.ReportSchedule) *pb.ReportSchedule {
	var nextRunAt, lastRunAt *timestamppb.Timestamp
	if schedule.NextRunAt.Valid {
		nextRunAt = timestamppb.New(schedule.NextRunAt.Time)
	}
	if schedule.LastRunAt.Valid {
		lastRunAt = timestamppb.New(schedule.LastRunAt.Time)
	}

	return &pb.ReportSchedule{
		Id:             schedule.ID.String(),
		ReportId:       schedule.ReportID.String(),
		Name:           schedule.Name,
		CronExpression: schedule.CronExpression,
		IsActive:       schedule.IsActive,
		NextRunAt:      nextRunAt,
		LastRunAt:      lastRunAt,
		CreatedBy:      schedule.CreatedBy,
		CreatedAt:      timestamppb.New(schedule.CreatedAt),
		UpdatedBy:      schedule.UpdatedBy.String,
		UpdatedAt:      timestamppb.New(schedule.UpdatedAt),
	}
}

// =============================================================================
// Report Alert Mappings
// =============================================================================

func (m *ViewerMapper) ToProtoReportAlert(alert generated.ReportAlert) *pb.ReportAlert {
	var alertRules *structpb.Struct
	if alert.AlertRules != nil {
		var rulesMap map[string]interface{}
		if err := json.Unmarshal(alert.AlertRules, &rulesMap); err == nil {
			alertRules, _ = structpb.NewStruct(rulesMap)
		}
	}

	return &pb.ReportAlert{
		Id:                   alert.ID.String(),
		ReportId:             alert.ReportID.String(),
		Name:                 alert.Name,
		AlertType:            alert.AlertType,
		AlertRules:           alertRules,
		IsActive:             alert.IsActive,
		NotificationChannels: alert.NotificationChannels,
		CreatedBy:            alert.CreatedBy,
		CreatedAt:            timestamppb.New(alert.CreatedAt),
		UpdatedBy:            alert.UpdatedBy.String,
		UpdatedAt:            timestamppb.New(alert.UpdatedAt),
	}
}

// =============================================================================
// Report Subscription Mappings
// =============================================================================

func (m *ViewerMapper) ToProtoReportSubscription(subscription generated.ReportSubscription) *pb.ReportSubscription {
	return &pb.ReportSubscription{
		Id:               subscription.ID.String(),
		ReportId:         subscription.ReportID.String(),
		UserId:           subscription.UserID,
		SubscriptionType: subscription.SubscriptionType,
		Frequency:        subscription.Frequency,
		IsActive:         subscription.IsActive,
		DeliveryChannels: subscription.DeliveryChannels,
		CreatedBy:        subscription.CreatedBy,
		CreatedAt:        timestamppb.New(subscription.CreatedAt),
		UpdatedBy:        subscription.UpdatedBy.String,
		UpdatedAt:        timestamppb.New(subscription.UpdatedAt),
	}
}