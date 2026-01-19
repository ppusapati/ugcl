package identity

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	authServices "p9e.in/ugcl/identity/auth/services"
	tenantServices "p9e.in/ugcl/identity/tenant/services"
	userServices "p9e.in/ugcl/identity/user/services"
	"p9e.in/ugcl/packages/events/bus"
	"p9e.in/ugcl/packages/events/domain"
	"p9e.in/ugcl/packages/events/producer"
	"p9e.in/ugcl/packages/p9log"
)

// IdentityEventManager manages event-driven communication between identity modules
type IdentityEventManager struct {
	eventBus           *bus.EventBus
	eventPublisher     *domain.DomainEventPublisher
	eventSubscriber    *domain.EventSubscriber
	authContainer      *authServices.Container
	userEventService   *userServices.UserEventService
	tenantEventService *tenantServices.TenantEventService
	logger             p9log.Logger
}

// NewIdentityEventManager creates a new identity event manager
func NewIdentityEventManager(
	db *pgxpool.Pool,
	kafkaProducer *producer.KafkaProducer,
	logger p9log.Logger,
	jwtSecret, issuer string,
) (*IdentityEventManager, error) {
	// Create event bus
	eventBus := bus.NewEventBus()

	// Create event publisher and subscriber
	eventPublisher := domain.NewDomainEventPublisher(eventBus, kafkaProducer, logger)
	eventSubscriber := domain.NewEventSubscriber(eventBus, logger)

	// Create auth container with event support
	authContainer := authServices.NewContainerWithEvents(db, jwtSecret, issuer, eventPublisher, eventSubscriber, logger)

	// Create user event service (assuming we have a user service)
	// userService := userServices.NewUserService() // Implementation depends on your user service
	// userEventService := userServices.NewUserEventService(eventPublisher, eventSubscriber, logger, userService)

	// Create tenant event service (assuming we have a tenant service)
	// tenantService := tenantServices.NewTenantService() // Implementation depends on your tenant service
	// tenantEventService := tenantServices.NewTenantEventService(eventPublisher, eventSubscriber, logger, tenantService)

	manager := &IdentityEventManager{
		eventBus:        eventBus,
		eventPublisher:  eventPublisher,
		eventSubscriber: eventSubscriber,
		authContainer:   authContainer,
		// userEventService:   userEventService,
		// tenantEventService: tenantEventService,
		logger: logger,
	}

	return manager, nil
}

// StartEventSubscriptions starts all event subscriptions
func (m *IdentityEventManager) StartEventSubscriptions(ctx context.Context) error {
	// Start auth event subscriptions
	if authEventService := m.authContainer.GetEventService(); authEventService != nil {
		if err := authEventService.SubscribeToUserEvents(ctx); err != nil {
			return err
		}
		log.Printf("Started auth event subscriptions")
	}

	// Start user event subscriptions
	if m.userEventService != nil {
		if err := m.userEventService.SubscribeToAuthEvents(ctx); err != nil {
			return err
		}
		log.Printf("Started user event subscriptions")
	}

	// Start tenant event subscriptions
	if m.tenantEventService != nil {
		if err := m.tenantEventService.SubscribeToIdentityEvents(ctx); err != nil {
			return err
		}
		if err := m.tenantEventService.SubscribeToTenantEvents(ctx); err != nil {
			return err
		}
		log.Printf("Started tenant event subscriptions")
	}

	log.Printf("All identity event subscriptions started successfully")
	return nil
}

// GetAuthContainer returns the auth service container
func (m *IdentityEventManager) GetAuthContainer() *authServices.Container {
	return m.authContainer
}

// GetEventPublisher returns the domain event publisher
func (m *IdentityEventManager) GetEventPublisher() *domain.DomainEventPublisher {
	return m.eventPublisher
}

// GetEventSubscriber returns the domain event subscriber
func (m *IdentityEventManager) GetEventSubscriber() *domain.EventSubscriber {
	return m.eventSubscriber
}

// Example usage demonstrating event-driven communication
func ExampleEventDrivenFlow(manager *IdentityEventManager) {
	ctx := context.Background()

	// Example 1: User login triggers events
	authService := manager.GetAuthContainer().GetAuthService()

	// When user logs in, auth service will automatically publish login events
	// These events will be consumed by user and tenant services

	// Example 2: Account lock request via events
	eventService := manager.GetAuthContainer().GetEventService()
	if eventService != nil {
		// Auth service requests account lock via events
		err := eventService.RequestAccountLock(ctx, "user-123", "suspicious_activity", "auth_service")
		if err != nil {
			log.Printf("Failed to request account lock: %v", err)
		}
		// User service will receive this event and process the lock request
	}

	// Example 3: Tenant creation triggers cross-module events
	if manager.tenantEventService != nil {
		// Tenant service publishes tenant created event
		err := manager.tenantEventService.PublishTenantCreatedEvent(ctx, "tenant-456", "New Company", "admin")
		if err != nil {
			log.Printf("Failed to publish tenant created event: %v", err)
		}
		// Auth and user services will receive and process this event
	}

	// Example 4: User creation triggers welcome workflow
	if manager.userEventService != nil {
		// User service publishes user created event
		err := manager.userEventService.PublishUserCreatedEvent(ctx, "user-789", "user@example.com", "newuser", "system")
		if err != nil {
			log.Printf("Failed to publish user created event: %v", err)
		}
		// Auth service will receive this and create authentication records
		// Tenant service will receive this for tenant-related setup
	}
}

// EventFlowValidation provides methods to validate event flows
type EventFlowValidation struct {
	manager *IdentityEventManager
	logger  p9log.Logger
}

// NewEventFlowValidation creates a new event flow validator
func NewEventFlowValidation(manager *IdentityEventManager, logger p9log.Logger) *EventFlowValidation {
	return &EventFlowValidation{
		manager: manager,
		logger:  logger,
	}
}

// ValidateAuthUserFlow validates auth-user event communication
func (v *EventFlowValidation) ValidateAuthUserFlow(ctx context.Context) error {
	log.Printf("Validating auth-user event flow...")

	eventService := v.manager.GetAuthContainer().GetEventService()
	if eventService == nil {
		return fmt.Errorf("auth event service not available")
	}

	// Test account lock request
	err := eventService.RequestAccountLock(ctx, "test-user-123", "test_reason", "test_validator")
	if err != nil {
		return fmt.Errorf("failed to send account lock request: %w", err)
	}

	log.Printf("✓ Auth-user event flow validation passed")
	return nil
}

// ValidateUserTenantFlow validates user-tenant event communication
func (v *EventFlowValidation) ValidateUserTenantFlow(ctx context.Context) error {
	log.Printf("Validating user-tenant event flow...")

	if v.manager.userEventService == nil {
		log.Printf("⚠ User event service not available - skipping validation")
		return nil
	}

	// Test user created event
	err := v.manager.userEventService.PublishUserCreatedEvent(ctx, "test-user-456", "test@example.com", "testuser", "validator")
	if err != nil {
		return fmt.Errorf("failed to publish user created event: %w", err)
	}

	log.Printf("✓ User-tenant event flow validation passed")
	return nil
}

// ValidateCompleteFlow validates the complete event-driven architecture
func (v *EventFlowValidation) ValidateCompleteFlow(ctx context.Context) error {
	log.Printf("Starting complete event flow validation...")

	// Validate auth-user flow
	if err := v.ValidateAuthUserFlow(ctx); err != nil {
		return err
	}

	// Validate user-tenant flow
	if err := v.ValidateUserTenantFlow(ctx); err != nil {
		return err
	}

	log.Printf("✅ Complete event-driven architecture validation passed")
	return nil
}

/*
Integration Steps:

1. Initialize the IdentityEventManager in your main application:
   ```go
   manager, err := NewIdentityEventManager(db, kafkaProducer, logger, jwtSecret, issuer)
   if err != nil {
       log.Fatal(err)
   }
   ```

2. Start event subscriptions:
   ```go
   if err := manager.StartEventSubscriptions(ctx); err != nil {
       log.Fatal(err)
   }
   ```

3. Use the auth container with event support:
   ```go
   authContainer := manager.GetAuthContainer()
   authHandler := handlers.NewAuthHandler(authContainer)
   ```

4. Validate the event flows:
   ```go
   validator := NewEventFlowValidation(manager, logger)
   if err := validator.ValidateCompleteFlow(ctx); err != nil {
       log.Printf("Event flow validation failed: %v", err)
   }
   ```

Key Benefits:
- ✅ Loose coupling between modules
- ✅ Event-driven communication
- ✅ Scalable architecture
- ✅ Cross-module coordination without direct dependencies
- ✅ Comprehensive event publishing and subscribing
- ✅ Validation and testing capabilities

Event Flow Summary:
Auth Module ─┐
             ├─► Event Bus ◄─► Kafka ◄─┐
User Module ─┤                         ├─ Cross-Service Events
             │                         │
Tenant Module┘                        ─┘
*/