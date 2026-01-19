package services

import (
	"context"
	"fmt"

	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/p9log"
)

type TenantEventService struct {
	eventPublisher  *domain.DomainEventPublisher
	eventSubscriber *domain.EventSubscriber
	logger          p9log.Helper
	tenantService   ITenantService
}

func NewTenantEventService(
	eventPublisher *domain.DomainEventPublisher,
	eventSubscriber *domain.EventSubscriber,
	logger p9log.Logger,
	tenantService ITenantService,
) *TenantEventService {
	return &TenantEventService{
		eventPublisher:  eventPublisher,
		eventSubscriber: eventSubscriber,
		logger:          *p9log.NewHelper(p9log.With(logger, "component", "TenantEventService")),
		tenantService:   tenantService,
	}
}

// PublishTenantCreatedEvent publishes a tenant created event
func (s *TenantEventService) PublishTenantCreatedEvent(ctx context.Context, tenantID, tenantName, createdBy string) error {
	event := &domain.TenantEvent{
		TenantID:   tenantID,
		TenantName: tenantName,
		Action:     "created",
		ChangedBy:  createdBy,
	}

	return s.eventPublisher.PublishTenantCreated(ctx, event)
}

// PublishTenantUpdatedEvent publishes a tenant updated event
func (s *TenantEventService) PublishTenantUpdatedEvent(ctx context.Context, tenantID, tenantName, updatedBy string, changes map[string]string) error {
	event := &domain.TenantEvent{
		TenantID:   tenantID,
		TenantName: tenantName,
		Action:     "updated",
		ChangedBy:  updatedBy,
		Attributes: changes,
	}

	return s.eventPublisher.PublishTenantUpdated(ctx, event)
}

// PublishTenantDeactivatedEvent publishes a tenant deactivated event
func (s *TenantEventService) PublishTenantDeactivatedEvent(ctx context.Context, tenantID, tenantName, reason, deactivatedBy string) error {
	event := &domain.TenantEvent{
		TenantID:   tenantID,
		TenantName: tenantName,
		Action:     "deactivated",
		ChangedBy:  deactivatedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeTenantDeactivated, tenantID, "tenant").
		WithTenant(event).
		WithPriority(domain.PriorityHigh).
		WithSource("tenant").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// PublishTenantUserAddedEvent publishes a tenant user added event
func (s *TenantEventService) PublishTenantUserAddedEvent(ctx context.Context, tenantID, tenantName, userID, addedBy string) error {
	event := &domain.TenantEvent{
		TenantID:   tenantID,
		TenantName: tenantName,
		UserID:     userID,
		Action:     "user_added",
		ChangedBy:  addedBy,
	}

	return s.eventPublisher.PublishTenantUserAdded(ctx, event)
}

// PublishTenantUserRemovedEvent publishes a tenant user removed event
func (s *TenantEventService) PublishTenantUserRemovedEvent(ctx context.Context, tenantID, tenantName, userID, reason, removedBy string) error {
	event := &domain.TenantEvent{
		TenantID:   tenantID,
		TenantName: tenantName,
		UserID:     userID,
		Action:     "user_removed",
		ChangedBy:  removedBy,
		Attributes: map[string]string{
			"reason": reason,
		},
	}

	domainEvent := domain.NewEventBuilder(domain.EventTypeTenantUserRemoved, tenantID, "tenant").
		WithTenant(event).
		WithPriority(domain.PriorityMedium).
		WithSource("tenant").
		Build()

	return s.eventPublisher.PublishEvent(ctx, domainEvent)
}

// SubscribeToIdentityEvents sets up event subscribers for identity-related events
func (s *TenantEventService) SubscribeToIdentityEvents(ctx context.Context) error {
	// Subscribe to identity events that affect tenants
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
func (s *TenantEventService) handleIdentityEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing identity event: %s of type %s", event.ID, event.Type)

	switch event.Type {
	case domain.EventTypeUserCreated:
		return s.handleUserCreatedEvent(ctx, event)
	case domain.EventTypeUserDeactivated:
		return s.handleUserDeactivatedEvent(ctx, event)
	}

	return nil
}

// handleUserCreatedEvent processes user created events
func (s *TenantEventService) handleUserCreatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing user created event for user: %s", event.AggregateID)

	// Extract tenant information if available
	if tenantID, exists := event.Data["tenant_id"]; exists {
		if tenantIDStr, ok := tenantID.(string); ok && tenantIDStr != "" {
			s.logger.Infof("User %s created in tenant: %s", event.AggregateID, tenantIDStr)

			// Optionally publish tenant user added event
			// return s.PublishTenantUserAddedEvent(ctx, tenantIDStr, "", event.AggregateID, "system")
		}
	}

	return nil
}

// handleUserDeactivatedEvent processes user deactivated events
func (s *TenantEventService) handleUserDeactivatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing user deactivated event for user: %s", event.AggregateID)

	// Remove user from all tenants or handle tenant-specific deactivation
	// This could involve cleanup of tenant-specific resources, permissions, etc.

	return nil
}

// SubscribeToTenantEvents sets up event subscribers for tenant-specific events
func (s *TenantEventService) SubscribeToTenantEvents(ctx context.Context) error {
	// Subscribe to tenant events
	disposable, err := s.eventSubscriber.SubscribeToAllEvents(s.handleTenantEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to tenant events: %w", err)
	}

	s.logger.Infof("Successfully subscribed to tenant events")

	// Store disposable for cleanup if needed
	_ = disposable

	return nil
}

// handleTenantEvent processes incoming tenant events
func (s *TenantEventService) handleTenantEvent(ctx context.Context, event *domain.DomainEvent) error {
	if event.Source != "tenant" {
		return nil // Only process tenant events
	}

	s.logger.Infof("Processing tenant event: %s of type %s", event.ID, event.Type)

	switch event.Type {
	case domain.EventTypeTenantCreated:
		return s.handleTenantCreatedEvent(ctx, event)
	case domain.EventTypeTenantUpdated:
		return s.handleTenantUpdatedEvent(ctx, event)
	case domain.EventTypeTenantDeactivated:
		return s.handleTenantDeactivatedEvent(ctx, event)
	case domain.EventTypeTenantUserAdded:
		return s.handleTenantUserAddedEvent(ctx, event)
	case domain.EventTypeTenantUserRemoved:
		return s.handleTenantUserRemovedEvent(ctx, event)
	}

	return nil
}

// Event handlers for tenant-specific events
func (s *TenantEventService) handleTenantCreatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing tenant created event for tenant: %s", event.AggregateID)
	// Handle post-creation tasks like setting up default configurations
	return nil
}

func (s *TenantEventService) handleTenantUpdatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing tenant updated event for tenant: %s", event.AggregateID)
	// Handle configuration updates, cache invalidation, etc.
	return nil
}

func (s *TenantEventService) handleTenantDeactivatedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing tenant deactivated event for tenant: %s", event.AggregateID)
	// Handle cleanup of tenant resources, user sessions, etc.
	return nil
}

func (s *TenantEventService) handleTenantUserAddedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing tenant user added event for tenant: %s", event.AggregateID)
	// Handle user onboarding tasks, permission setup, etc.
	return nil
}

func (s *TenantEventService) handleTenantUserRemovedEvent(ctx context.Context, event *domain.DomainEvent) error {
	s.logger.Infof("Processing tenant user removed event for tenant: %s", event.AggregateID)
	// Handle user cleanup tasks, permission revocation, etc.
	return nil
}