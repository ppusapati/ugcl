package mappers

import (
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/insightviewer/api/proto"
	"p9e.in/ugcl/insightviewer/models"
)

// DomainReportRunToProto converts domain model to proto
func (m *ViewerMapper) DomainReportRunToProto(run *models.ReportRun) *pb.ReportRun {
	protoRun := &pb.ReportRun{
		Id:       run.ID.String(),
		ReportId: run.ReportID.String(),
		RunBy:    run.RunBy,
		RunAt:    timestamppb.New(run.StartedAt),
		Status:   run.Status,
	}

	if run.DurationMs != nil {
		protoRun.DurationMs = *run.DurationMs
	}

	if run.ErrorMsg != nil {
		protoRun.ErrorMsg = *run.ErrorMsg
	}

	return protoRun
}

// DomainReportResultToProto converts domain model to proto
func (m *ViewerMapper) DomainReportResultToProto(result *models.ReportResult) *pb.ReportResult {
	var resultData *structpb.Struct
	if result.ResultData != nil {
		resultData, _ = structpb.NewStruct(result.ResultData)
	}

	return &pb.ReportResult{
		Id:         result.ID.String(),
		RunId:      result.RunID.String(),
		ResultData: resultData,
		CreatedAt:  timestamppb.New(result.CreatedAt),
	}
}

// DomainReportCacheToProto converts domain model to proto
func (m *ViewerMapper) DomainReportCacheToProto(cache *models.ReportCache) *pb.ReportCache {
	var resultData *structpb.Struct
	if cache.ResultData != nil {
		resultData, _ = structpb.NewStruct(cache.ResultData)
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

// DomainReportExportToProto converts domain model to proto
func (m *ViewerMapper) DomainReportExportToProto(export *models.ReportExport) *pb.ReportExport {
	protoExport := &pb.ReportExport{
		Id:        export.ID.String(),
		RunId:     export.RunID.String(),
		Format:    export.Format,
		Status:    export.Status,
		CreatedBy: export.CreatedBy,
		CreatedAt: timestamppb.New(export.CreatedAt),
	}

	if export.FilePath != nil {
		protoExport.FilePath = *export.FilePath
	}

	if export.FileSize != nil {
		protoExport.FileSize = *export.FileSize
	}

	if export.CompletedAt != nil {
		protoExport.CompletedAt = timestamppb.New(*export.CompletedAt)
	}

	return protoExport
}

// DomainReportScheduleToProto converts domain model to proto
func (m *ViewerMapper) DomainReportScheduleToProto(schedule *models.ReportSchedule) *pb.ReportSchedule {
	protoSchedule := &pb.ReportSchedule{
		Id:             schedule.ID.String(),
		ReportId:       schedule.ReportID.String(),
		Name:           schedule.Name,
		CronExpression: schedule.CronExpression,
		IsActive:       schedule.IsActive,
		CreatedBy:      schedule.CreatedBy,
		CreatedAt:      timestamppb.New(schedule.CreatedAt),
		UpdatedBy:      schedule.UpdatedBy,
		UpdatedAt:      timestamppb.New(schedule.UpdatedAt),
	}

	if schedule.NextRunAt != nil {
		protoSchedule.NextRunAt = timestamppb.New(*schedule.NextRunAt)
	}

	if schedule.LastRunAt != nil {
		protoSchedule.LastRunAt = timestamppb.New(*schedule.LastRunAt)
	}

	return protoSchedule
}

// DomainReportAlertToProto converts domain model to proto
func (m *ViewerMapper) DomainReportAlertToProto(alert *models.ReportAlert) *pb.ReportAlert {
	var alertRules *structpb.Struct
	if alert.AlertRules != nil {
		alertRules, _ = structpb.NewStruct(alert.AlertRules)
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
		UpdatedBy:            alert.UpdatedBy,
		UpdatedAt:            timestamppb.New(alert.UpdatedAt),
	}
}

// DomainReportSubscriptionToProto converts domain model to proto
func (m *ViewerMapper) DomainReportSubscriptionToProto(subscription *models.ReportSubscription) *pb.ReportSubscription {
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
		UpdatedBy:        subscription.UpdatedBy,
		UpdatedAt:        timestamppb.New(subscription.UpdatedAt),
	}
}