package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"

	"p9e.in/ugcl/identity/auth/models"
	serviceif "p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

// -------------------- Mocks --------------------

type mockUowFactory struct{ u uow.UnitOfWork }

func (f *mockUowFactory) Create() uow.UnitOfWork { return f.u }

type mockUow struct {
	users    *mockUserRepo
	sessions *mockSessionRepo
	audit    *mockAuditRepo
	tokens   *mockTokenRepo
}

func (m *mockUow) Begin(ctx context.Context) error                                 { return nil }
func (m *mockUow) Commit(ctx context.Context) error                                { return nil }
func (m *mockUow) Rollback(ctx context.Context) error                              { return nil }
func (m *mockUow) Users() interface{ /* narrowed in methods below */ } { return m.users }
func (m *mockUow) Sessions() interface{ /* narrowed in methods below */ } { return m.sessions }
func (m *mockUow) ApiKeys() interface{} { return nil }
func (m *mockUow) Tokens() interface{ /* narrowed in methods below */ } { return m.tokens }
func (m *mockUow) TwoFactor() interface{} { return nil }
func (m *mockUow) Audit() interface{ /* narrowed in methods below */ } { return m.audit }
func (m *mockUow) Execute(ctx context.Context, fn func(ctx context.Context, uow uow.UnitOfWork) error) error {
	return fn(ctx, m)
}
func (m *mockUow) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context, uow uow.UnitOfWork) error) error {
	return fn(ctx, m)
}

// Ensure mockUow implements uow.UnitOfWork
var _ uow.UnitOfWork = (*mockUow)(nil)

// Repositories

type mockUserRepo struct {
	GetByEmailFunc     func(ctx context.Context, email string) (*models.User, error)
	GetByUsernameFunc  func(ctx context.Context, username string) (*models.User, error)
	UpdateFunc         func(ctx context.Context, user *models.User) (*models.User, error)
	UpdateLastLoginHit int
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) { return nil, errors.New("not implemented") }
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)     { return nil, errors.New("not implemented") }
func (m *mockUserRepo) GetByUserID(ctx context.Context, userID string) (*models.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetByEmailFunc != nil { return m.GetByEmailFunc(ctx, email) }
	return nil, nil
}
func (m *mockUserRepo) GetByPhone(ctx context.Context, identifier string) (*models.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.GetByUsernameFunc != nil { return m.GetByUsernameFunc(ctx, username) }
	return nil, nil
}
func (m *mockUserRepo) GetByEmailOrUsername(ctx context.Context, identifier string) (*models.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) (*models.User, error) {
	if m.UpdateFunc != nil { return m.UpdateFunc(ctx, user) }
	return user, nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error { return errors.New("not implemented") }
func (m *mockUserRepo) Activate(ctx context.Context, id uuid.UUID) error { return errors.New("not implemented") }
func (m *mockUserRepo) Deactivate(ctx context.Context, id uuid.UUID) error { return errors.New("not implemented") }
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash, passwordSalt string) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) UpdateFailedLoginAttempts(ctx context.Context, id uuid.UUID, attempts int32) error {
	return nil
}
func (m *mockUserRepo) LockAccount(ctx context.Context, id uuid.UUID, lockedUntil *time.Time) error { return nil }
func (m *mockUserRepo) UnlockAccount(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) MarkEmailVerified(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) MarkPhoneVerified(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error { return nil }
func (m *mockUserRepo) UpdatePhone(ctx context.Context, id uuid.UUID, phone *string) error { return nil }
func (m *mockUserRepo) EnableTwoFactor(ctx context.Context, id uuid.UUID, secret string) error { return nil }
func (m *mockUserRepo) DisableTwoFactor(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) UpdateBackupCodes(ctx context.Context, id uuid.UUID, codes []string) error { return nil }
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error { m.UpdateLastLoginHit++; return nil }
func (m *mockUserRepo) UpdateLastPasswordChange(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) List(ctx context.Context, filter *models.UserFilter, pagination *models.Pagination) (*models.PaginatedResponse[*models.User], error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) Count(ctx context.Context, filter *models.UserFilter) (int64, error) { return 0, errors.New("not implemented") }
func (m *mockUserRepo) Search(ctx context.Context, query string, pagination *models.Pagination) (*models.PaginatedResponse[*models.User], error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*models.UserTenant, error) {
	return nil, nil
}
func (m *mockUserRepo) GetDefaultTenant(ctx context.Context, userID uuid.UUID) (*models.UserTenant, error) { return nil, nil }
func (m *mockUserRepo) EmailExists(ctx context.Context, email string) (bool, error) { return false, nil }
func (m *mockUserRepo) UsernameExists(ctx context.Context, username string) (bool, error) { return false, nil }
func (m *mockUserRepo) UserIDExists(ctx context.Context, userID string) (bool, error) { return false, nil }

// Sessions

type mockSessionRepo struct {
	CreateFunc func(ctx context.Context, session *models.Session) (*models.Session, error)
}

func (m *mockSessionRepo) Create(ctx context.Context, session *models.Session) (*models.Session, error) {
	if m.CreateFunc != nil { return m.CreateFunc(ctx, session) }
	return session, nil
}
func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)    { return nil, errors.New("not implemented") }
func (m *mockSessionRepo) GetBySessionID(ctx context.Context, sessionID string) (*models.Session, error) {
	return nil, errors.New("not implemented")
}
func (m *mockSessionRepo) Update(ctx context.Context, session *models.Session) (*models.Session, error) {
	return session, nil
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockSessionRepo) GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) GetActiveSessionsByUserIDAndTenant(ctx context.Context, userID uuid.UUID, tenantID string) ([]*models.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) UpdateLastAccessed(ctx context.Context, sessionID string) error { return nil }
func (m *mockSessionRepo) RevokeSession(ctx context.Context, sessionID string, reason *string) error { return nil }
func (m *mockSessionRepo) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, reason *string) error { return nil }
func (m *mockSessionRepo) RevokeAllUserSessionsExcept(ctx context.Context, userID uuid.UUID, exceptSessionID string, reason *string) error { return nil }
func (m *mockSessionRepo) IsSessionActive(ctx context.Context, sessionID string) (bool, error) { return true, nil }
func (m *mockSessionRepo) ValidateRefreshToken(ctx context.Context, sessionID string, refreshTokenHash string) (bool, error) {
	return true, nil
}
func (m *mockSessionRepo) List(ctx context.Context, filter *models.SessionFilter, pagination *models.Pagination) (*models.PaginatedResponse[*models.Session], error) {
	return nil, nil
}
func (m *mockSessionRepo) Count(ctx context.Context, filter *models.SessionFilter) (int64, error) { return 0, nil }
func (m *mockSessionRepo) CleanupExpiredSessions(ctx context.Context) error { return nil }
func (m *mockSessionRepo) CleanupRevokedSessions(ctx context.Context, olderThan time.Time) error { return nil }
func (m *mockSessionRepo) GetActiveSessionsCount(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockSessionRepo) GetSessionsCountByTenant(ctx context.Context, tenantID string) (int64, error) { return 0, nil }
func (m *mockSessionRepo) GetUserActiveSessionsCount(ctx context.Context, userID uuid.UUID) (int64, error) { return 0, nil }
func (m *mockSessionRepo) GetActiveSessionsByUser(ctx context.Context, userUUID uuid.UUID) ([]*models.Session, error) { return nil, nil }
func (m *mockSessionRepo) Deactivate(ctx context.Context, sessionID string) error { return nil }
func (m *mockSessionRepo) DeactivateAll(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *mockSessionRepo) DeactivateAllExcept(ctx context.Context, userID uuid.UUID, sessionID string) error { return nil }

// Audit

type mockAuditRepo struct {
	LastLoginAttempt *models.LoginAttempt
	LastSecurityEvt  *models.SecurityEvent
}

func (m *mockAuditRepo) CreateSecurityEvent(ctx context.Context, event *models.SecurityEvent) (*models.SecurityEvent, error) {
	m.LastSecurityEvt = event
	return event, nil
}
func (m *mockAuditRepo) GetSecurityEventsByUser(ctx context.Context, userID uuid.UUID, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error) {
	return nil, nil
}
func (m *mockAuditRepo) GetSecurityEventsByType(ctx context.Context, eventType string, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error) {
	return nil, nil
}
func (m *mockAuditRepo) GetSecurityEventsByTenant(ctx context.Context, tenantID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error) {
	return nil, nil
}
func (m *mockAuditRepo) CleanupOldSecurityEvents(ctx context.Context, olderThan time.Time) error { return nil }
func (m *mockAuditRepo) CreateLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) (*models.LoginAttempt, error) {
	m.LastLoginAttempt = attempt
	return attempt, nil
}
func (m *mockAuditRepo) GetLoginAttemptsByUser(ctx context.Context, userID uuid.UUID, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error) {
	return nil, nil
}
func (m *mockAuditRepo) GetLoginAttemptsByIdentifier(ctx context.Context, identifier string, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error) {
	return nil, nil
}
func (m *mockAuditRepo) GetLoginAttemptsByIP(ctx context.Context, ipAddress string, pagination *models.Pagination) (*models.PaginatedResponse[*models.LoginAttempt], error) {
	return nil, nil
}
func (m *mockAuditRepo) CleanupOldLoginAttempts(ctx context.Context, olderThan time.Time) error { return nil }
func (m *mockAuditRepo) GetFailedLoginCount(ctx context.Context, identifier string, since time.Time) (int64, error) { return 0, nil }
func (m *mockAuditRepo) GetFailedLoginCountByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error) { return 0, nil }
func (m *mockAuditRepo) GetSuspiciousActivities(ctx context.Context, since time.Time, pagination *models.Pagination) (*models.PaginatedResponse[*models.SecurityEvent], error) {
	return nil, nil
}
func (m *mockAuditRepo) CreateUserTenant(ctx context.Context, userTenant *models.UserTenant) (*models.UserTenant, error) { return nil, nil }
func (m *mockAuditRepo) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*models.UserTenant, error) { return nil, nil }
func (m *mockAuditRepo) GetTenantUsers(ctx context.Context, tenantID string, pagination *models.Pagination) (*models.PaginatedResponse[*models.UserTenant], error) { return nil, nil }
func (m *mockAuditRepo) UpdateUserTenantRoles(ctx context.Context, userID uuid.UUID, tenantID string, roles []string) error { return nil }
func (m *mockAuditRepo) SetDefaultTenant(ctx context.Context, userID uuid.UUID, tenantID string) error { return nil }
func (m *mockAuditRepo) DeactivateUserTenant(ctx context.Context, userID uuid.UUID, tenantID string) error { return nil }

// Tokens (unused here, stubbed)

type mockTokenRepo struct{}

// JWT & Events

type mockJWTService struct{
	Pair *serviceif.TokenPair
}

func (m *mockJWTService) GenerateTokens(user *models.User, session *models.Session, tenantID string, roles []string) (*serviceif.TokenPair, error) {
	if m.Pair != nil { return m.Pair, nil }
	return &serviceif.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresAt: time.Now().Add(15*time.Minute), TokenType: "Bearer"}, nil
}
func (m *mockJWTService) ValidateToken(tokenString string) (*serviceif.TokenClaims, error) { return nil, errors.New("not implemented") }
func (m *mockJWTService) RefreshToken(refreshTokenString string) (*serviceif.TokenPair, error) { return nil, errors.New("not implemented") }
func (m *mockJWTService) ExtractJTI(tokenString string) (string, error) { return "jti", nil }

type mockEventService struct{
	FailedLoginPublished bool
	UserLoginPublished   bool
}

func (m *mockEventService) PublishUserLoginEvent(ctx context.Context, user *models.User, sessionID, tenantID, ipAddress string) error {
	m.UserLoginPublished = true
	return nil
}
func (m *mockEventService) PublishUserLogoutEvent(ctx context.Context, userID, sessionID, reason string) error { return nil }
func (m *mockEventService) PublishAccountLockEvent(ctx context.Context, userID, reason, changedBy string) error { return nil }
func (m *mockEventService) PublishAccountUnlockEvent(ctx context.Context, userID, reason, changedBy string) error { return nil }
func (m *mockEventService) PublishTwoFactorEnabledEvent(ctx context.Context, userID, method string) error { return nil }
func (m *mockEventService) PublishTwoFactorDisabledEvent(ctx context.Context, userID string) error { return nil }
func (m *mockEventService) PublishSessionRevokedEvent(ctx context.Context, userID, sessionID, reason string) error { return nil }
func (m *mockEventService) PublishFailedLoginEvent(ctx context.Context, identifier, reason, ipAddress string) error {
	m.FailedLoginPublished = true
	return nil
}
func (m *mockEventService) SubscribeToUserEvents(ctx context.Context) error { return nil }

// Helpers

func hashPassword(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h)
}

// -------------------- Tests --------------------

func TestAuthService_Login_UserNotFound(t *testing.T) {
	users := &mockUserRepo{
		GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("not found")
		},
	}
	sessions := &mockSessionRepo{}
	audit := &mockAuditRepo{}
	muow := &mockUow{users: users, sessions: sessions, audit: audit, tokens: &mockTokenRepo{}}
	factory := &mockUowFactory{u: muow}
	jwt := &mockJWTService{}
	events := &mockEventService{}

	svc := NewAuthService(factory, jwt, events)

	req := &serviceif.LoginRequest{
		IdentifierType: "email",
		Identifier:     "nobody@example.com",
		Password:       "irrelevant",
		TenantID:       "",
		DeviceInfo:     "UA",
		IPAddress:      "127.0.0.1",
		RememberMe:     false,
	}

	_, err := svc.Login(context.Background(), req)
	if err == nil {
		t.Fatalf("expected authentication failure, got nil")
	}
	if audit.LastLoginAttempt == nil || audit.LastLoginAttempt.Success {
		t.Fatalf("expected failed login attempt to be logged")
	}
}

func TestAuthService_Login_Success_CreatesSessionAndReturnsTokens(t *testing.T) {
	password := "StrongP@ssw0rd"
	user := &models.User{
		ID:           uuid.New(),
		UserID:       uuid.NewString(),
		Username:     "john",
		Email:        "john@example.com",
		PasswordHash: hashPassword(password),
		IsActive:     true,
	}

	users := &mockUserRepo{
		GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) { return user, nil },
		UpdateFunc:     func(ctx context.Context, u *models.User) (*models.User, error) { return u, nil },
	}

	createdSession := &models.Session{ID: uuid.New(), SessionID: "session_123", UserID: user.ID, IsActive: true, ExpiresAt: time.Now().Add(24*time.Hour)}
	sessions := &mockSessionRepo{
		CreateFunc: func(ctx context.Context, s *models.Session) (*models.Session, error) {
			return createdSession, nil
		},
	}

	audit := &mockAuditRepo{}
	muow := &mockUow{users: users, sessions: sessions, audit: audit, tokens: &mockTokenRepo{}}
	factory := &mockUowFactory{u: muow}
	jwt := &mockJWTService{Pair: &serviceif.TokenPair{AccessToken: "acc", RefreshToken: "ref", ExpiresAt: time.Now().Add(15 * time.Minute), TokenType: "Bearer"}}
	events := &mockEventService{}

	svc := NewAuthService(factory, jwt, events)

	req := &serviceif.LoginRequest{
		IdentifierType: "email",
		Identifier:     user.Email,
		Password:       password,
		TenantID:       "tenant-1",
		DeviceInfo:     "UA",
		IPAddress:      "127.0.0.1",
		RememberMe:     false,
	}

	res, err := svc.Login(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil response")
	}
	if res.AccessToken != "acc" || res.RefreshToken != "ref" {
		t.Errorf("unexpected tokens: got (%s,%s)", res.AccessToken, res.RefreshToken)
	}
	if res.Session == nil || res.SessionID != createdSession.SessionID {
		t.Errorf("expected session created with id %s", createdSession.SessionID)
	}
	if users.UpdateLastLoginHit == 0 {
		t.Errorf("expected UpdateLastLogin to be called")
	}
	if audit.LastLoginAttempt == nil || !audit.LastLoginAttempt.Success {
		t.Errorf("expected successful login attempt to be logged")
	}
}
