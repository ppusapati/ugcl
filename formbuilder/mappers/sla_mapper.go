package mappers

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	proto "p9e.in/ugcl/formbuilder/api/v2/workflow"
	db "p9e.in/ugcl/formbuilder/db/generated"
)

// SLA Instance mappers
func CreateSLAInstanceParams(instanceID, slaRuleID uuid.UUID, state string, startTime, dueTime time.Time, assignedTo *string) db.CreateSLAInstanceParams {
	return db.CreateSLAInstanceParams{
		ID:         uuid.New(),
		InstanceID: instanceID,
		SlaRuleID:  slaRuleID,
		State:      state,
		StartTime:  startTime,
		DueTime:    dueTime,
		Status:     "active",
		AssignedTo: assignedTo,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func SLAInstanceToProto(model *db.SlaInstance) *proto.SLAInstanceInfo {
	if model == nil {
		return nil
	}

	proto := &proto.SLAInstanceInfo{
		Id:         model.ID.String(),
		InstanceId: model.InstanceID.String(),
		SlaRuleId:  model.SlaRuleID.String(),
		State:      model.State,
		StartTime:  timestamppb.New(model.StartTime),
		DueTime:    timestamppb.New(model.DueTime),
		Status:     model.Status,
		CreatedAt:  timestamppb.New(model.CreatedAt),
		UpdatedAt:  timestamppb.New(model.UpdatedAt),
	}

	if model.CompletionTime.Valid {
		proto.CompletionTime = timestamppb.New(model.CompletionTime.Time)
	}

	if model.BreachTime.Valid {
		proto.BreachTime = timestamppb.New(model.BreachTime.Time)
	}

	if model.AssignedTo != nil {
		proto.AssignedTo = *model.AssignedTo
	}

	return proto
}

// SLA Violation mappers
func CreateSLAViolationParams(slaInstanceID, instanceID, slaRuleID uuid.UUID, violationTime time.Time, breachDuration string, severity string) db.CreateSLAViolationParams {
	return db.CreateSLAViolationParams{
		ID:            uuid.New(),
		SlaInstanceID: slaInstanceID,
		InstanceID:    instanceID,
		SlaRuleID:     slaRuleID,
		ViolationTime: violationTime,
		// Note: BreachDuration needs proper interval handling
		Severity:  severity,
		CreatedAt: time.Now(),
	}
}

func SLAViolationToProto(model *db.SlaViolation) *proto.SLAViolationInfo {
	if model == nil {
		return nil
	}

	proto := &proto.SLAViolationInfo{
		Id:            model.ID.String(),
		SlaInstanceId: model.SlaInstanceID.String(),
		InstanceId:    model.InstanceID.String(),
		SlaRuleId:     model.SlaRuleID.String(),
		ViolationTime: timestamppb.New(model.ViolationTime),
		Severity:      model.Severity,
		CreatedAt:     timestamppb.New(model.CreatedAt),
	}

	if model.Resolved != nil {
		proto.Resolved = *model.Resolved
	}

	if model.ResolvedAt.Valid {
		proto.ResolvedAt = timestamppb.New(model.ResolvedAt.Time)
	}

	if model.ResolvedBy != nil {
		proto.ResolvedBy = *model.ResolvedBy
	}

	if model.ResolutionNotes != nil {
		proto.ResolutionNotes = *model.ResolutionNotes
	}

	return proto
}

// SLA Escalation Instance mappers
func CreateSLAEscalationInstanceParams(slaInstanceID uuid.UUID, escalationLevel int32, triggeredAt time.Time, actionsExecuted []byte) db.CreateSLAEscalationInstanceParams {
	return db.CreateSLAEscalationInstanceParams{
		ID:              uuid.New(),
		SlaInstanceID:   slaInstanceID,
		EscalationLevel: escalationLevel,
		TriggeredAt:     triggeredAt,
		ActionsExecuted: actionsExecuted,
		Status:          "pending",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func SLAEscalationInstanceToProto(model *db.SlaEscalationInstance) *proto.SLAEscalationInfo {
	if model == nil {
		return nil
	}

	prt := &proto.SLAEscalationInfo{
		Id:              model.ID.String(),
		SlaInstanceId:   model.SlaInstanceID.String(),
		EscalationLevel: model.EscalationLevel,
		TriggeredAt:     timestamppb.New(model.TriggeredAt),
		Status:          model.Status,
		CreatedAt:       timestamppb.New(model.CreatedAt),
		UpdatedAt:       timestamppb.New(model.UpdatedAt),
	}

	if model.ErrorMessage != nil {
		prt.ErrorMessage = *model.ErrorMessage
	}

	if model.RetryCount != nil {
		prt.RetryCount = *model.RetryCount
	}

	if model.NextRetryAt.Valid {
		prt.NextRetryAt = timestamppb.New(model.NextRetryAt.Time)
	}

	// Unmarshal actions executed
	if len(model.ActionsExecuted) > 0 {
		var actions []*proto.EscalationAction
		if err := json.Unmarshal(model.ActionsExecuted, &actions); err == nil {
			prt.ActionsExecuted = actions
		}
	}

	return prt
}

// SLA Notification mappers
func CreateSLANotificationParams(slaInstanceID *uuid.UUID, slaViolationID *uuid.UUID, escalationInstanceID *uuid.UUID, notificationType, recipient, channel, message string) db.CreateSLANotificationParams {
	params := db.CreateSLANotificationParams{
		ID:               uuid.New(),
		NotificationType: notificationType,
		Recipient:        recipient,
		Channel:          channel,
		Message:          message,
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	if slaInstanceID != nil {
		params.SlaInstanceID = *slaInstanceID
	}

	if slaViolationID != nil {
		params.SlaViolationID = uuid.NullUUID{UUID: *slaViolationID, Valid: true}
	}

	if escalationInstanceID != nil {
		params.EscalationInstanceID = uuid.NullUUID{UUID: *escalationInstanceID, Valid: true}
	}

	return params
}

func SLANotificationToProto(model *db.SlaNotification) *proto.SLANotificationInfo {
	if model == nil {
		return nil
	}

	proto := &proto.SLANotificationInfo{
		Id:               model.ID.String(),
		SlaInstanceId:    model.SlaInstanceID.String(),
		NotificationType: model.NotificationType,
		Recipient:        model.Recipient,
		Channel:          model.Channel,
		Message:          model.Message,
		Status:           model.Status,
		CreatedAt:        timestamppb.New(model.CreatedAt),
	}

	if model.SlaViolationID.Valid {
		proto.SlaViolationId = model.SlaViolationID.UUID.String()
	}

	if model.EscalationInstanceID.Valid {
		proto.EscalationInstanceId = model.EscalationInstanceID.UUID.String()
	}

	if model.Subject != nil {
		proto.Subject = *model.Subject
	}

	if model.TemplateUsed != nil {
		proto.TemplateUsed = *model.TemplateUsed
	}

	if model.SentAt.Valid {
		proto.SentAt = timestamppb.New(model.SentAt.Time)
	}

	if model.DeliveredAt.Valid {
		proto.DeliveredAt = timestamppb.New(model.DeliveredAt.Time)
	}

	if model.ErrorMessage != nil {
		proto.ErrorMessage = *model.ErrorMessage
	}

	if model.RetryCount != nil {
		proto.RetryCount = *model.RetryCount
	}

	return proto
}

// SLA Metrics mappers
func CreateSLAMetricParams(metricDate time.Time, slaRuleID uuid.UUID, state, assignedRole *string, totalInstances, completedOnTime, breachedInstances int32, compliancePercentage float64) db.CreateSLAMetricParams {
	return db.CreateSLAMetricParams{
		ID:                uuid.New(),
		MetricDate:        metricDate,
		SlaRuleID:         slaRuleID,
		State:             state,
		AssignedRole:      assignedRole,
		TotalInstances:    totalInstances,
		CompletedOnTime:   completedOnTime,
		BreachedInstances: breachedInstances,
		// CompliancePercentage: needs proper numeric handling
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func SLAMetricToProto(model *db.SlaMetric) *proto.SLAMetricInfo {
	if model == nil {
		return nil
	}

	proto := &proto.SLAMetricInfo{
		Id:                model.ID.String(),
		MetricDate:        timestamppb.New(model.MetricDate),
		SlaRuleId:         model.SlaRuleID.String(),
		TotalInstances:    model.TotalInstances,
		CompletedOnTime:   model.CompletedOnTime,
		BreachedInstances: model.BreachedInstances,
		CreatedAt:         timestamppb.New(model.CreatedAt),
		UpdatedAt:         timestamppb.New(model.UpdatedAt),
	}

	if model.State != nil {
		proto.State = *model.State
	}

	if model.AssignedRole != nil {
		proto.AssignedRole = *model.AssignedRole
	}

	// Handle compliance percentage conversion from pgtype.Numeric
	// This would need proper numeric conversion based on your requirements

	return proto
}

// SLA Pause Log mappers
func CreateSLAPauseLogParams(slaInstanceID uuid.UUID, action, reason string, pausedBy string, pausedAt time.Time) db.CreateSLAPauseLogParams {
	return db.CreateSLAPauseLogParams{
		ID:            uuid.New(),
		SlaInstanceID: slaInstanceID,
		Action:        action,
		Reason:        stringPtr(reason),
		PausedBy:      pausedBy,
		PausedAt:      pausedAt,
		CreatedAt:     time.Now(),
	}
}

func SLAPauseLogToProto(model *db.SlaPauseLog) *proto.SLAPauseLogInfo {
	if model == nil {
		return nil
	}

	proto := &proto.SLAPauseLogInfo{
		Id:            model.ID.String(),
		SlaInstanceId: model.SlaInstanceID.String(),
		Action:        model.Action,
		PausedBy:      model.PausedBy,
		PausedAt:      timestamppb.New(model.PausedAt),
		CreatedAt:     timestamppb.New(model.CreatedAt),
	}

	if model.Reason != nil {
		proto.Reason = *model.Reason
	}

	if model.ResumedAt.Valid {
		proto.ResumedAt = timestamppb.New(model.ResumedAt.Time)
	}

	return proto
}

// Dashboard and reporting mappers
func SLADashboardDataToProto(models []db.GetSLADashboardDataRow) []*proto.SLADashboardInfo {
	var result []*proto.SLADashboardInfo

	for _, model := range models {
		proto := &proto.SLADashboardInfo{
			SlaRuleId:          model.SlaRuleID.String(),
			SlaRuleName:        model.SlaRuleName,
			State:              model.State,
			TotalInstances:     int32(model.TotalInstances),
			ActiveInstances:    int32(model.ActiveInstances),
			CompletedInstances: int32(model.CompletedInstances),
			BreachedInstances:  int32(model.BreachedInstances),
			OverdueInstances:   int32(model.OverdueInstances),
			// CompliancePercentage: needs proper numeric conversion
		}
		result = append(result, proto)
	}

	return result
}

func SLAInstancesWithDetailsToProto(models []db.GetSLAInstancesWithDetailsRow) []*proto.SLAInstanceDetailInfo {
	var result []*proto.SLAInstanceDetailInfo

	for _, model := range models {
		proto := &proto.SLAInstanceDetailInfo{
			Id:                model.ID.String(),
			InstanceId:        model.InstanceID.String(),
			SlaRuleId:         model.SlaRuleID.String(),
			State:             model.State,
			StartTime:         timestamppb.New(model.StartTime),
			DueTime:           timestamppb.New(model.DueTime),
			Status:            model.Status,
			FormId:            model.FormID.String(),
			InstanceCreatedBy: model.InstanceCreatedBy,
			SlaRuleName:       model.SlaRuleName,
			HoursRemaining:    float32(model.HoursRemaining),
			IsOverdue:         model.IsOverdue,
			CreatedAt:         timestamppb.New(model.CreatedAt),
			UpdatedAt:         timestamppb.New(model.UpdatedAt),
		}

		if model.CompletionTime.Valid {
			proto.CompletionTime = timestamppb.New(model.CompletionTime.Time)
		}

		if model.BreachTime.Valid {
			proto.BreachTime = timestamppb.New(model.BreachTime.Time)
		}

		if model.AssignedTo != nil {
			proto.AssignedTo = *model.AssignedTo
		}

		result = append(result, proto)
	}

	return result
}
