package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/p9log"
)

type eventService struct {
	authModule      *auth.Module
	eventPublisher  *domain.DomainEventPublisher
	eventSubscriber *domain.EventSubscriber
	logger          p9log.Helper
}

func NewEventService(authModule *auth.Module, eventPublisher *domain.DomainEventPublisher, eventSubscriber *domain.EventSubscriber, logger p9log.Logger) interfaces.EventService {
	return &eventService{
		authModule:      authModule,
		eventPublisher:  eventPublisher,
		eventSubscriber: eventSubscriber,
		logger:          *p9log.NewHelper(p9log.With(logger, "component", "AuthEventService")),
	}
}

// PublishUserLoginEvent publishes a user login event
func (s *eventService) PublishUserLoginEvent(ctx context.Context, user *models.User, sessionID, tenantID, ipAddress string) error {
	event := &domain.IdentityEvent{
		UserID:    user.ID.String(),
		Action:    "login",
		ChangedBy: user.ID.String(),
		Attributes: map[string]string{
			"session_id": sessionID,
			"tenant_id":  tenantID,
			"ip_address": ipAddress,
			"email":      user.Email,
			"username":   user.Username,
		},
	}

	return s.eventPublisher.PublishUserCreated(ctx, event)
}

// PublishUserLogoutEvent publishes a user logout event
func (s *eventService) PublishUserLogoutEvent(ctx context.Context, userID, sessionID, reason string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "logout",
		ChangedBy: userID,
		Attributes: map[string]string{
			"session_id": sessionID,
			"reason":     reason,
		},
	}

	// Create a custom logout event since we don't have a specific one
	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityLow).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishAccountLockEvent publishes an account lock event
func (s *eventService) PublishAccountLockEvent(ctx context.Context, userID, reason, changedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "account_locked",
		ChangedBy: changedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserDeactivated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityHigh).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishAccountUnlockEvent publishes an account unlock event
func (s *eventService) PublishAccountUnlockEvent(ctx context.Context, userID, reason, changedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "account_unlocked",
		ChangedBy: changedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishTwoFactorEnabledEvent publishes a 2FA enabled event
func (s *eventService) PublishTwoFactorEnabledEvent(ctx context.Context, userID, method string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "two_factor_enabled",
		ChangedBy: userID,
		Attributes: map[string]string{
			"method": method,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishTwoFactorDisabledEvent publishes a 2FA disabled event
func (s *eventService) PublishTwoFactorDisabledEvent(ctx context.Context, userID string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "two_factor_disabled",
		ChangedBy: userID,
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishSessionRevokedEvent publishes a session revoked event
func (s *eventService) PublishSessionRevokedEvent(ctx context.Context, userID, sessionID, reason string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "session_revoked",
		ChangedBy: userID,
		Attributes: map[string]string{
			"session_id": sessionID,
			"reason":     reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishFailedLoginEvent publishes a failed login attempt event
func (s *eventService) PublishFailedLoginEvent(ctx context.Context, identifier, reason, ipAddress string) error {
	event := &domain.IdentityEvent{
		UserID:    identifier, // Use identifier as placeholder
		Action:    "login_failed",
		ChangedBy: "system",
		Attributes: map[string]string{
			"identifier": identifier,
			"reason":     reason,
			"ip_address": ipAddress,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, identifier, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityHigh).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// SubscribeToUserEvents sets up event subscribers for user-related events
func (s *eventService) SubscribeToUserEvents(ctx context.Context) error {
	// Subscribe to identity events from other services
	disposable, err := s.eventSubscriber.SubscribeToAllEvents(s.handleIdentityEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to identity events: %w", err)
	}

	s.logger.Infof("Successfully subscribed to identity events")

	// Store disposable for cleanup if needed
	_ = disposable

	return nil
}

// handleIdentityEvent processes incoming identity events
func (s *eventService) handleIdentityEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing identity event: %s of type %s", event.ID, event.Type)

	switch event.Type {
	case domain.EventTypeUserCreated:
		return s.handleUserCreatedEvent(ctx, event)
	case domain.EventTypeUserUpdated:
		return s.handleUserUpdatedEvent(ctx, event)
	case domain.EventTypeUserDeactivated:
		return s.handleUserDeactivatedEvent(ctx, event)
	case domain.EventTypeRoleAssigned:
		return s.handleRoleAssignedEvent(ctx, event)
	case domain.EventTypeRoleRevoked:
		return s.handleRoleRevokedEvent(ctx, event)
	}

	return nil
}

// handleUserCreatedEvent processes user created events
func (s *eventService) handleUserCreatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing user created event for user: %s", event.AggregateID)

	// Update local auth records if needed
	// This could include creating auth-specific user records, setting up default sessions, etc.

	return nil
}

// handleUserUpdatedEvent processes user updated events
func (s *eventService) handleUserUpdatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing user updated event for user: %s", event.AggregateID)

	// Update local auth records if needed
	// This could include updating user permissions cache, invalidating sessions, etc.

	return nil
}

// handleUserDeactivatedEvent processes user deactivated events
func (s *eventService) handleUserDeactivatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing user deactivated event for user: %s", event.AggregateID)

	// Revoke all sessions for the deactivated user
	err := s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		userUUID, err := uuid.Parse(event.AggregateID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %w", err)
		}
		sessions, err := uow.Sessions().GetActiveSessionsByUserID(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("failed to get user sessions: %w", err)
		}

		for _, session := range sessions {
			if err := uow.Sessions().Delete(ctx, session.ID); err != nil {
				s.logger.Errorf("Failed to delete session %s for deactivated user %s: %v", session.ID, event.AggregateID, err)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Errorf("Failed to revoke sessions for deactivated user %s: %v", event.AggregateID, err)
		return err
	}

	s.logger.Infof("Successfully revoked all sessions for deactivated user: %s", event.AggregateID)
	return nil
}

// handleRoleAssignedEvent processes role assigned events
func (s *eventService) handleRoleAssignedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing role assigned event for user: %s", event.AggregateID)

	// Invalidate permission cache or update JWT claims if needed
	// This could include refreshing user permissions in active sessions

	return nil
}

// handleRoleRevokedEvent processes role revoked events
func (s *eventService) handleRoleRevokedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing role revoked event for user: %s", event.AggregateID)

	// Invalidate permission cache or update JWT claims if needed
	// This could include refreshing user permissions in active sessions

	return nil
}

// RequestAccountLock sends an event to request account locking (for cross-module communication)
func (s *eventService) RequestAccountLock(ctx context.Context, userID, reason, requestedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "account_lock_requested",
		ChangedBy: requestedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder("identity.user.lock_requested", userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityHigh).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// RequestAccountUnlock sends an event to request account unlocking (for cross-module communication)
func (s *eventService) RequestAccountUnlock(ctx context.Context, userID, reason, requestedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "account_unlock_requested",
		ChangedBy: requestedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder("identity.user.unlock_requested", userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityHigh).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// RequestAccountStatus sends an event to request account status (for cross-module communication)
func (s *eventService) RequestAccountStatus(ctx context.Context, userID, requestedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "account_status_requested",
		ChangedBy: requestedBy,
	}

	domainEvent := domain.NewEventBuilder("identity.user.status_requested", userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("auth").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}