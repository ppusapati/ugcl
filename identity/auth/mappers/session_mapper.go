package mappers

import (
	"encoding/json"
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sqlcgen "p9e.in/ugcl/identity/auth/db/generated"
	"p9e.in/ugcl/identity/auth/models"
)

type SessionMapper struct{}

func NewSessionMapper() *SessionMapper {
	return &SessionMapper{}
}

// Session mapping functions

func (m *SessionMapper) FromSQLCSession(sqlcSession sqlcgen.AuthSession) (*models.Session, error) {
	session := &models.Session{
		ID:               sqlcSession.ID,
		SessionID:        sqlcSession.SessionID,
		UserID:           uuid.UUID(sqlcSession.UserID.Bytes),
		RefreshTokenHash: sqlcSession.RefreshTokenHash,
		IsActive:         func() bool { if sqlcSession.IsActive != nil { return *sqlcSession.IsActive }; return false }(),
		ExpiresAt:        sqlcSession.ExpiresAt.Time,
		CreatedAt:        sqlcSession.CreatedAt,
	}

	if sqlcSession.TenantID != nil {
		session.TenantID = sqlcSession.TenantID
	}

	// Parse JSONB device info
	if len(sqlcSession.DeviceInfo) > 0 {
		var deviceInfo map[string]interface{}
		if err := json.Unmarshal(sqlcSession.DeviceInfo, &deviceInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
		}
		session.DeviceInfo = deviceInfo
	}

	if sqlcSession.IpAddress != nil {
		ipStr := sqlcSession.IpAddress.String()
		session.IPAddress = &ipStr
	}
	if sqlcSession.UserAgent != nil {
		session.UserAgent = sqlcSession.UserAgent
	}
	if sqlcSession.LastAccessedAt.Valid {
		session.LastAccessedAt = &sqlcSession.LastAccessedAt.Time
	}
	if sqlcSession.RevokedAt.Valid {
		session.RevokedAt = &sqlcSession.RevokedAt.Time
	}
	if sqlcSession.RevokedReason != nil {
		session.RevokedReason = sqlcSession.RevokedReason
	}

	return session, nil
}

func (m *SessionMapper) ToCreateSessionParams(session *models.Session) (sqlcgen.CreateAuthSessionParams, error) {
	var deviceInfoBytes []byte
	var err error

	if session.DeviceInfo != nil {
		deviceInfoBytes, err = json.Marshal(session.DeviceInfo)
		if err != nil {
			return sqlcgen.CreateAuthSessionParams{}, fmt.Errorf("failed to marshal device info: %w", err)
		}
	}

	params := sqlcgen.CreateAuthSessionParams{
		SessionID:        session.SessionID,
		UserID:           pgtype.UUID{Bytes: session.UserID, Valid: true},
		RefreshTokenHash: session.RefreshTokenHash,
		DeviceInfo:       deviceInfoBytes,
		// IsActive:         &session.IsActive, // Field not available in CreateAuthSessionParams
		ExpiresAt:        pgtype.Timestamp{Time: session.ExpiresAt, Valid: true},
	}

	if session.TenantID != nil {
		params.TenantID = session.TenantID
	}
	if session.IPAddress != nil {
		if addr, err := netip.ParseAddr(*session.IPAddress); err == nil {
			params.IpAddress = &addr
		}
	}
	if session.UserAgent != nil {
		params.UserAgent = session.UserAgent
	}

	return params, nil
}

// TODO: Implement when UpdateAuthSessionParams is added to SQLC queries
// func (m *SessionMapper) ToUpdateSessionParams(session *models.Session) (sqlcgen.UpdateAuthSessionParams, error) {
// 	var deviceInfoBytes []byte
// 	var err error

// 	if session.DeviceInfo != nil {
// 		deviceInfoBytes, err = json.Marshal(session.DeviceInfo)
// 		if err != nil {
// 			return sqlcgen.UpdateAuthSessionParams{}, fmt.Errorf("failed to marshal device info: %w", err)
// 		}
// 	}

// 	params := sqlcgen.UpdateAuthSessionParams{
// 		ID:               session.ID,
// 		RefreshTokenHash: session.RefreshTokenHash,
// 		DeviceInfo:       deviceInfoBytes,
// 		IsActive:         &session.IsActive,
// 		ExpiresAt:        pgtype.Timestamp{Time: session.ExpiresAt, Valid: true},
// 	}

// 	if session.TenantID != nil {
// 		params.TenantID = session.TenantID
// 	}
// 	if session.IPAddress != nil {
// 		if addr, err := netip.ParseAddr(*session.IPAddress); err == nil {
// 			params.IpAddress = &addr
// 		}
// 	}
// 	if session.UserAgent != nil {
// 		params.UserAgent = session.UserAgent
// 	}
// 	if session.LastAccessedAt != nil {
// 		params.LastAccessedAt = pgtype.Timestamp{Time: *session.LastAccessedAt, Valid: true}
// 	}
// 	if session.RevokedAt != nil {
// 		params.RevokedAt = pgtype.Timestamp{Time: *session.RevokedAt, Valid: true}
// 	}
// 	if session.RevokedReason != nil {
// 		params.RevokedReason = session.RevokedReason
// 	}

// 	return params, nil
// }

