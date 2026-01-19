package mappers

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/models"
)

type AuditMapper struct{}

func NewAuditMapper() *AuditMapper {
	return &AuditMapper{}
}

// SecurityEvent mapping functions

func (m *AuditMapper) FromSQLCSecurityEvent(sqlcEvent sqlcgen.AuthSecurityEvent) (*models.SecurityEvent, error) {
	event := &models.SecurityEvent{
		ID:        sqlcEvent.ID,
		UserID:    uuid.UUID(sqlcEvent.UserID.Bytes),
		EventType: sqlcEvent.EventType,
		CreatedAt: sqlcEvent.CreatedAt,
	}

	// Parse JSONB event data
	if len(sqlcEvent.EventData) > 0 {
		var eventData map[string]interface{}
		if err := json.Unmarshal(sqlcEvent.EventData, &eventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}
		event.EventData = eventData
	}

	if sqlcEvent.IpAddress != nil {
		ipStr := sqlcEvent.IpAddress.String()
		event.IPAddress = &ipStr
	}
	if sqlcEvent.UserAgent != nil {
		event.UserAgent = sqlcEvent.UserAgent
	}
	if sqlcEvent.TenantID != nil {
		event.TenantID = sqlcEvent.TenantID
	}

	return event, nil
}

// TODO: Implement when CreateSecurityEventParams is added to SQLC queries
// func (m *AuditMapper) ToCreateSecurityEventParams(event *models.SecurityEvent) (sqlcgen.CreateSecurityEventParams, error) {
// 	var eventDataBytes []byte
// 	var err error

// 	if event.EventData != nil {
// 		eventDataBytes, err = json.Marshal(event.EventData)
// 		if err != nil {
// 			return sqlcgen.CreateSecurityEventParams{}, fmt.Errorf("failed to marshal event data: %w", err)
// 		}
// 	}

// 	params := sqlcgen.CreateSecurityEventParams{
// 		UserID:    pgtype.UUID{Bytes: event.UserID, Valid: true},
// 		EventType: event.EventType,
// 		EventData: eventDataBytes,
// 	}

// 	if event.IPAddress != nil {
// 		if addr, err := netip.ParseAddr(*event.IPAddress); err == nil {
// 			params.IpAddress = &addr
// 		}
// 	}
// 	if event.UserAgent != nil {
// 		params.UserAgent = event.UserAgent
// 	}
// 	if event.TenantID != nil {
// 		params.TenantID = event.TenantID
// 	}

// 	return params, nil
// }

// LoginAttempt mapping functions

func (m *AuditMapper) FromSQLCLoginAttempt(sqlcAttempt sqlcgen.AuthLoginAttempt) *models.LoginAttempt {
	attempt := &models.LoginAttempt{
		ID:          sqlcAttempt.ID,
		Identifier:  sqlcAttempt.Identifier,
		IPAddress:   sqlcAttempt.IpAddress.String(),
		Success:     sqlcAttempt.Success,
		AttemptedAt: sqlcAttempt.AttemptedAt.Time,
	}

	if sqlcAttempt.UserAgent != nil {
		attempt.UserAgent = sqlcAttempt.UserAgent
	}
	if sqlcAttempt.FailureReason != nil {
		attempt.FailureReason = sqlcAttempt.FailureReason
	}
	if sqlcAttempt.TenantID != nil {
		attempt.TenantID = sqlcAttempt.TenantID
	}
	if sqlcAttempt.UserID.Valid {
		userID := uuid.UUID(sqlcAttempt.UserID.Bytes)
		attempt.UserID = &userID
	}

	return attempt
}

// TODO: Implement when CreateLoginAttemptParams is added to SQLC queries
// func (m *AuditMapper) ToCreateLoginAttemptParams(attempt *models.LoginAttempt) (sqlcgen.CreateLoginAttemptParams, error) {
// 	params := sqlcgen.CreateLoginAttemptParams{
// 		Identifier:  attempt.Identifier,
// 		Success:     attempt.Success,
// 		AttemptedAt: pgtype.Timestamp{Time: attempt.AttemptedAt, Valid: true},
// 	}

// 	// Parse IP address
// 	if addr, err := netip.ParseAddr(attempt.IPAddress); err == nil {
// 		params.IpAddress = addr
// 	} else {
// 		return params, fmt.Errorf("failed to parse IP address: %w", err)
// 	}

// 	if attempt.UserAgent != nil {
// 		params.UserAgent = attempt.UserAgent
// 	}
// 	if attempt.FailureReason != nil {
// 		params.FailureReason = attempt.FailureReason
// 	}
// 	if attempt.TenantID != nil {
// 		params.TenantID = attempt.TenantID
// 	}
// 	if attempt.UserID != nil {
// 		params.UserID = pgtype.UUID{Bytes: *attempt.UserID, Valid: true}
// 	}

// 	return params, nil
// }