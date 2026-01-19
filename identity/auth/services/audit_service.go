package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	// "p9e.in/ugcl/identity/auth"
	"p9e.in/ugcl/identity/auth"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
)

type auditService struct {
	authModule *auth.Module
}

func NewAuditService(authModule *auth.Module) interfaces.AuditService {
	return &auditService{
		authModule: authModule,
	}
}

func (s *auditService) GetSecurityEvents(ctx context.Context, userID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	if pagination == nil {
		pagination = &models.Pagination{
			Limit:  50,
			Offset: 0,
		}
	}

	events, err := s.authModule.Audit().GetSecurityEventsByUser(ctx, userUUID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	return events, nil
}

func (s *auditService) GetLoginAttempts(ctx context.Context, userID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	if pagination == nil {
		pagination = &models.Pagination{
			Limit:  50,
			Offset: 0,
		}
	}

	attempts, err := s.authModule.Audit().GetLoginAttemptsByUser(ctx, userUUID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get login attempts: %w", err)
	}

	return attempts, nil
}

func (s *auditService) GetSuspiciousActivities(ctx context.Context, since time.Time, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error) {
	if pagination == nil {
		pagination = &models.Pagination{
			Limit:  100,
			Offset: 0,
		}
	}

	activities, err := s.authModule.Audit().GetSuspiciousActivities(ctx, since, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get suspicious activities: %w", err)
	}

	return activities, nil
}

func (s *auditService) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	_, err := s.authModule.Audit().CreateSecurityEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to log security event: %w", err)
	}

	return nil
}
