# Authentication & Session Management Module

## 1. Module Overview

The Authentication (Auth) module handles user authentication, session management, two-factor authentication, security monitoring, and token management for the UGCL platform. It provides secure login/logout flows, session tracking, and comprehensive security event logging.

**Purpose:** Secure authentication and authorization with multi-tenant support, enhanced session management with entity and organizational context, and comprehensive security monitoring.

**Key Features:**
- Username/email/phone login with password
- OTP-based login
- Two-factor authentication (TOTP, SMS, Email, Backup Codes)
- JWT-based access and refresh tokens
- Session management with device tracking
- Account security (lock/unlock)
- Security event logging and monitoring
- Token revocation and lifecycle management
- Entity and organizational context in sessions
- Multi-tenant authentication

## 2. Architecture

### Module Structure
```
identity/auth/
├── proto/
│   └── auth.proto           # Service and message definitions
├── db/
│   ├── schema/
│   │   └── schema.sql      # Sessions, tokens, security events
│   └── generated/          # SQLC generated code
├── repository/             # Data access layer
│   ├── session_repository.go
│   ├── token_repository.go
│   └── security_event_repository.go
├── services/               # Business logic layer
│   ├── auth_service.go
│   ├── session_service.go
│   ├── token_service.go
│   └── security_service.go
├── handlers/               # gRPC handlers
│   └── auth_handler.go
├── mappers/                # DTO mappers
│   └── auth_mappers.go
├── models/                 # Domain models
│   └── session.go
├── uow/                    # Unit of work
└── module.go               # FX module definition
```

### Database Schema

**Tables:**
- `sessions` - Active user sessions with device info
- `tokens` - JWT tokens (access, refresh, reset)
- `security_events` - Audit log of security events
- `two_factor_backups` - Backup codes for 2FA

**Key Relationships:**
- Session → User (N:1)
- Session → Tenant (N:1)
- Token → User (N:1)
- SecurityEvent → User (N:1)

### Key Dependencies
- `identity/user` - User account validation
- `identity/tenant` - Tenant context
- `identity/entity` - Entity resolution for sessions
- JWT library for token generation
- TOTP library for two-factor auth
- bcrypt for password verification

## 3. Quick Start

### User Login

```go
import (
    authv2 "p9e.in/ugcl/identity/auth/api/v2/auth"
    "connectrpc.com/connect"
)

client := authv2.NewAuthServiceClient(httpClient, baseURL)

// Login with username and password
loginResp, err := client.Login(ctx, connect.NewRequest(&authv2.LoginRequest{
    Identifier: &authv2.LoginRequest_Username{
        Username: "john.doe",
    },
    Password:   "SecurePass123!",
    TenantId:   wrapperspb.String("tenant-uuid"),
    DeviceInfo: wrapperspb.String("Mozilla/5.0 ..."),
    IpAddress:  wrapperspb.String("192.168.1.100"),
    RememberMe: true,
}))

accessToken := loginResp.Msg.AccessToken
refreshToken := loginResp.Msg.RefreshToken
sessionID := loginResp.Msg.SessionId
userSession := loginResp.Msg.User

// Check if 2FA required
if loginResp.Msg.RequiresTwoFactor {
    // Prompt for 2FA code
    verify2FA, err := client.VerifyTwoFactor(ctx, connect.NewRequest(&authv2.VerifyTwoFactorRequest{
        UserId: userSession.UserId,
        Code:   "123456",
        Method: authv2.TwoFactorMethod_TWO_FACTOR_METHOD_TOTP,
    }))
}
```

### Common Operations

**Refresh token:**
```go
refreshed, err := client.RefreshToken(ctx, connect.NewRequest(&authv2.RefreshTokenRequest{
    RefreshToken: refreshToken,
    DeviceInfo:   wrapperspb.String("Mozilla/5.0 ..."),
}))

newAccessToken := refreshed.Msg.AccessToken
newRefreshToken := refreshed.Msg.RefreshToken
```

**Validate token:**
```go
validation, err := client.ValidateToken(ctx, connect.NewRequest(&authv2.ValidateTokenRequest{
    Token:              accessToken,
    RequiredPermission: wrapperspb.String("documents.read"),
    Resource:           wrapperspb.String("document-uuid"),
}))

if validation.Msg.Valid {
    user := validation.Msg.User
    permissions := validation.Msg.Permissions
}
```

**Logout:**
```go
err := client.Logout(ctx, connect.NewRequest(&authv2.LogoutRequest{
    SessionId:  sessionID,
    DeviceInfo: wrapperspb.String("Mozilla/5.0 ..."),
}))
```

## 4. API Reference

### Authentication
- `Login` - Authenticate with username/email/phone and password
- `LoginWithOTP` - Authenticate using OTP code
- `Logout` - End current session
- `LogoutAll` - End all user sessions

### Two-Factor Authentication
- `EnableTwoFactor` - Enable 2FA for user
- `DisableTwoFactor` - Disable 2FA for user
- `VerifyTwoFactor` - Verify 2FA code
- `GenerateBackupCodes` - Generate 2FA backup codes

### Account Security
- `LockAccount` - Lock user account
- `UnlockAccount` - Unlock user account
- `GetAccountStatus` - Get account security status

### Token Management
- `RefreshToken` - Refresh access token using refresh token
- `ValidateToken` - Validate token and check permissions
- `RevokeToken` - Revoke specific token
- `RevokeAllTokens` - Revoke all user tokens

### Session Management
- `GetActiveSessions` - Get all active user sessions
- `RevokeSession` - Revoke specific session
- `RevokeAllSessions` - Revoke all user sessions

### Security Events
- `GetSecurityEvents` - Query security event log
- `RecordSecurityEvent` - Log security event

## 5. Database Schema

### Table: sessions

```sql
- session_id (UUID, PK)
- user_id (UUID, FK → users)
- tenant_id (UUID, optional)
- device_info (TEXT)
- ip_address (VARCHAR)
- created_at (TIMESTAMP)
- last_accessed (TIMESTAMP)
- expires_at (TIMESTAMP)
- is_active (BOOLEAN)
```

### Table: tokens

```sql
- token_id (UUID, PK)
- user_id (UUID, FK → users)
- token_type (ENUM) -- ACCESS, REFRESH, RESET
- token_hash (VARCHAR)
- expires_at (TIMESTAMP)
- revoked_at (TIMESTAMP)
- created_at (TIMESTAMP)
```

### Table: security_events

```sql
- event_id (UUID, PK)
- user_id (UUID, FK → users)
- event_type (ENUM) -- LOGIN_SUCCESS, LOGIN_FAILURE, PASSWORD_CHANGE, etc.
- description (TEXT)
- ip_address (VARCHAR)
- device_info (TEXT)
- tenant_id (UUID, optional)
- timestamp (TIMESTAMP)
```

### Indexes
- `idx_sessions_user_id` - User session lookup
- `idx_sessions_expires_at` - Session cleanup
- `idx_tokens_user_id` - Token lookup
- `idx_security_events_user_id` - Event history
- `idx_security_events_timestamp` - Time-based queries

## 6. Configuration

### Environment Variables
```env
# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=7d
JWT_ISSUER=ugcl-platform

# Session Configuration
SESSION_TIMEOUT=30m
SESSION_ABSOLUTE_TIMEOUT=24h
MAX_SESSIONS_PER_USER=5

# Security Configuration
MAX_LOGIN_ATTEMPTS=5
ACCOUNT_LOCKOUT_DURATION=30m
PASSWORD_RESET_TOKEN_EXPIRY=1h

# Two-Factor Authentication
TOTP_ISSUER=UGCL
TOTP_PERIOD=30
BACKUP_CODES_COUNT=10
```

### Module-Specific Settings
- Token signing algorithm (HS256, RS256)
- Session storage backend (database, Redis)
- 2FA methods enabled
- IP whitelist/blacklist

## 7. Examples

### Complete Authentication Flow with 2FA

```go
// 1. Initial login
loginResp, err := client.Login(ctx, connect.NewRequest(&authv2.LoginRequest{
    Identifier: &authv2.LoginRequest_Email{
        Email: "user@ugcl.com",
    },
    Password:   "Password123!",
    TenantId:   wrapperspb.String(tenantID),
    IpAddress:  wrapperspb.String("192.168.1.100"),
    DeviceInfo: wrapperspb.String("Chrome 120.0"),
}))

if loginResp.Msg.RequiresTwoFactor {
    fmt.Println("2FA required")

    // 2. User enters TOTP code from authenticator app
    verify, err := client.VerifyTwoFactor(ctx, connect.NewRequest(&authv2.VerifyTwoFactorRequest{
        UserId: loginResp.Msg.User.UserId,
        Code:   userProvidedCode,
        Method: authv2.TwoFactorMethod_TWO_FACTOR_METHOD_TOTP,
    }))

    if !verify.Msg.Valid {
        fmt.Println("Invalid 2FA code")
        return
    }
}

// 3. Store tokens securely
accessToken := loginResp.Msg.AccessToken
refreshToken := loginResp.Msg.RefreshToken
sessionID := loginResp.Msg.SessionId

// 4. Use access token for API calls
// When access token expires, refresh it
refreshed, err := client.RefreshToken(ctx, connect.NewRequest(&authv2.RefreshTokenRequest{
    RefreshToken: refreshToken,
}))
```

### Enable Two-Factor Authentication

```go
// 1. Enable 2FA
enable2FA, err := client.EnableTwoFactor(ctx, connect.NewRequest(&authv2.EnableTwoFactorRequest{
    UserId: userID,
    Method: authv2.TwoFactorMethod_TWO_FACTOR_METHOD_TOTP,
}))

// 2. Show QR code to user
qrCode := enable2FA.Msg.QrCode
secret := enable2FA.Msg.Secret
backupCodes := enable2FA.Msg.BackupCodes

fmt.Println("Scan this QR code with your authenticator app:")
// Display qrCode image
fmt.Printf("Secret: %s\n", secret)
fmt.Println("Backup codes (save these securely):")
for i, code := range backupCodes {
    fmt.Printf("%d. %s\n", i+1, code)
}

// 3. User verifies setup by entering first code
verify, err := client.VerifyTwoFactor(ctx, connect.NewRequest(&authv2.VerifyTwoFactorRequest{
    UserId: userID,
    Code:   userEnteredCode,
    Method: authv2.TwoFactorMethod_TWO_FACTOR_METHOD_TOTP,
}))

if verify.Msg.Valid {
    fmt.Println("2FA successfully enabled!")
}
```

### Session Management

```go
// Get all active sessions for user
sessions, err := client.GetActiveSessions(ctx, connect.NewRequest(&authv2.GetActiveSessionsRequest{
    UserId: userID,
}))

fmt.Printf("Active sessions: %d\n", len(sessions.Msg.Sessions))
for _, session := range sessions.Msg.Sessions {
    fmt.Printf("Session: %s\n", session.SessionId)
    fmt.Printf("  Device: %s\n", session.GetDeviceInfo())
    fmt.Printf("  IP: %s\n", session.GetIpAddress())
    fmt.Printf("  Created: %s\n", session.CreatedAt.AsTime())
    fmt.Printf("  Last accessed: %s\n", session.GetLastAccessed().AsTime())
}

// Revoke specific session (e.g., user saw unknown device)
err = client.RevokeSession(ctx, connect.NewRequest(&authv2.RevokeSessionRequest{
    SessionId: suspiciousSessionID,
    UserId:    userID,
}))

// Or revoke all sessions except current (e.g., password changed)
err = client.RevokeAllSessions(ctx, connect.NewRequest(&authv2.RevokeAllSessionsRequest{
    UserId:          userID,
    ExceptSessionId: wrapperspb.String(currentSessionID),
}))
```

### Security Event Monitoring

```go
// Get recent security events
events, err := client.GetSecurityEvents(ctx, connect.NewRequest(&authv2.GetSecurityEventsRequest{
    UserId:    userID,
    FromDate:  timestamppb.New(time.Now().Add(-7*24*time.Hour)),
    ToDate:    timestamppb.Now(),
    EventTypes: []authv2.SecurityEventType{
        authv2.SecurityEventType_SECURITY_EVENT_TYPE_LOGIN_FAILURE,
        authv2.SecurityEventType_SECURITY_EVENT_TYPE_SUSPICIOUS_ACTIVITY,
    },
    PageSize: 100,
}))

fmt.Println("Recent security events:")
for _, event := range events.Msg.Events {
    fmt.Printf("[%s] %s: %s (IP: %s)\n",
        event.Timestamp.AsTime().Format("2006-01-02 15:04:05"),
        event.EventType.String(),
        event.Description,
        event.GetIpAddress())
}
```

## 8. Integration

### With Other Modules

**Identity/User Module:**
- Validates user credentials during login
- Updates last_login timestamp
- Manages password hashes

**Identity/Tenant Module:**
- Validates tenant context
- Includes tenant info in session
- Tenant-scoped authentication

**Identity/Entity Module:**
- Resolves entity for user
- Includes entity context in UserSession
- Entity-based permission resolution

**All Application Modules:**
- Use ValidateToken for request authentication
- Check permissions in UserSession
- Log security events for audit

## 9. Development

### Modifying the Module

**Add New Security Event Type:**
1. Update `SecurityEventType` enum in `auth.proto`
2. Update security_event_type in database
3. Add handling in security service
4. Regenerate proto and SQLC

**Add New 2FA Method:**
1. Update `TwoFactorMethod` enum
2. Implement verification logic
3. Update EnableTwoFactor handler

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f identity/auth/db/sqlc.yaml

# Run migrations
make migrate-up
```

### Testing Guidelines

**Unit Tests:**
```go
func TestLogin(t *testing.T) {
    mockUserRepo := &MockUserRepository{}
    mockPasswordHasher := &MockPasswordHasher{}
    svc := services.NewAuthService(mockUserRepo, mockPasswordHasher, jwtService)

    resp, err := svc.Login(ctx, req)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.AccessToken)
}
```

**Integration Tests:**
- Test complete login flow
- Test token refresh
- Test 2FA verification
- Test session management

## 10. Troubleshooting

### Common Issues

**Issue: Invalid credentials**
- **Cause:** Wrong password or user not found
- **Solution:** Verify username/password, check user status

**Issue: Token expired**
- **Cause:** Access token past expiry time
- **Solution:** Use refresh token to get new access token

**Issue: Account locked**
- **Cause:** Too many failed login attempts
- **Solution:** Use UnlockAccount or wait for lockout duration

**Issue: 2FA verification fails**
- **Cause:** Code mismatch or time drift
- **Solution:** Check device time synchronization, try backup code

**Issue: Session not found**
- **Cause:** Session expired or revoked
- **Solution:** Re-authenticate to create new session

### Performance Tips

1. Use Redis for session storage
2. Implement token blacklist with TTL
3. Cleanup expired sessions/tokens regularly
4. Cache user permissions
5. Use connection pooling for database

### Debugging

**Check active sessions:**
```sql
SELECT session_id, user_id, device_info, ip_address,
       created_at, last_accessed, expires_at
FROM sessions
WHERE user_id = 'user-uuid' AND is_active = TRUE;
```

**Find failed login attempts:**
```sql
SELECT timestamp, ip_address, device_info, description
FROM security_events
WHERE user_id = 'user-uuid'
  AND event_type = 'LOGIN_FAILURE'
  AND timestamp > NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC;
```

**Verify token validity:**
```sql
SELECT token_type, expires_at, revoked_at
FROM tokens
WHERE token_hash = 'hashed-token'
  AND user_id = 'user-uuid';
```

## Additional Resources

- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [TOTP RFC 6238](https://tools.ietf.org/html/rfc6238)
- [Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
