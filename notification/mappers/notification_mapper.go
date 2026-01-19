package mappers

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"p9e.in/ugcl/notification/api/v2/notification"
	db "p9e.in/ugcl/notification/db/generated"
)

// NotificationMapper handles conversions between proto and SQLC models
type NotificationMapper struct{}

// NewNotificationMapper creates a new notification mapper
func NewNotificationMapper() *NotificationMapper {
	return &NotificationMapper{}
}

// ProtoToSQLCNotification converts proto Notification to SQLC Notification
func (m *NotificationMapper) ProtoToSQLCNotification(proto *notification.Notification) (*db.Notification, error) {
	if proto == nil {
		return nil, nil
	}

	sqlcNotification := &db.Notification{
		Type:          m.ProtoToSQLCNotificationType(proto.Type),
		Channel:       m.ProtoToSQLCNotificationChannel(proto.Channel),
		Priority:      m.ProtoToSQLCNotificationPriority(proto.Priority),
		RecipientID:   proto.RecipientId,
		RecipientType: proto.RecipientType,
		Subject:       proto.Subject,
		Message:       proto.Message,
	}

	// Handle UUID fields
	if proto.Id != "" {
		if id, err := uuid.Parse(proto.Id); err == nil {
			sqlcNotification.ID = id
		}
	}

	if proto.TemplateId != "" {
		if templateID, err := uuid.Parse(proto.TemplateId); err == nil {
			sqlcNotification.TemplateID = pgtype.UUID{Bytes: templateID, Valid: true}
		}
	}

	// Handle template data
	if proto.TemplateData != nil {
		if templateData, err := m.ProtoMapToJSON(proto.TemplateData); err == nil {
			sqlcNotification.TemplateData = templateData
		}
	}

	// Handle status
	sqlcNotification.Status = m.ProtoToSQLCNotificationStatus(proto.Status)

	// Handle optional fields
	if proto.SourceId != "" {
		sqlcNotification.SourceID = &proto.SourceId
	}
	if proto.SourceType != "" {
		sqlcNotification.SourceType = &proto.SourceType
	}

	// Handle timestamps
	if proto.ScheduledAt != nil {
		sqlcNotification.ScheduledAt = m.ProtoTimestampToPgTimestamp(proto.ScheduledAt)
	}
	if proto.SentAt != nil {
		sqlcNotification.SentAt = m.ProtoTimestampToNullTime(proto.SentAt)
	}
	if proto.DeliveredAt != nil {
		sqlcNotification.DeliveredAt = m.ProtoTimestampToNullTime(proto.DeliveredAt)
	}
	if proto.ReadAt != nil {
		sqlcNotification.ReadAt = m.ProtoTimestampToPgTimestamp(proto.ReadAt)
	}
	if proto.NextRetryAt != nil {
		sqlcNotification.NextRetryAt = m.ProtoTimestampToNullTime(proto.NextRetryAt)
	}

	// Handle other fields
	if proto.ErrorMessage != "" {
		sqlcNotification.ErrorMessage = &proto.ErrorMessage
	}
	if proto.RetryCount > 0 {
		sqlcNotification.RetryCount = &proto.RetryCount
	}

	// Handle metadata
	if proto.Metadata != nil {
		if metadata, err := m.ProtoStringMapToJSON(proto.Metadata); err == nil {
			sqlcNotification.Metadata = metadata
		}
	}

	// Handle timestamps
	if proto.CreatedAt != nil {
		sqlcNotification.CreatedAt = proto.CreatedAt.AsTime()
	}
	if proto.UpdatedAt != nil {
		sqlcNotification.UpdatedAt = proto.UpdatedAt.AsTime()
	}

	return sqlcNotification, nil
}

// SQLCToProtoNotification converts SQLC Notification to proto Notification
func (m *NotificationMapper) SQLCToProtoNotification(sqlc *db.Notification) (*notification.Notification, error) {
	if sqlc == nil {
		return nil, nil
	}

	proto := &notification.Notification{
		Id:            sqlc.ID.String(),
		Type:          m.SQLCToProtoNotificationType(sqlc.Type),
		Channel:       m.SQLCToProtoNotificationChannel(sqlc.Channel),
		Priority:      m.SQLCToProtoNotificationPriority(sqlc.Priority),
		RecipientId:   sqlc.RecipientID,
		RecipientType: sqlc.RecipientType,
		Subject:       sqlc.Subject,
		Message:       sqlc.Message,
		Status:        m.SQLCToProtoNotificationStatus(sqlc.Status),
		RetryCount:    m.Int32PtrToInt32(sqlc.RetryCount),
		CreatedAt:     timestamppb.New(sqlc.CreatedAt),
		UpdatedAt:     timestamppb.New(sqlc.UpdatedAt),
	}

	// Handle template ID
	if sqlc.TemplateID.Valid {
		templateUUID := uuid.UUID(sqlc.TemplateID.Bytes)
		proto.TemplateId = templateUUID.String()
	}

	// Handle template data
	if len(sqlc.TemplateData) > 0 {
		if templateData, err := m.JSONToProtoMap(sqlc.TemplateData); err == nil {
			proto.TemplateData = templateData
		}
	}

	// Handle optional fields
	if sqlc.SourceID != nil {
		proto.SourceId = *sqlc.SourceID
	}
	if sqlc.SourceType != nil {
		proto.SourceType = *sqlc.SourceType
	}
	if sqlc.ErrorMessage != nil {
		proto.ErrorMessage = *sqlc.ErrorMessage
	}

	// Handle timestamps
	if sqlc.ScheduledAt.Valid {
		proto.ScheduledAt = timestamppb.New(sqlc.ScheduledAt.Time)
	}
	if sqlc.SentAt.Valid {
		proto.SentAt = timestamppb.New(sqlc.SentAt.Time)
	}
	if sqlc.DeliveredAt.Valid {
		proto.DeliveredAt = timestamppb.New(sqlc.DeliveredAt.Time)
	}
	if sqlc.ReadAt.Valid {
		proto.ReadAt = timestamppb.New(sqlc.ReadAt.Time)
	}
	if sqlc.NextRetryAt.Valid {
		proto.NextRetryAt = timestamppb.New(sqlc.NextRetryAt.Time)
	}

	// Handle metadata
	if len(sqlc.Metadata) > 0 {
		if metadata, err := m.JSONToProtoStringMap(sqlc.Metadata); err == nil {
			proto.Metadata = metadata
		}
	}

	return proto, nil
}

// Enum conversion methods
func (m *NotificationMapper) ProtoToSQLCNotificationType(proto notification.NotificationType) db.NotificationType {
	switch proto {
	case notification.NotificationType_WORKFLOW_STATE_CHANGE:
		return db.NotificationTypeWORKFLOWSTATECHANGE
	case notification.NotificationType_WORKFLOW_ASSIGNMENT:
		return db.NotificationTypeWORKFLOWASSIGNMENT
	case notification.NotificationType_WORKFLOW_APPROVAL_REQUEST:
		return db.NotificationTypeWORKFLOWAPPROVALREQUEST
	case notification.NotificationType_WORKFLOW_APPROVAL_RESPONSE:
		return db.NotificationTypeWORKFLOWAPPROVALRESPONSE
	case notification.NotificationType_SLA_WARNING:
		return db.NotificationTypeSLAWARNING
	case notification.NotificationType_SLA_BREACH:
		return db.NotificationTypeSLABREACH
	case notification.NotificationType_SLA_ESCALATION:
		return db.NotificationTypeSLAESCALATION
	case notification.NotificationType_SLA_COMPLETION:
		return db.NotificationTypeSLACOMPLETION
	case notification.NotificationType_FORM_SUBMISSION:
		return db.NotificationTypeFORMSUBMISSION
	case notification.NotificationType_FORM_UPDATE:
		return db.NotificationTypeFORMUPDATE
	case notification.NotificationType_SYSTEM_ALERT:
		return db.NotificationTypeSYSTEMALERT
	case notification.NotificationType_CUSTOM:
		return db.NotificationTypeCUSTOM
	default:
		return db.NotificationTypeCUSTOM
	}
}

func (m *NotificationMapper) SQLCToProtoNotificationType(sqlc db.NotificationType) notification.NotificationType {
	switch sqlc {
	case db.NotificationTypeWORKFLOWSTATECHANGE:
		return notification.NotificationType_WORKFLOW_STATE_CHANGE
	case db.NotificationTypeWORKFLOWASSIGNMENT:
		return notification.NotificationType_WORKFLOW_ASSIGNMENT
	case db.NotificationTypeWORKFLOWAPPROVALREQUEST:
		return notification.NotificationType_WORKFLOW_APPROVAL_REQUEST
	case db.NotificationTypeWORKFLOWAPPROVALRESPONSE:
		return notification.NotificationType_WORKFLOW_APPROVAL_RESPONSE
	case db.NotificationTypeSLAWARNING:
		return notification.NotificationType_SLA_WARNING
	case db.NotificationTypeSLABREACH:
		return notification.NotificationType_SLA_BREACH
	case db.NotificationTypeSLAESCALATION:
		return notification.NotificationType_SLA_ESCALATION
	case db.NotificationTypeSLACOMPLETION:
		return notification.NotificationType_SLA_COMPLETION
	case db.NotificationTypeFORMSUBMISSION:
		return notification.NotificationType_FORM_SUBMISSION
	case db.NotificationTypeFORMUPDATE:
		return notification.NotificationType_FORM_UPDATE
	case db.NotificationTypeSYSTEMALERT:
		return notification.NotificationType_SYSTEM_ALERT
	case db.NotificationTypeCUSTOM:
		return notification.NotificationType_CUSTOM
	default:
		return notification.NotificationType_CUSTOM
	}
}

func (m *NotificationMapper) ProtoToSQLCNotificationChannel(proto notification.NotificationChannel) db.NotificationChannel {
	switch proto {
	case notification.NotificationChannel_EMAIL:
		return db.NotificationChannelEMAIL
	case notification.NotificationChannel_SMS:
		return db.NotificationChannelSMS
	case notification.NotificationChannel_IN_APP:
		return db.NotificationChannelINAPP
	case notification.NotificationChannel_PUSH:
		return db.NotificationChannelPUSH
	case notification.NotificationChannel_WEBHOOK:
		return db.NotificationChannelWEBHOOK
	case notification.NotificationChannel_SLACK:
		return db.NotificationChannelSLACK
	case notification.NotificationChannel_TEAMS:
		return db.NotificationChannelTEAMS
	default:
		return db.NotificationChannelEMAIL
	}
}

func (m *NotificationMapper) SQLCToProtoNotificationChannel(sqlc db.NotificationChannel) notification.NotificationChannel {
	switch sqlc {
	case db.NotificationChannelEMAIL:
		return notification.NotificationChannel_EMAIL
	case db.NotificationChannelSMS:
		return notification.NotificationChannel_SMS
	case db.NotificationChannelINAPP:
		return notification.NotificationChannel_IN_APP
	case db.NotificationChannelPUSH:
		return notification.NotificationChannel_PUSH
	case db.NotificationChannelWEBHOOK:
		return notification.NotificationChannel_WEBHOOK
	case db.NotificationChannelSLACK:
		return notification.NotificationChannel_SLACK
	case db.NotificationChannelTEAMS:
		return notification.NotificationChannel_TEAMS
	default:
		return notification.NotificationChannel_EMAIL
	}
}

func (m *NotificationMapper) ProtoToSQLCNotificationPriority(proto notification.NotificationPriority) db.NullNotificationPriority {
	var priority db.NotificationPriority
	switch proto {
	case notification.NotificationPriority_LOW:
		priority = db.NotificationPriorityLOW
	case notification.NotificationPriority_NORMAL:
		priority = db.NotificationPriorityNORMAL
	case notification.NotificationPriority_HIGH:
		priority = db.NotificationPriorityHIGH
	case notification.NotificationPriority_URGENT:
		priority = db.NotificationPriorityURGENT
	case notification.NotificationPriority_CRITICAL:
		priority = db.NotificationPriorityCRITICAL
	default:
		priority = db.NotificationPriorityNORMAL
	}
	return db.NullNotificationPriority{NotificationPriority: priority, Valid: true}
}

func (m *NotificationMapper) SQLCToProtoNotificationPriority(sqlc db.NullNotificationPriority) notification.NotificationPriority {
	if !sqlc.Valid {
		return notification.NotificationPriority_NORMAL
	}
	switch sqlc.NotificationPriority {
	case db.NotificationPriorityLOW:
		return notification.NotificationPriority_LOW
	case db.NotificationPriorityNORMAL:
		return notification.NotificationPriority_NORMAL
	case db.NotificationPriorityHIGH:
		return notification.NotificationPriority_HIGH
	case db.NotificationPriorityURGENT:
		return notification.NotificationPriority_URGENT
	case db.NotificationPriorityCRITICAL:
		return notification.NotificationPriority_CRITICAL
	default:
		return notification.NotificationPriority_NORMAL
	}
}

func (m *NotificationMapper) ProtoToSQLCNotificationStatus(proto notification.NotificationStatus) db.NullNotificationStatus {
	var status db.NotificationStatus
	switch proto {
	case notification.NotificationStatus_PENDING:
		status = db.NotificationStatusPENDING
	case notification.NotificationStatus_SCHEDULED:
		status = db.NotificationStatusSCHEDULED
	case notification.NotificationStatus_SENT:
		status = db.NotificationStatusSENT
	case notification.NotificationStatus_DELIVERED:
		status = db.NotificationStatusDELIVERED
	case notification.NotificationStatus_READ:
		status = db.NotificationStatusREAD
	case notification.NotificationStatus_FAILED:
		status = db.NotificationStatusFAILED
	case notification.NotificationStatus_CANCELLED:
		status = db.NotificationStatusCANCELLED
	default:
		status = db.NotificationStatusPENDING
	}
	return db.NullNotificationStatus{NotificationStatus: status, Valid: true}
}

func (m *NotificationMapper) SQLCToProtoNotificationStatus(sqlc db.NullNotificationStatus) notification.NotificationStatus {
	if !sqlc.Valid {
		return notification.NotificationStatus_PENDING
	}
	switch sqlc.NotificationStatus {
	case db.NotificationStatusPENDING:
		return notification.NotificationStatus_PENDING
	case db.NotificationStatusSCHEDULED:
		return notification.NotificationStatus_SCHEDULED
	case db.NotificationStatusSENT:
		return notification.NotificationStatus_SENT
	case db.NotificationStatusDELIVERED:
		return notification.NotificationStatus_DELIVERED
	case db.NotificationStatusREAD:
		return notification.NotificationStatus_READ
	case db.NotificationStatusFAILED:
		return notification.NotificationStatus_FAILED
	case db.NotificationStatusCANCELLED:
		return notification.NotificationStatus_CANCELLED
	default:
		return notification.NotificationStatus_PENDING
	}
}

// Helper methods for type conversions
func (m *NotificationMapper) ProtoTimestampToPgTimestamp(ts *timestamppb.Timestamp) pgtype.Timestamptz {
	if ts == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: ts.AsTime(), Valid: true}
}

func (m *NotificationMapper) ProtoTimestampToNullTime(ts *timestamppb.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: ts.AsTime(), Valid: true}
}

func (m *NotificationMapper) Int32PtrToInt32(ptr *int32) int32 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func (m *NotificationMapper) ProtoMapToJSON(protoMap map[string]*anypb.Any) ([]byte, error) {
	if protoMap == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	for key, value := range protoMap {
		if value != nil {
			// Convert Any to interface{} - this is a simplified conversion
			// In a real implementation, you might want to handle specific types
			result[key] = value.String()
		}
	}

	return json.Marshal(result)
}

func (m *NotificationMapper) JSONToProtoMap(data []byte) (map[string]*anypb.Any, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, err
	}

	result := make(map[string]*anypb.Any)
	for key, value := range jsonMap {
		// Convert interface{} to Any - this is a simplified conversion
		if anyValue, err := anypb.New(&anypb.Any{Value: []byte(value.(string))}); err == nil {
			result[key] = anyValue
		}
	}

	return result, nil
}

func (m *NotificationMapper) ProtoStringMapToJSON(protoMap map[string]string) ([]byte, error) {
	if protoMap == nil {
		return nil, nil
	}
	return json.Marshal(protoMap)
}

func (m *NotificationMapper) JSONToProtoStringMap(data []byte) (map[string]string, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
