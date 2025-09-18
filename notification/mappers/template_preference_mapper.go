package mappers

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	"p9e.in/ugcl/notification/api/v2/notification"
	db "p9e.in/ugcl/notification/db/generated"
)

// TemplatePreferenceMapper handles conversions for templates and preferences
type TemplatePreferenceMapper struct {
	notificationMapper *NotificationMapper
}

// NewTemplatePreferenceMapper creates a new template preference mapper
func NewTemplatePreferenceMapper() *TemplatePreferenceMapper {
	return &TemplatePreferenceMapper{
		notificationMapper: NewNotificationMapper(),
	}
}

// Template Mappers

// ProtoToSQLCTemplate converts proto NotificationTemplate to SQLC NotificationTemplate
func (m *TemplatePreferenceMapper) ProtoToSQLCTemplate(proto *notification.NotificationTemplate) (*db.NotificationTemplate, error) {
	if proto == nil {
		return nil, nil
	}

	sqlcTemplate := &db.NotificationTemplate{
		Name:            proto.Name,
		Type:            m.notificationMapper.ProtoToSQLCNotificationType(proto.Type),
		Channel:         m.notificationMapper.ProtoToSQLCNotificationChannel(proto.Channel),
		SubjectTemplate: proto.SubjectTemplate,
		BodyTemplate:    proto.BodyTemplate,
		CreatedBy:       proto.CreatedBy,
	}

	// Handle UUID
	if proto.Id != "" {
		if id, err := uuid.Parse(proto.Id); err == nil {
			sqlcTemplate.ID = id
		}
	}

	// Handle optional fields
	if proto.Language != "" {
		sqlcTemplate.Language = &proto.Language
	}

	if proto.Variables != nil {
		if variables, err := json.Marshal(proto.Variables); err == nil {
			sqlcTemplate.Variables = variables
		}
	}

	if proto.Active {
		sqlcTemplate.Active = &proto.Active
	}

	// Handle timestamps
	if proto.CreatedAt != nil {
		sqlcTemplate.CreatedAt = proto.CreatedAt.AsTime()
	}
	if proto.UpdatedAt != nil {
		sqlcTemplate.UpdatedAt = proto.UpdatedAt.AsTime()
	}

	return sqlcTemplate, nil
}

// SQLCToProtoTemplate converts SQLC NotificationTemplate to proto NotificationTemplate
func (m *TemplatePreferenceMapper) SQLCToProtoTemplate(sqlc *db.NotificationTemplate) (*notification.NotificationTemplate, error) {
	if sqlc == nil {
		return nil, nil
	}

	proto := &notification.NotificationTemplate{
		Id:              sqlc.ID.String(),
		Name:            sqlc.Name,
		Type:            m.notificationMapper.SQLCToProtoNotificationType(sqlc.Type),
		Channel:         m.notificationMapper.SQLCToProtoNotificationChannel(sqlc.Channel),
		SubjectTemplate: sqlc.SubjectTemplate,
		BodyTemplate:    sqlc.BodyTemplate,
		CreatedBy:       sqlc.CreatedBy,
		CreatedAt:       timestamppb.New(sqlc.CreatedAt),
		UpdatedAt:       timestamppb.New(sqlc.UpdatedAt),
	}

	// Handle optional fields
	if sqlc.Language != nil {
		proto.Language = *sqlc.Language
	}

	if sqlc.Active != nil {
		proto.Active = *sqlc.Active
	}

	// Handle variables
	if len(sqlc.Variables) > 0 {
		var variables map[string]string
		if err := json.Unmarshal(sqlc.Variables, &variables); err == nil {
			proto.Variables = variables
		}
	}

	return proto, nil
}

// Preference Mappers

// ProtoToSQLCPreference converts proto NotificationPreference to SQLC NotificationPreference
func (m *TemplatePreferenceMapper) ProtoToSQLCPreference(proto *notification.NotificationPreference) (*db.NotificationPreference, error) {
	if proto == nil {
		return nil, nil
	}

	sqlcPreference := &db.NotificationPreference{
		UserID: proto.UserId,
		Type:   m.notificationMapper.ProtoToSQLCNotificationType(proto.Type),
	}

	// Handle UUID
	if proto.UserId != "" {
		// Note: UserID is stored as string in the database, not UUID
		sqlcPreference.UserID = proto.UserId
	}

	// Handle enabled channels
	if len(proto.EnabledChannels) > 0 {
		channels := make([]db.NotificationChannel, len(proto.EnabledChannels))
		for i, channel := range proto.EnabledChannels {
			channels[i] = m.notificationMapper.ProtoToSQLCNotificationChannel(channel)
		}
		sqlcPreference.EnabledChannels = channels
	}

	// Handle enabled flag
	if proto.Enabled {
		sqlcPreference.Enabled = &proto.Enabled
	}

	// Handle settings
	if proto.Settings != nil {
		if settings, err := json.Marshal(proto.Settings); err == nil {
			sqlcPreference.Settings = settings
		}
	}

	// Handle timestamps
	if proto.CreatedAt != nil {
		sqlcPreference.CreatedAt = proto.CreatedAt.AsTime()
	}
	if proto.UpdatedAt != nil {
		sqlcPreference.UpdatedAt = proto.UpdatedAt.AsTime()
	}

	return sqlcPreference, nil
}

// SQLCToProtoPreference converts SQLC NotificationPreference to proto NotificationPreference
func (m *TemplatePreferenceMapper) SQLCToProtoPreference(sqlc *db.NotificationPreference) (*notification.NotificationPreference, error) {
	if sqlc == nil {
		return nil, nil
	}

	proto := &notification.NotificationPreference{
		UserId:    sqlc.UserID,
		Type:      m.notificationMapper.SQLCToProtoNotificationType(sqlc.Type),
		CreatedAt: timestamppb.New(sqlc.CreatedAt),
		UpdatedAt: timestamppb.New(sqlc.UpdatedAt),
	}

	// Handle enabled channels
	if len(sqlc.EnabledChannels) > 0 {
		channels := make([]notification.NotificationChannel, len(sqlc.EnabledChannels))
		for i, channel := range sqlc.EnabledChannels {
			channels[i] = m.notificationMapper.SQLCToProtoNotificationChannel(channel)
		}
		proto.EnabledChannels = channels
	}

	// Handle enabled flag
	if sqlc.Enabled != nil {
		proto.Enabled = *sqlc.Enabled
	}

	// Handle settings
	if len(sqlc.Settings) > 0 {
		var settings map[string]string
		if err := json.Unmarshal(sqlc.Settings, &settings); err == nil {
			proto.Settings = settings
		}
	}

	return proto, nil
}

// Specialized Notification Mappers for Workflow and SLA Events

// ProtoToSQLCWorkflowNotification converts proto WorkflowStateChangeNotification to SQLC WorkflowStateNotification
func (m *TemplatePreferenceMapper) ProtoToSQLCWorkflowNotification(proto *notification.WorkflowStateChangeNotification, notificationID uuid.UUID) (*db.WorkflowStateNotification, error) {
	if proto == nil {
		return nil, nil
	}

	sqlcWorkflow := &db.WorkflowStateNotification{
		NotificationID: pgtype.UUID{Bytes: notificationID, Valid: true},
		CurrentState:   proto.CurrentState,
		ChangedBy:      proto.ChangedBy,
	}

	// Handle UUIDs
	if proto.InstanceId != "" {
		if instanceID, err := uuid.Parse(proto.InstanceId); err == nil {
			sqlcWorkflow.InstanceID = instanceID
		}
	}

	if proto.FormId != "" {
		if formID, err := uuid.Parse(proto.FormId); err == nil {
			sqlcWorkflow.FormID = formID
		}
	}

	// Handle optional fields
	if proto.PreviousState != "" {
		sqlcWorkflow.PreviousState = &proto.PreviousState
	}
	if proto.AssignedTo != "" {
		sqlcWorkflow.AssignedTo = &proto.AssignedTo
	}
	if proto.AssignedRole != "" {
		sqlcWorkflow.AssignedRole = &proto.AssignedRole
	}
	if proto.TransitionEvent != "" {
		sqlcWorkflow.TransitionEvent = &proto.TransitionEvent
	}
	if proto.Comment != "" {
		sqlcWorkflow.Comment = &proto.Comment
	}

	// Handle form data
	if proto.FormData != nil {
		if formData, err := m.notificationMapper.ProtoMapToJSON(proto.FormData); err == nil {
			sqlcWorkflow.FormData = formData
		}
	}

	// Handle timestamp
	if proto.ChangedAt != nil {
		sqlcWorkflow.ChangedAt = pgtype.Timestamptz{Time: proto.ChangedAt.AsTime(), Valid: true}
	}

	return sqlcWorkflow, nil
}

// SQLCToProtoWorkflowNotification converts SQLC WorkflowStateNotification to proto WorkflowStateChangeNotification
func (m *TemplatePreferenceMapper) SQLCToProtoWorkflowNotification(sqlc *db.WorkflowStateNotification) (*notification.WorkflowStateChangeNotification, error) {
	if sqlc == nil {
		return nil, nil
	}

	proto := &notification.WorkflowStateChangeNotification{
		InstanceId:   sqlc.InstanceID.String(),
		FormId:       sqlc.FormID.String(),
		CurrentState: sqlc.CurrentState,
		ChangedBy:    sqlc.ChangedBy,
	}

	// Handle optional fields
	if sqlc.PreviousState != nil {
		proto.PreviousState = *sqlc.PreviousState
	}
	if sqlc.AssignedTo != nil {
		proto.AssignedTo = *sqlc.AssignedTo
	}
	if sqlc.AssignedRole != nil {
		proto.AssignedRole = *sqlc.AssignedRole
	}
	if sqlc.TransitionEvent != nil {
		proto.TransitionEvent = *sqlc.TransitionEvent
	}
	if sqlc.Comment != nil {
		proto.Comment = *sqlc.Comment
	}

	// Handle form data
	if len(sqlc.FormData) > 0 {
		if formData, err := m.notificationMapper.JSONToProtoMap(sqlc.FormData); err == nil {
			proto.FormData = formData
		}
	}

	// Handle timestamp
	if sqlc.ChangedAt.Valid {
		proto.ChangedAt = timestamppb.New(sqlc.ChangedAt.Time)
	}

	return proto, nil
}

// ProtoToSQLCSLANotification converts proto SLAEventNotification to SQLC SlaEventNotification
func (m *TemplatePreferenceMapper) ProtoToSQLCSLANotification(proto *notification.SLAEventNotification, notificationID uuid.UUID) (*db.SlaEventNotification, error) {
	if proto == nil {
		return nil, nil
	}

	sqlcSLA := &db.SlaEventNotification{
		NotificationID: pgtype.UUID{Bytes: notificationID, Valid: true},
		SlaRuleName:    proto.SlaRuleName,
		EventType:      m.ProtoToSQLCSLAEventType(proto.EventType),
		State:          proto.State,
	}

	// Handle UUIDs
	if proto.SlaInstanceId != "" {
		if slaInstanceID, err := uuid.Parse(proto.SlaInstanceId); err == nil {
			sqlcSLA.SlaInstanceID = slaInstanceID
		}
	}

	if proto.InstanceId != "" {
		if instanceID, err := uuid.Parse(proto.InstanceId); err == nil {
			sqlcSLA.InstanceID = instanceID
		}
	}

	if proto.SlaRuleId != "" {
		if slaRuleID, err := uuid.Parse(proto.SlaRuleId); err == nil {
			sqlcSLA.SlaRuleID = slaRuleID
		}
	}

	// Handle timestamps
	if proto.DueTime != nil {
		sqlcSLA.DueTime = proto.DueTime.AsTime()
	}
	if proto.BreachTime != nil {
		sqlcSLA.BreachTime = sql.NullTime{Time: proto.BreachTime.AsTime(), Valid: true}
	}

	// Handle optional fields
	if proto.AssignedTo != "" {
		sqlcSLA.AssignedTo = &proto.AssignedTo
	}
	if proto.Severity != "" {
		sqlcSLA.Severity = &proto.Severity
	}
	if proto.EscalationLevel > 0 {
		sqlcSLA.EscalationLevel = &proto.EscalationLevel
	}
	if proto.BreachDuration != "" {
		sqlcSLA.BreachDuration = &proto.BreachDuration
	}

	// Handle context
	if proto.Context != nil {
		if context, err := json.Marshal(proto.Context); err == nil {
			sqlcSLA.Context = context
		}
	}

	return sqlcSLA, nil
}

// SQLCToProtoSLANotification converts SQLC SlaEventNotification to proto SLAEventNotification
func (m *TemplatePreferenceMapper) SQLCToProtoSLANotification(sqlc *db.SlaEventNotification) (*notification.SLAEventNotification, error) {
	if sqlc == nil {
		return nil, nil
	}

	proto := &notification.SLAEventNotification{
		SlaInstanceId: sqlc.SlaInstanceID.String(),
		InstanceId:    sqlc.InstanceID.String(),
		SlaRuleId:     sqlc.SlaRuleID.String(),
		SlaRuleName:   sqlc.SlaRuleName,
		EventType:     m.SQLCToProtoSLAEventType(sqlc.EventType),
		State:         sqlc.State,
		DueTime:       timestamppb.New(sqlc.DueTime),
	}

	// Handle optional fields
	if sqlc.BreachTime.Valid {
		proto.BreachTime = timestamppb.New(sqlc.BreachTime.Time)
	}
	if sqlc.AssignedTo != nil {
		proto.AssignedTo = *sqlc.AssignedTo
	}
	if sqlc.Severity != nil {
		proto.Severity = *sqlc.Severity
	}
	if sqlc.EscalationLevel != nil {
		proto.EscalationLevel = *sqlc.EscalationLevel
	}
	if sqlc.BreachDuration != nil {
		proto.BreachDuration = *sqlc.BreachDuration
	}

	// Handle context
	if len(sqlc.Context) > 0 {
		var context map[string]string
		if err := json.Unmarshal(sqlc.Context, &context); err == nil {
			proto.Context = context
		}
	}

	return proto, nil
}

// SLA Event Type conversions
func (m *TemplatePreferenceMapper) ProtoToSQLCSLAEventType(proto notification.SLAEventType) db.SlaEventType {
	switch proto {
	case notification.SLAEventType_SLA_STARTED:
		return db.SlaEventTypeSLASTARTED
	case notification.SLAEventType_SLA_WARNING_THRESHOLD:
		return db.SlaEventTypeSLAWARNINGTHRESHOLD
	case notification.SLAEventType_SLA_BREACH_IMMINENT:
		return db.SlaEventTypeSLABREACHIMMINENT
	case notification.SLAEventType_SLA_BREACHED:
		return db.SlaEventTypeSLABREACHED
	case notification.SLAEventType_SLA_ESCALATED:
		return db.SlaEventTypeSLAESCALATED
	case notification.SLAEventType_SLA_RESOLVED:
		return db.SlaEventTypeSLARESOLVED
	case notification.SLAEventType_SLA_PAUSED:
		return db.SlaEventTypeSLAPAUSED
	case notification.SLAEventType_SLA_RESUMED:
		return db.SlaEventTypeSLARESUMED
	default:
		return db.SlaEventTypeSLASTARTED
	}
}

func (m *TemplatePreferenceMapper) SQLCToProtoSLAEventType(sqlc db.SlaEventType) notification.SLAEventType {
	switch sqlc {
	case db.SlaEventTypeSLASTARTED:
		return notification.SLAEventType_SLA_STARTED
	case db.SlaEventTypeSLAWARNINGTHRESHOLD:
		return notification.SLAEventType_SLA_WARNING_THRESHOLD
	case db.SlaEventTypeSLABREACHIMMINENT:
		return notification.SLAEventType_SLA_BREACH_IMMINENT
	case db.SlaEventTypeSLABREACHED:
		return notification.SLAEventType_SLA_BREACHED
	case db.SlaEventTypeSLAESCALATED:
		return notification.SLAEventType_SLA_ESCALATED
	case db.SlaEventTypeSLARESOLVED:
		return notification.SLAEventType_SLA_RESOLVED
	case db.SlaEventTypeSLAPAUSED:
		return notification.SLAEventType_SLA_PAUSED
	case db.SlaEventTypeSLARESUMED:
		return notification.SLAEventType_SLA_RESUMED
	default:
		return notification.SLAEventType_SLA_STARTED
	}
}
