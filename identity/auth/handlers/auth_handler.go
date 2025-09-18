package handlers

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	authpb "p9e.in/ugcl/identity/auth/api/v2/auth"
	"p9e.in/ugcl/identity/auth/api/v2/auth/authconnect"
	"p9e.in/ugcl/identity/auth/mappers"
	"p9e.in/ugcl/identity/auth/services/interfaces"
)

type AuthHandler struct {
	authService      interfaces.AuthService
	sessionService   interfaces.SessionService
	jwtService       interfaces.JWTService
	twoFactorService interfaces.TwoFactorService
	eventService     interfaces.EventService
	protobufMapper   *mappers.ProtobufMapper
	errorHandler     *ErrorHandler
}

func NewAuthHandler(
	authService interfaces.AuthService,
	sessionService interfaces.SessionService,
	jwtService interfaces.JWTService,
	twoFactorService interfaces.TwoFactorService,
	eventService interfaces.EventService,
	protobufMapper *mappers.ProtobufMapper,
	errorHandler *ErrorHandler,
) authconnect.AuthServiceHandler {
	return &AuthHandler{
		authService:      authService,
		sessionService:   sessionService,
		jwtService:       jwtService,
		twoFactorService: twoFactorService,
		eventService:     eventService,
		protobufMapper:   protobufMapper,
		errorHandler:     errorHandler,
	}
}

func (h *AuthHandler) Login(ctx context.Context, req *connect.Request[authpb.LoginRequest]) (*connect.Response[authpb.LoginResponse], error) {
	log.Printf("Login request for identifier: %v", req.Msg.Identifier)

	// Extract tenant context from request
	tenantID := h.extractTenantFromContext(ctx, req.Header())
	if req.Msg.TenantId != nil && *req.Msg.TenantId != "" {
		tenantID = *req.Msg.TenantId
	}

	// Extract identifier and type
	identifierType, identifier := h.protobufMapper.ExtractIdentifierFromLoginRequest(req.Msg)
	if identifier == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid identifier"))
	}

	// Create service request
	loginReq := &interfaces.LoginRequest{
		IdentifierType: identifierType,
		Identifier:     identifier,
		Password:       req.Msg.Password,
		TenantID:       tenantID,
		DeviceInfo:     func() string { if req.Msg.DeviceInfo != nil { return *req.Msg.DeviceInfo }; return "" }(),
		IPAddress:      func() string { if req.Msg.IpAddress != nil { return *req.Msg.IpAddress }; return "" }(),
		RememberMe:     req.Msg.RememberMe,
	}

	// Call auth service
	loginResp, err := h.authService.Login(ctx, loginReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service response to protobuf response
	response := h.protobufMapper.BuildLoginResponse(loginResp)

	return connect.NewResponse(response), nil
}

func (h *AuthHandler) LoginWithOTP(ctx context.Context, req *connect.Request[authpb.LoginOTPRequest]) (*connect.Response[authpb.LoginResponse], error) {
	log.Printf("LoginWithOTP request for identifier: %v", req.Msg.Identifier)

	// Extract tenant context
	tenantID := h.extractTenantFromContext(ctx, req.Header())
	if req.Msg.TenantId != nil && *req.Msg.TenantId != "" {
		tenantID = *req.Msg.TenantId
	}

	// Extract identifier and type
	identifierType, identifier := h.protobufMapper.ExtractIdentifierFromOTPRequest(req.Msg)
	if identifier == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid identifier"))
	}

	// Create service request
	otpReq := &interfaces.LoginOTPRequest{
		IdentifierType: identifierType,
		Identifier:     identifier,
		OTPCode:        req.Msg.OtpCode,
		TenantID:       tenantID,
		DeviceInfo:     func() string { if req.Msg.DeviceInfo != nil { return *req.Msg.DeviceInfo }; return "" }(),
		IPAddress:      func() string { if req.Msg.IpAddress != nil { return *req.Msg.IpAddress }; return "" }(),
	}

	// Call auth service
	loginResp, err := h.authService.LoginWithOTP(ctx, otpReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert service response to protobuf response
	response := h.protobufMapper.BuildLoginResponse(loginResp)

	return connect.NewResponse(response), nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *connect.Request[authpb.LogoutRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Logout request for session: %s", req.Msg.SessionId)

	if req.Msg.SessionId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("session ID is required"))
	}

	err := h.authService.Logout(ctx, req.Msg.SessionId, func() string { if req.Msg.DeviceInfo != nil { return *req.Msg.DeviceInfo }; return "" }())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Helper method to extract tenant context from request
func (h *AuthHandler) extractTenantFromContext(ctx context.Context, headers map[string][]string) string {
	// Check for tenant ID in headers
	if tenantIDs := headers["X-Tenant-ID"]; len(tenantIDs) > 0 && tenantIDs[0] != "" {
		return tenantIDs[0]
	}

	// Check for tenant ID in subdomain (if using subdomain-based tenancy)
	if hosts := headers["Host"]; len(hosts) > 0 && hosts[0] != "" {
		// Extract subdomain logic can be added here
		// For now, we'll just return empty string
	}

	// TODO: Extract tenant from JWT token if present
	if authHeaders := headers["Authorization"]; len(authHeaders) > 0 && authHeaders[0] != "" {
		authHeader := authHeaders[0]
		// Parse Bearer token and extract tenant from claims
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			if claims, err := h.jwtService.ValidateToken(token); err == nil {
				if claims.TenantID != "" {
					return claims.TenantID
				}
			}
		}
	}

	return ""
}

func (h *AuthHandler) LogoutAll(ctx context.Context, req *connect.Request[authpb.LogoutAllRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("LogoutAll request for user: %s", req.Msg.UserId)

	err := h.authService.LogoutAll(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *connect.Request[authpb.RefreshTokenRequest]) (*connect.Response[authpb.RefreshTokenResponse], error) {
	log.Printf("RefreshToken request")

	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("refresh token is required"))
	}

	resp, err := h.authService.RefreshToken(ctx, req.Msg.RefreshToken, func() string { if req.Msg.DeviceInfo != nil { return *req.Msg.DeviceInfo }; return "" }())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authpb.RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    timestamppb.New(resp.ExpiresAt),
	}), nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *connect.Request[authpb.ValidateTokenRequest]) (*connect.Response[authpb.ValidateTokenResponse], error) {
	log.Printf("ValidateToken request")

	if req.Msg.Token == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("token is required"))
	}

	resp, err := h.authService.ValidateToken(ctx, req.Msg.Token, func() string { if req.Msg.RequiredPermission != nil { return *req.Msg.RequiredPermission }; return "" }(), func() string { if req.Msg.Resource != nil { return *req.Msg.Resource }; return "" }())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &authpb.ValidateTokenResponse{
		Valid:     resp.Valid,
		ExpiresAt: timestamppb.New(resp.ExpiresAt),
	}

	if resp.Valid && resp.User != nil {
		response.User = &authpb.UserSession{
			UserId:   resp.User.UserID,
			Username: resp.User.Username,
			Email:    resp.User.Email,
			TenantId: &resp.TenantID,
		}
		response.Permissions = resp.Permissions
	}

	return connect.NewResponse(response), nil
}

func (h *AuthHandler) RevokeToken(ctx context.Context, req *connect.Request[authpb.RevokeTokenRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("RevokeToken request for token type: %v", req.Msg.TokenType)

	if req.Msg.Token == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("token is required"))
	}

	tokenType := "access"
	switch req.Msg.TokenType {
	case authpb.TokenType_TOKEN_TYPE_ACCESS:
		tokenType = "access"
	case authpb.TokenType_TOKEN_TYPE_REFRESH:
		tokenType = "refresh"
	}

	err := h.authService.RevokeToken(ctx, req.Msg.Token, tokenType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Two-Factor Authentication
func (h *AuthHandler) EnableTwoFactor(ctx context.Context, req *connect.Request[authpb.EnableTwoFactorRequest]) (*connect.Response[authpb.EnableTwoFactorResponse], error) {
	log.Printf("EnableTwoFactor request for user: %s, method: %v", req.Msg.UserId, req.Msg.Method)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	methodStr := req.Msg.Method.String()
	if methodStr == "TWO_FACTOR_METHOD_UNSPECIFIED" {
		methodStr = "totp" // Default to TOTP
	} else {
		// Convert enum to lowercase string
		switch req.Msg.Method {
		case authpb.TwoFactorMethod_TWO_FACTOR_METHOD_TOTP:
			methodStr = "totp"
		case authpb.TwoFactorMethod_TWO_FACTOR_METHOD_SMS:
			methodStr = "sms"
		case authpb.TwoFactorMethod_TWO_FACTOR_METHOD_EMAIL:
			methodStr = "email"
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported two-factor method"))
		}
	}

	resp, err := h.twoFactorService.EnableTwoFactor(ctx, req.Msg.UserId, methodStr)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authpb.EnableTwoFactorResponse{
		Secret:      resp.Secret,
		QrCode:      resp.QRCode,
		BackupCodes: resp.BackupCodes,
	}), nil
}

func (h *AuthHandler) DisableTwoFactor(ctx context.Context, req *connect.Request[authpb.DisableTwoFactorRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("DisableTwoFactor request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	if req.Msg.VerificationCode == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("verification code is required"))
	}

	err := h.twoFactorService.DisableTwoFactor(ctx, req.Msg.UserId, req.Msg.VerificationCode)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *AuthHandler) VerifyTwoFactor(ctx context.Context, req *connect.Request[authpb.VerifyTwoFactorRequest]) (*connect.Response[authpb.VerifyTwoFactorResponse], error) {
	log.Printf("VerifyTwoFactor request for user: %s, method: %v", req.Msg.UserId, req.Msg.Method)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	if req.Msg.Code == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("verification code is required"))
	}

	methodStr := "totp" // Default to TOTP
	switch req.Msg.Method {
	case authpb.TwoFactorMethod_TWO_FACTOR_METHOD_TOTP:
		methodStr = "totp"
	case authpb.TwoFactorMethod_TWO_FACTOR_METHOD_SMS:
		methodStr = "sms"
	case authpb.TwoFactorMethod_TWO_FACTOR_METHOD_EMAIL:
		methodStr = "email"
	case authpb.TwoFactorMethod_TWO_FACTOR_METHOD_BACKUP_CODE:
		methodStr = "backup_code"
	}

	err := h.twoFactorService.VerifyTwoFactor(ctx, req.Msg.UserId, req.Msg.Code, methodStr)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authpb.VerifyTwoFactorResponse{
		Valid:   true,
		Message: "Two-factor authentication verified successfully",
	}), nil
}

func (h *AuthHandler) GenerateBackupCodes(ctx context.Context, req *connect.Request[authpb.GenerateBackupCodesRequest]) (*connect.Response[authpb.GenerateBackupCodesResponse], error) {
	log.Printf("GenerateBackupCodes request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	backupCodes, err := h.twoFactorService.GenerateBackupCodes(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authpb.GenerateBackupCodesResponse{
		BackupCodes: backupCodes,
	}), nil
}

// Account Security
func (h *AuthHandler) LockAccount(ctx context.Context, req *connect.Request[authpb.LockAccountRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("LockAccount request for user: %s, reason: %s", req.Msg.UserId, req.Msg.Reason)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	// Request account lock via events
	if h.eventService != nil {
		err := h.eventService.RequestAccountLock(ctx, req.Msg.UserId, req.Msg.Reason, "auth_service")
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to request account lock: %w", err))
		}
	} else {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("account locking not available - event service not configured"))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *AuthHandler) UnlockAccount(ctx context.Context, req *connect.Request[authpb.UnlockAccountRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("UnlockAccount request for user: %s, reason: %s", req.Msg.UserId, req.Msg.Reason)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	// Request account unlock via events
	if h.eventService != nil {
		err := h.eventService.RequestAccountUnlock(ctx, req.Msg.UserId, req.Msg.Reason, "auth_service")
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to request account unlock: %w", err))
		}
	} else {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("account unlocking not available - event service not configured"))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *AuthHandler) GetAccountStatus(ctx context.Context, req *connect.Request[authpb.GetAccountStatusRequest]) (*connect.Response[authpb.GetAccountStatusResponse], error) {
	log.Printf("GetAccountStatus request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	// Request account status via events
	if h.eventService != nil {
		err := h.eventService.RequestAccountStatus(ctx, req.Msg.UserId, "auth_service")
		if err != nil {
			log.Printf("Warning: failed to request account status via events: %v", err)
		}
	}

	return connect.NewResponse(&authpb.GetAccountStatusResponse{
		Status:              authpb.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		FailedLoginAttempts: 0,
		LastSuccessfulLogin: timestamppb.Now(),
	}), nil
}

// Additional Token Management
func (h *AuthHandler) RevokeAllTokens(ctx context.Context, req *connect.Request[authpb.RevokeAllTokensRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("RevokeAllTokens request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	// Revoke all sessions for the user
	err := h.authService.LogoutAll(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *AuthHandler) GetActiveSessions(ctx context.Context, req *connect.Request[authpb.GetActiveSessionsRequest]) (*connect.Response[authpb.GetActiveSessionsResponse], error) {
	log.Printf("GetActiveSessions request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	// Get active sessions from service
	sessions, err := h.sessionService.GetActiveSessions(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to protobuf sessions
	pbSessions := make([]*authpb.Session, len(sessions))
	for i, session := range sessions {
		pbSessions[i] = h.protobufMapper.ToProtobufSession(session)
	}

	return connect.NewResponse(&authpb.GetActiveSessionsResponse{
		Sessions: pbSessions,
	}), nil
}

func (h *AuthHandler) RevokeSession(ctx context.Context, req *connect.Request[authpb.RevokeSessionRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("RevokeSession request for session: %s, user: %s", req.Msg.SessionId, req.Msg.UserId)

	if req.Msg.SessionId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("session ID is required"))
	}
	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	err := h.sessionService.RevokeSession(ctx, req.Msg.SessionId, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *AuthHandler) RevokeAllSessions(ctx context.Context, req *connect.Request[authpb.RevokeAllSessionsRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("RevokeAllSessions request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	err := h.sessionService.RevokeAllSessions(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Security Events
func (h *AuthHandler) GetSecurityEvents(ctx context.Context, req *connect.Request[authpb.GetSecurityEventsRequest]) (*connect.Response[authpb.GetSecurityEventsResponse], error) {
	log.Printf("GetSecurityEvents request for user: %s", req.Msg.UserId)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}

	// TODO: Implement security events retrieval from audit service
	// events, err := h.serviceContainer.GetAuditService().GetSecurityEvents(ctx, req.Msg.UserId, pagination)
	// if err != nil {
	//     return nil, connect.NewError(connect.CodeInternal, err)
	// }

	return connect.NewResponse(&authpb.GetSecurityEventsResponse{
		Events:     []*authpb.SecurityEvent{},
		TotalCount: 0,
	}), nil
}

func (h *AuthHandler) RecordSecurityEvent(ctx context.Context, req *connect.Request[authpb.RecordSecurityEventRequest]) (*connect.Response[emptypb.Empty], error) {
	log.Printf("RecordSecurityEvent request for user: %s, event: %v", req.Msg.UserId, req.Msg.EventType)

	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user ID is required"))
	}
	if req.Msg.EventType == authpb.SecurityEventType_SECURITY_EVENT_TYPE_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("event type is required"))
	}

	// TODO: Implement security event recording through audit service
	// err := h.serviceContainer.GetAuditService().LogSecurityEvent(ctx, securityEvent)
	// if err != nil {
	//     return nil, connect.NewError(connect.CodeInternal, err)
	// }

	return connect.NewResponse(&emptypb.Empty{}), nil
}
