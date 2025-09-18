package interfaces

import (
	"context"

	"p9e.in/ugcl/identity/auth/models"
)

// EventService handles auth-related event publishing and subscribing
type EventService interface {
	// Event Publishing Methods
	PublishUserLoginEvent(ctx context.Context, user *models.User, sessionID, tenantID, ipAddress string) error
	PublishUserLogoutEvent(ctx context.Context, userID, sessionID, reason string) error
	PublishAccountLockEvent(ctx context.Context, userID, reason, changedBy string) error
	PublishAccountUnlockEvent(ctx context.Context, userID, reason, changedBy string) error
	PublishTwoFactorEnabledEvent(ctx context.Context, userID, method string) error
	PublishTwoFactorDisabledEvent(ctx context.Context, userID string) error
	PublishSessionRevokedEvent(ctx context.Context, userID, sessionID, reason string) error
	PublishFailedLoginEvent(ctx context.Context, identifier, reason, ipAddress string) error

	// Event Subscription Methods
	SubscribeToUserEvents(ctx context.Context) error

	// Cross-Module Communication Methods (for requesting actions from other modules)
	RequestAccountLock(ctx context.Context, userID, reason, requestedBy string) error
	RequestAccountUnlock(ctx context.Context, userID, reason, requestedBy string) error
	RequestAccountStatus(ctx context.Context, userID, requestedBy string) error
}