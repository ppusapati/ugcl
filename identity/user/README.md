# User Management Module

## 1. Module Overview

The User Management module provides core user account functionality including user CRUD operations, role management, password management, email/phone verification, profile management, and user preferences. It serves as the foundation for authentication and identity management across the UGCL platform.

**Purpose:** Centralized user account management with multi-tenant support, role-based access control, secure password handling, and user preference management.

**Key Features:**
- User account creation and management
- Multi-tenant user support (user can belong to multiple tenants)
- Role assignment per tenant
- Secure password management with strength validation
- Password reset via email/SMS
- Email and phone verification
- Two-factor authentication setup
- User profile and avatar management
- User preferences (language, timezone, notifications)
- User status management (active, inactive, suspended, locked)

## 2. Architecture

### Module Structure
```
identity/user/
├── proto/
│   ├── user.proto           # User service definitions
│   └── permission.proto     # Permission service definitions
├── db/
│   ├── schema/              # Database schemas
│   └── generated/           # SQLC generated code
├── repository/              # Data access layer
│   ├── user_repository.go
│   ├── role_repository.go
│   └── permission_repository.go
├── services/                # Business logic layer
│   ├── user_service.go
│   ├── permission_service.go
│   ├── permission_resolver.go
│   ├── entity_role_provider.go
│   ├── organization_hierarchy_provider.go
│   └── resource_ownership_provider.go
├── handler/                 # gRPC handlers
│   └── user_handler.go
├── mappers/                 # DTO mappers
│   ├── user_mapper.go
│   └── permission_mapper.go
├── models/                  # Domain models
│   ├── user.go
│   └── permissions.go
├── helper/                  # Helper functions
├── uow/                     # Unit of work pattern
└── module.go                # FX module definition
```

### Database Schema

**Tables:**
- `users` - Core user accounts
- `roles` - Role definitions
- `permissions` - Permission definitions
- `user_roles` - User-role assignments
- `role_permissions` - Role-permission mappings
- `user_tenant_roles` - Multi-tenant role assignments

**Key Relationships:**
- User → Tenants (N:N via tenant_ids array)
- User → Roles (N:N via user_roles, scoped by tenant)
- Roles → Permissions (N:N via role_permissions)

### Key Dependencies
- `identity/auth` - Authentication and sessions
- `identity/tenant` - Tenant management
- `identity/entity` - Entity abstraction
- Password hashing (bcrypt)
- Email/SMS services for verification
- Connect-Go RPC framework

## 3. Quick Start

### Create a User

```go
import (
    userv2 "p9e.in/ugcl/identity/user/api/v2/user"
    "connectrpc.com/connect"
)

client := userv2.NewUserServiceClient(httpClient, baseURL)

// Create new user
user, err := client.CreateUser(ctx, connect.NewRequest(&userv2.CreateUserRequest{
    User: &userv2.User{
        Username:       wrapperspb.String("john.doe"),
        Email:          wrapperspb.String("john.doe@ugcl.com"),
        Fullname:       wrapperspb.String("John Doe"),
        Phone:          wrapperspb.String("+91-9876543210"),
        Password:       "SecurePass123!",
        ConfirmPassword: "SecurePass123!",
        Gender:         userv2.Gender_MALE,
        TenantIds:      []string{"tenant-uuid-1"},
        IsActive:       true,
        Status:         userv2.UserStatus_USER_STATUS_ACTIVE,
    },
}))
```

### Common Operations

**Get user:**
```go
user, err := client.GetUser(ctx, connect.NewRequest(&userv2.UserIdentifier{
    Identifier: &userv2.UserIdentifier_Uuid{
        Uuid: "user-uuid",
    },
}))
```

**Assign role to user:**
```go
err := client.AssignRoleToUser(ctx, connect.NewRequest(&userv2.AssignRoleRequest{
    UserId: "user-uuid",
    RoleId: "role-uuid",
}))
```

**Change password:**
```go
err := client.ChangePassword(ctx, connect.NewRequest(&userv2.ChangePasswordRequest{
    UserId:          "user-uuid",
    CurrentPassword: "OldPass123!",
    NewPassword:     "NewSecurePass456!",
    ConfirmPassword: "NewSecurePass456!",
}))
```

**Send verification email:**
```go
err := client.SendVerificationEmail(ctx, connect.NewRequest(&userv2.SendVerificationRequest{
    UserId: "user-uuid",
    Type:   userv2.VerificationType_VERIFICATION_TYPE_EMAIL,
}))
```

## 4. API Reference

### User CRUD Operations
- `ListUsers` - List users with filtering and pagination
- `GetUser` - Get user by ID or UUID
- `CreateUser` - Create new user account
- `UpdateUser` - Update user details with field mask
- `DeleteUser` - Delete user account (with force option)

### Role Management
- `AssignRoleToUser` - Assign role to user
- `RevokeRoleFromUser` - Remove role from user

### Password Management
- `ChangePassword` - Change password with current password validation
- `InitiatePasswordReset` - Start password reset flow (email/SMS)
- `ResetPassword` - Complete password reset with token
- `ValidatePasswordStrength` - Check password strength

### Email/Phone Verification
- `SendVerificationEmail` - Send email verification code
- `VerifyEmail` - Verify email with code
- `SendVerificationSMS` - Send SMS verification code
- `VerifyPhone` - Verify phone with code

### Profile Management
- `UpdateProfile` - Update user profile information
- `UploadAvatar` - Upload user avatar image
- `DeleteAvatar` - Remove user avatar

### User Preferences
- `GetUserPreferences` - Get user preferences
- `UpdateUserPreferences` - Update preferences with field mask

## 5. Database Schema

### Table: users

```sql
-- Identity fields
- id (SERIAL, PK)
- uuid (UUID, UNIQUE, generated)
- username (VARCHAR, UNIQUE)
- email (VARCHAR, UNIQUE)
- phone (VARCHAR, UNIQUE)
- fullname (VARCHAR)
- password_hash (VARCHAR) -- bcrypt hashed

-- Personal information
- gender (ENUM) -- MALE, FEMALE, OTHER, UNKNOWN
- avatar (BYTEA or VARCHAR for URL)

-- Multi-tenancy
- tenant_ids (TEXT[]) -- Array of tenant UUIDs

-- Security
- two_factor_enabled (BOOLEAN)
- two_factor_secret (VARCHAR, encrypted)
- email_verified (BOOLEAN)
- phone_verified (BOOLEAN)

-- Status
- is_active (BOOLEAN)
- status (ENUM) -- ACTIVE, INACTIVE, SUSPENDED, PENDING_VERIFICATION, LOCKED

-- Preferences (separate table or JSONB)
- preferences_id (FK → user_preferences)

-- Audit
- created_at, updated_at
- last_login (TIMESTAMP)
```

### Table: user_tenant_roles

```sql
- id (SERIAL, PK)
- user_id (FK → users)
- tenant_id (UUID)
- role_ids (TEXT[]) -- Array of role UUIDs for this tenant
- is_active (BOOLEAN)
- assigned_at (TIMESTAMP)
```

### Table: user_preferences

```sql
- user_id (FK → users, PK)
- language (VARCHAR) -- en, hi, etc.
- timezone (VARCHAR) -- Asia/Kolkata, UTC, etc.
- date_format (VARCHAR) -- DD/MM/YYYY, MM/DD/YYYY
- time_format (VARCHAR) -- 12h, 24h
- email_notifications (BOOLEAN)
- sms_notifications (BOOLEAN)
- push_notifications (BOOLEAN)
- custom_preferences (JSONB)
```

### Indexes
- `idx_users_username` - Username lookup
- `idx_users_email` - Email lookup
- `idx_users_phone` - Phone lookup
- `idx_users_uuid` - UUID lookup
- `idx_users_tenant_ids` - GIN index for tenant membership
- `idx_users_status` - Status filtering
- `idx_users_last_login` - Login analytics

## 6. Configuration

### Environment Variables
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl

# Password policy
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_DIGIT=true
PASSWORD_REQUIRE_SPECIAL=true

# Verification
EMAIL_VERIFICATION_EXPIRY=24h
SMS_VERIFICATION_EXPIRY=10m
RESET_TOKEN_EXPIRY=1h

# Two-factor authentication
TOTP_ISSUER=UGCL
TOTP_PERIOD=30
```

### Module-Specific Settings
- Password hashing cost (bcrypt rounds)
- Verification code length
- Session timeout
- Avatar size limits

## 7. Examples

### Complete User Registration Flow

```go
// 1. Validate password strength
validation, err := client.ValidatePasswordStrength(ctx, connect.NewRequest(&userv2.ValidatePasswordStrengthRequest{
    Password: "MySecurePass123!",
    Username: wrapperspb.String("john.doe"),
    Email:    wrapperspb.String("john.doe@ugcl.com"),
}))

if !validation.Msg.IsValid {
    fmt.Println("Password requirements not met:")
    for _, req := range validation.Msg.RequirementsFailed {
        fmt.Printf("- %s\n", req)
    }
    return
}

// 2. Create user
user, err := client.CreateUser(ctx, connect.NewRequest(&userv2.CreateUserRequest{
    User: &userv2.User{
        Username:       wrapperspb.String("john.doe"),
        Email:          wrapperspb.String("john.doe@ugcl.com"),
        Fullname:       wrapperspb.String("John Doe"),
        Phone:          wrapperspb.String("+91-9876543210"),
        Password:       "MySecurePass123!",
        ConfirmPassword: "MySecurePass123!",
        TenantIds:      []string{tenantID},
        Status:         userv2.UserStatus_USER_STATUS_PENDING_VERIFICATION,
    },
}))

userID := user.Msg.GetUuid()

// 3. Send email verification
err = client.SendVerificationEmail(ctx, connect.NewRequest(&userv2.SendVerificationRequest{
    UserId: userID,
    Type:   userv2.VerificationType_VERIFICATION_TYPE_EMAIL,
}))

// 4. User verifies email (from email link)
err = client.VerifyEmail(ctx, connect.NewRequest(&userv2.VerifyEmailRequest{
    UserId:           userID,
    VerificationCode: "123456",
}))

// 5. Update user status to active
updated, err := client.UpdateUser(ctx, connect.NewRequest(&userv2.UpdateUserRequest{
    User: &userv2.User{
        Uuid:   wrapperspb.String(userID),
        Status: userv2.UserStatus_USER_STATUS_ACTIVE,
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"status"},
    },
}))
```

### Password Reset Flow

```go
// 1. User initiates password reset
reset, err := client.InitiatePasswordReset(ctx, connect.NewRequest(&userv2.InitiatePasswordResetRequest{
    Identifier: &userv2.InitiatePasswordResetRequest_Email{
        Email: "john.doe@ugcl.com",
    },
    TenantId: wrapperspb.String(tenantID),
}))

fmt.Printf("Reset token sent to: %s\n", reset.Msg.MaskedContact)
fmt.Printf("Expires at: %s\n", reset.Msg.ExpiresAt.AsTime())

// 2. User receives email with token and submits new password
err = client.ResetPassword(ctx, connect.NewRequest(&userv2.ResetPasswordRequest{
    ResetToken:      reset.Msg.ResetToken,
    NewPassword:     "NewSecurePass456!",
    ConfirmPassword: "NewSecurePass456!",
}))
```

### Multi-Tenant User Management

```go
// Create user with multiple tenant access
user, err := client.CreateUser(ctx, connect.NewRequest(&userv2.CreateUserRequest{
    User: &userv2.User{
        Username:       wrapperspb.String("multi.tenant"),
        Email:          wrapperspb.String("user@example.com"),
        Password:       "Password123!",
        ConfirmPassword: "Password123!",
        TenantIds:      []string{"tenant-1-uuid", "tenant-2-uuid"},
        TenantRoles: []*userv2.UserTenantRole{
            {
                TenantId: "tenant-1-uuid",
                Roles:    []string{"admin-role-uuid"},
                IsActive: true,
            },
            {
                TenantId: "tenant-2-uuid",
                Roles:    []string{"user-role-uuid"},
                IsActive: true,
            },
        },
    },
}))
```

## 8. Integration

### With Other Modules

**Identity/Auth Module:**
- Auth uses user credentials for login
- Auth session includes user information
- Two-factor authentication coordination

**Identity/Tenant Module:**
- Users belong to tenants via tenant_ids
- Tenant context determines available roles
- Tenant-specific user filtering

**Identity/Entity Module:**
- User accounts linked to entities
- Entity.user_id → User.uuid
- Single user can have multiple entity types

**Employee/Contractor/Vendor Modules:**
- Domain records create user accounts
- User credentials for portal access
- Profile synchronization

## 9. Development

### Modifying the Module

**Add New User Field:**
1. Update user proto message
2. Update database schema
3. Run migration
4. Regenerate SQLC and proto
5. Update mappers

**Add New Permission:**
1. Define in permissions table
2. Assign to roles
3. Use in permission checks

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f identity/user/db/sqlc.yaml

# Run migrations
make migrate-up
```

### Testing Guidelines

**Unit Tests:**
```go
func TestCreateUser(t *testing.T) {
    mockRepo := &MockUserRepository{}
    svc := services.NewUserService(mockRepo, passwordHasher)

    user, err := svc.CreateUser(ctx, req)
    assert.NoError(t, err)
    assert.True(t, user.EmailVerified == false)
}
```

**Integration Tests:**
- Test password hashing
- Test email verification flow
- Test multi-tenant access
- Test role assignments

## 10. Troubleshooting

### Common Issues

**Issue: Duplicate username/email/phone**
- **Cause:** User already exists
- **Solution:** Check uniqueness before creation

**Issue: Weak password rejected**
- **Cause:** Password doesn't meet policy
- **Solution:** Use ValidatePasswordStrength first

**Issue: Email verification fails**
- **Cause:** Code expired or invalid
- **Solution:** Resend verification email

**Issue: Password reset token expired**
- **Cause:** Token older than expiry time
- **Solution:** Initiate new password reset

**Issue: User locked out**
- **Cause:** Too many failed login attempts
- **Solution:** Use UnlockAccount in auth module

### Performance Tips

1. Index username, email, phone for fast lookups
2. Cache user roles and permissions
3. Use partial indexes on active users
4. Batch user imports with transactions

### Debugging

**Check user existence:**
```sql
SELECT uuid, username, email, status, email_verified
FROM users
WHERE email = 'user@example.com';
```

**Verify tenant membership:**
```sql
SELECT username, tenant_ids
FROM users
WHERE 'tenant-uuid' = ANY(tenant_ids);
```

**Check user roles:**
```sql
SELECT u.username, utr.tenant_id, utr.role_ids
FROM users u
JOIN user_tenant_roles utr ON u.id = utr.user_id
WHERE u.uuid = 'user-uuid';
```

## Additional Resources

- [OWASP Password Guidelines](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Bcrypt Password Hashing](https://github.com/golang/crypto/tree/master/bcrypt)
- [Multi-Tenancy Patterns](https://docs.microsoft.com/en-us/azure/architecture/guide/multitenant/overview)
- [RBAC Best Practices](https://www.okta.com/identity-101/role-based-access-control-rbac/)
