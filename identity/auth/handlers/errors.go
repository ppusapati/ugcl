package handlers

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Domain errors
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAccountDisabled     = errors.New("account is disabled")
	ErrAccountLocked       = errors.New("account is locked")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrTwoFactorRequired   = errors.New("two factor authentication required")
	ErrInvalidTwoFactor    = errors.New("invalid two factor code")
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrUserNotInTenant     = errors.New("user not in tenant")
	ErrPasswordTooWeak     = errors.New("password does not meet requirements")
	ErrInvalidIdentifier   = errors.New("invalid identifier")
	ErrInvalidRequest      = errors.New("invalid request")
)

// ErrorHandler provides methods to convert domain errors to gRPC status errors
type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

func (e *ErrorHandler) HandleError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrUserNotFound):
		return status.Error(codes.NotFound, "User not found")

	case errors.Is(err, ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "Invalid credentials")

	case errors.Is(err, ErrAccountDisabled):
		return status.Error(codes.PermissionDenied, "Account is disabled")

	case errors.Is(err, ErrAccountLocked):
		return status.Error(codes.PermissionDenied, "Account is locked")

	case errors.Is(err, ErrSessionNotFound):
		return status.Error(codes.NotFound, "Session not found")

	case errors.Is(err, ErrSessionExpired):
		return status.Error(codes.Unauthenticated, "Session expired")

	case errors.Is(err, ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "Invalid token")

	case errors.Is(err, ErrTokenExpired):
		return status.Error(codes.Unauthenticated, "Token expired")

	case errors.Is(err, ErrTwoFactorRequired):
		return status.Error(codes.FailedPrecondition, "Two factor authentication required")

	case errors.Is(err, ErrInvalidTwoFactor):
		return status.Error(codes.InvalidArgument, "Invalid two factor code")

	case errors.Is(err, ErrTenantNotFound):
		return status.Error(codes.NotFound, "Tenant not found")

	case errors.Is(err, ErrUserNotInTenant):
		return status.Error(codes.PermissionDenied, "User not in tenant")

	case errors.Is(err, ErrPasswordTooWeak):
		return status.Error(codes.InvalidArgument, "Password does not meet requirements")

	case errors.Is(err, ErrInvalidIdentifier):
		return status.Error(codes.InvalidArgument, "Invalid identifier")

	case errors.Is(err, ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, "Invalid request")

	default:
		// For unknown errors, return internal error but don't expose details
		return status.Error(codes.Internal, "Internal server error")
	}
}

// Validation helpers
func (e *ErrorHandler) ValidateLoginRequest(req interface{}) error {
	// Add specific validation logic here
	return nil
}

func (e *ErrorHandler) ValidateUserID(userID string) error {
	if userID == "" {
		return ErrInvalidRequest
	}
	return nil
}

func (e *ErrorHandler) ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return ErrInvalidRequest
	}
	return nil
}

func (e *ErrorHandler) ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooWeak
	}
	// Add more password validation rules here
	return nil
}

func (e *ErrorHandler) ValidateTenantID(tenantID string) error {
	if tenantID == "" {
		return ErrInvalidRequest
	}
	return nil
}