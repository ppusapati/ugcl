package services

import (
	"context"
	"fmt"

	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/p9log"
)

type UserEventService struct {
	eventPublisher  *domain.DomainEventPublisher
	eventSubscriber *domain.EventSubscriber
	logger          p9log.Helper
	userService     IUserService
}

func NewUserEventService(
	eventPublisher *domain.DomainEventPublisher,
	eventSubscriber *domain.EventSubscriber,
	logger p9log.Logger,
	userService IUserService,
) *UserEventService {
	return &UserEventService{
		eventPublisher:  eventPublisher,
		eventSubscriber: eventSubscriber,
		logger:          *p9log.NewHelper(p9log.With(logger, "component", "UserEventService")),
		userService:     userService,
	}
}

// PublishUserCreatedEvent publishes a user created event
func (s *UserEventService) PublishUserCreatedEvent(ctx context.Context, userID, email, username, createdBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "created",
		ChangedBy: createdBy,
		Attributes: map[string]string{
			"email":    email,
			"username": username,
		},
	}

	return s.eventPublisher.PublishUserCreated(ctx, event)
}

// PublishUserUpdatedEvent publishes a user updated event
func (s *UserEventService) PublishUserUpdatedEvent(ctx context.Context, userID, updatedBy string, changes map[string]string) error {
	event := &domain.IdentityEvent{
		UserID:     userID,
		Action:     "updated",
		ChangedBy:  updatedBy,
		Attributes: changes,
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("user").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishUserDeactivatedEvent publishes a user deactivated event
func (s *UserEventService) PublishUserDeactivatedEvent(ctx context.Context, userID, reason, deactivatedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "deactivated",
		ChangedBy: deactivatedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserDeactivated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityHigh).
		WithSource("user").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishRoleAssignedEvent publishes a role assigned event
func (s *UserEventService) PublishRoleAssignedEvent(ctx context.Context, userID, roleID, assignedBy string, permissions []string) error {
	event := &domain.IdentityEvent{
		UserID:      userID,
		RoleID:      roleID,
		Action:      "role_assigned",
		ChangedBy:   assignedBy,
		Permissions: permissions,
	}

	return s.eventPublisher.PublishUserCreated(ctx, event) // Using existing method for role events
}

// PublishRoleRevokedEvent publishes a role revoked event
func (s *UserEventService) PublishRoleRevokedEvent(ctx context.Context, userID, roleID, revokedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		RoleID:    roleID,
		Action:    "role_revoked",
		ChangedBy: revokedBy,
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeRoleRevoked, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("user").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishAccountLockedEvent publishes an account locked event
func (s *UserEventService) PublishAccountLockedEvent(ctx context.Context, userID, reason, lockedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "account_locked",
		ChangedBy: lockedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserDeactivated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityHigh).
		WithSource("user").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishAccountUnlockedEvent publishes an account unlocked event
func (s *UserEventService) PublishAccountUnlockedEvent(ctx context.Context, userID, reason, unlockedBy string) error {
	event := &domain.IdentityEvent{
		UserID:    userID,
		Action:    "account_unlocked",
		ChangedBy: unlockedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeUserUpdated, userID, "user").
		WithIdentity(event).
		WithPriority(domain.PriorityMedium).
		WithSource("user").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// SubscribeToAuthEvents sets up event subscribers for auth-related events
func (s *UserEventService) SubscribeToAuthEvents(ctx context.Context) error {
	// Subscribe to auth-related events
	disposable, err := s.eventSubscriber.SubscribeToAllEvents(s.handleAuthEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to auth events: %w", err)
	}

	s.logger.Infof("Successfully subscribed to auth events")

	// Store disposable for cleanup if needed
	_ = disposable

	return nil
}

// handleAuthEvent processes incoming auth events
func (s *UserEventService) handleAuthEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing auth event: %s of type %s", event.ID, event.Type)

	// Handle event types that are relevant to user service
	switch {
	case event.Source == "auth" && s.isAccountLockRequest(event):
		return s.handleAccountLockRequest(ctx, event)
	case event.Source == "auth" && s.isAccountUnlockRequest(event):
		return s.handleAccountUnlockRequest(ctx, event)
	case event.Source == "auth" && s.isAccountStatusRequest(event):
		return s.handleAccountStatusRequest(ctx, event)
	}

	return nil
}

// Helper methods to identify event types
func (s *UserEventService) isAccountLockRequest(event *domain.DomainEvent) bool {
	if data, ok := event.Data["action"].(string); ok {
		return data == "account_lock_requested"
	}
	return false
}

func (s *UserEventService) isAccountUnlockRequest(event *domain.DomainEvent) bool {
	if data, ok := event.Data["action"].(string); ok {
		return data == "account_unlock_requested"
	}
	return false
}

func (s *UserEventService) isAccountStatusRequest(event *domain.DomainEvent) bool {
	if data, ok := event.Data["action"].(string); ok {
		return data == "account_status_requested"
	}
	return false
}

// handleAccountLockRequest processes account lock requests from auth service
func (s *UserEventService) handleAccountLockRequest(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing account lock request for user: %s", event.AggregateID)

	reason := "system_request"
	if data, ok := event.Data["reason"].(string); ok {
		reason = data
	}

	changedBy := "system"
	if data, ok := event.Data["changed_by"].(string); ok {
		changedBy = data
	}

	// Call user service to lock the account
	// err := s.userService.LockAccount(ctx, event.AggregateID, reason)
	// if err != nil {
	//     s.logger.Errorf("Failed to lock account for user %s: %v", event.AggregateID, err)
	//     return err
	// }

	// Publish account locked event
	return s.PublishAccountLockedEvent(ctx, event.AggregateID, reason, changedBy)
}

// handleAccountUnlockRequest processes account unlock requests from auth service
func (s *UserEventService) handleAccountUnlockRequest(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing account unlock request for user: %s", event.AggregateID)

	reason := "system_request"
	if data, ok := event.Data["reason"].(string); ok {
		reason = data
	}

	changedBy := "system"
	if data, ok := event.Data["changed_by"].(string); ok {
		changedBy = data
	}

	// Call user service to unlock the account
	// err := s.userService.UnlockAccount(ctx, event.AggregateID, reason)
	// if err != nil {
	//     s.logger.Errorf("Failed to unlock account for user %s: %v", event.AggregateID, err)
	//     return err
	// }

	// Publish account unlocked event
	return s.PublishAccountUnlockedEvent(ctx, event.AggregateID, reason, changedBy)
}

// handleAccountStatusRequest processes account status requests from auth service
func (s *UserEventService) handleAccountStatusRequest(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing account status request for user: %s", event.AggregateID)

	// Get user status and publish update event
	// user, err := s.userService.GetUser(ctx, event.AggregateID)
	// if err != nil {
	//     s.logger.Errorf("Failed to get user status for user %s: %v", event.AggregateID, err)
	//     return err
	// }

	// For now, just log that we received the request
	s.logger.Infof("Account status requested for user: %s", event.AggregateID)

	return nil
}