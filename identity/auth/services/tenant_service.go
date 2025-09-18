package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

type tenantService struct {
	authModule *auth.Module
}

func NewTenantService(authModule *auth.Module) interfaces.TenantService {
	return &tenantService{
		authModule: authModule,
	}
}

func (s *tenantService) SwitchTenant(ctx context.Context, userID, targetTenantID, currentSessionID string) (*interfaces.SwitchTenantResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var response *interfaces.SwitchTenantResponse

	err = s.authModule.GetUnitOfWork().ExecuteInTransaction(ctx, func(ctx context.Context, uow uow.UnitOfWork) error {
		// Get user
		user, err := uow.Users().GetByID(ctx, userUUID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		// Validate user has access to target tenant
		hasAccess, err := s.validateUserTenantAccess(ctx, uow, userID, targetTenantID)
		if err != nil {
			return fmt.Errorf("failed to validate tenant access: %w", err)
		}

		if !hasAccess {
			s.logSecurityEvent(ctx, uow, userUUID, "tenant_switch_denied", map[string]interface{}{
				"target_tenant_id": targetTenantID,
				"reason":           "access_denied",
			})
			return fmt.Errorf("user does not have access to target tenant")
		}

		// Get current session to validate and update
		currentSession, err := uow.Sessions().GetBySessionID(ctx, currentSessionID)
		if err != nil || currentSession == nil {
			return fmt.Errorf("current session not found")
		}

		// Verify session belongs to user
		if currentSession.UserID != userUUID {
			return fmt.Errorf("session does not belong to user")
		}

		// Create new session with target tenant context
		newSession := &models.Session{
			SessionID:        generateSessionID(),
			UserID:           user.ID,
			RefreshTokenHash: generateRefreshTokenHash(),
			IsActive:         true,
			ExpiresAt:        time.Now().Add(24 * time.Hour),
			DeviceInfo:       currentSession.DeviceInfo,
			IPAddress:        currentSession.IPAddress,
			TenantID:         &targetTenantID,
		}

		createdSession, err := uow.Sessions().Create(ctx, newSession)
		if err != nil {
			return fmt.Errorf("failed to create new session: %w", err)
		}

		// Deactivate old session
		err = uow.Sessions().Deactivate(ctx, currentSessionID)
		if err != nil {
			log.Printf("Warning: failed to deactivate old session: %v", err)
		}

		// Get user tenants for response
		// tenants, err := uow.Audit().GetUserTenants(ctx, userUUID)
		// if err != nil {
		// 	log.Printf("Warning: failed to get user tenants: %v", err)
		// 	tenants = []*models.UserTenant{}
		// }

		// Generate new tokens
		accessToken, refreshToken, expiresAt, err := s.generateTokens(user, createdSession)
		if err != nil {
			return fmt.Errorf("failed to generate tokens: %w", err)
		}

		// Log tenant switch
		s.logSecurityEvent(ctx, uow, userUUID, "tenant_switched", map[string]interface{}{
			"target_tenant_id":    targetTenantID,
			"previous_session_id": currentSessionID,
			"new_session_id":      createdSession.SessionID,
		})

		response = &interfaces.SwitchTenantResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    expiresAt,
			User:         user,
			Session:      createdSession,
			SessionID:    createdSession.SessionID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *tenantService) GetUserTenants(ctx context.Context, userID string) ([]*models.UserTenant, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	tenants, err := s.authModule.Audit().GetUserTenants(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	return tenants, nil
}

func (s *tenantService) ValidateUserTenantAccess(ctx context.Context, userID, tenantID string) (bool, error) {
	return s.validateUserTenantAccess(ctx, nil, userID, tenantID)
}

func (s *tenantService) GetUserRolesInTenant(ctx context.Context, userID, tenantID string) ([]string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user tenants
	tenants, err := s.authModule.Audit().GetUserTenants(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	// Find tenant and return roles
	for _, tenant := range tenants {
		if tenant.TenantID == tenantID && tenant.IsActive {
			return tenant.Roles, nil
		}
	}

	return nil, fmt.Errorf("user does not have access to tenant or tenant not found")
}

// Helper methods

func (s *tenantService) validateUserTenantAccess(ctx context.Context, uow uow.UnitOfWork, userID, tenantID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}

	var tenants []*models.UserTenant

	if uow != nil {
		// Use existing transaction
		tenants, err = uow.Audit().GetUserTenants(ctx, userUUID)
	} else {
		// Create new query
		tenants, err = s.authModule.Audit().GetUserTenants(ctx, userUUID)
	}

	if err != nil {
		return false, fmt.Errorf("failed to get user tenants: %w", err)
	}

	// Check if user has access to the tenant
	for _, tenant := range tenants {
		if tenant.TenantID == tenantID && tenant.IsActive {
			return true, nil
		}
	}

	return false, nil
}

func (s *tenantService) generateTokens(user *models.User, session *models.Session) (string, string, time.Time, error) {
	// Get JWT service from module
	jwtService := s.authModule.GetJWTService()

	// Extract tenant ID from session
	tenantID := ""
	if session.TenantID != nil {
		tenantID = *session.TenantID
	}

	// TODO: Get user roles from database based on user and tenant
	var roles []string

	// Generate tokens
	tokenPair, err := jwtService.GenerateTokens(user, session, tenantID, roles)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresAt, nil
}

func (s *tenantService) logSecurityEvent(ctx context.Context, uow uow.UnitOfWork, userID uuid.UUID, eventType string, eventData map[string]interface{}) {
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
