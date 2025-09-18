package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/repository/interfaces"
	serviceInterfaces "p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

type sessionService struct {
	uowFactory        uow.UnitOfWorkFactory
	sessionRepository interfaces.SessionRepository
}

func NewSessionService(uowFactory uow.UnitOfWorkFactory, sessionRepository interfaces.SessionRepository) serviceInterfaces.SessionService {
	return &sessionService{
		uowFactory:        uowFactory,
		sessionRepository: sessionRepository,
	}
}

func (s *sessionService) GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	sessions, err := s.sessionRepository.GetActiveSessionsByUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return sessions, nil
}

func (s *sessionService) RevokeSession(ctx context.Context, sessionID, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.uowFactory.Create().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get session to validate it belongs to the user
		session, err := uow.Sessions().GetBySessionID(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("session not found: %w", err)
		}

		if session == nil {
			return fmt.Errorf("session not found")
		}

		// Verify session belongs to the user
		if session.UserID != userUUID {
			return fmt.Errorf("session does not belong to user")
		}

		// Check if session is already inactive
		if !session.IsActive {
			return fmt.Errorf("session is already inactive")
		}

		// Deactivate the session
		err = uow.Sessions().Deactivate(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to deactivate session: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, uow, userUUID, "session_revoked", map[string]interface{}{
			"session_id": sessionID,
			"revoked_by": "user",
		})

		return nil
	})
}

func (s *sessionService) RevokeAllSessions(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.uowFactory.Create().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get current active sessions count for logging
		activeSessions, err := uow.Sessions().GetActiveSessionsByUser(ctx, userUUID)
		if err != nil {
			log.Printf("Warning: failed to get active sessions count: %v", err)
		}

		// Deactivate all user sessions
		err = uow.Sessions().DeactivateAll(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("failed to deactivate all sessions: %w", err)
		}

		// Log security event
		s.logSecurityEvent(ctx, uow, userUUID, "all_sessions_revoked", map[string]interface{}{
			"user_id":       userUUID,
			"session_count": len(activeSessions),
			"revoked_by":    "user",
		})

		return nil
	})
}

func (s *sessionService) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.sessionRepository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

func (s *sessionService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if session is active
	if !session.IsActive {
		return nil, fmt.Errorf("session is inactive")
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		// Automatically deactivate expired session
		_ = s.sessionRepository.Deactivate(ctx, sessionID)
		return nil, fmt.Errorf("session has expired")
	}

	// Update last accessed time
	err = s.sessionRepository.UpdateLastAccessed(ctx, sessionID)
	if err != nil {
		log.Printf("Warning: failed to update session last accessed time: %v", err)
	}

	return session, nil
}

// Helper methods

func (s *sessionService) logSecurityEvent(ctx context.Context, uow uow.UnitOfWork, userID uuid.UUID, eventType string, eventData map[string]interface{}) {
	event := &models.SecurityEvent{
		UserID:    userID,
		EventType: eventType,
		EventData: eventData,
		CreatedAt: time.Now(),
	}

	_, err := uow.Audit().CreateSecurityEvent(ctx, event)
	if err != nil {
		log.Printf("Failed to log security event: %v", err)
	}
}
