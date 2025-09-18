package auth

import (
	"go.uber.org/fx"

	"p9e.in/ugcl/identity/auth/api/v2/auth/authconnect"
	"p9e.in/ugcl/identity/auth/handlers"
	"p9e.in/ugcl/identity/auth/mappers"
	"p9e.in/ugcl/identity/auth/uow"
	"p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/packages/database/sqlc"
)

// AuthModuleParams contains the parameters needed for auth module initialization
type AuthModuleParams struct {
	fx.In
	DatabaseManager *sqlc.DatabaseManager
	JWTSecret       string `name:"jwt_secret"`
	JWTIssuer       string `name:"jwt_issuer"`
}

// AuthModule bundles all auth service dependencies
var AuthModule = fx.Module("auth",
	fx.Provide(
		// Core auth module instance
		NewAuthModule,

		// Mappers
		mappers.NewUserMapper,
		mappers.NewSessionMapper,
		mappers.NewApiKeyMapper,
		mappers.NewTokenMapper,
		mappers.NewAuditMapper,
		// TODO: Uncomment when protobuf schema is fixed
		// mappers.NewProtobufMapper,

		// UOW factory
		ProvideUOWFactory,

		// JWT Service (standalone)
		ProvideJWTService,

		// Email and SMS services (standalone)
		ProvideEmailService,
		ProvideSMSService,

		// Services that depend on auth module
		ProvideOTPService,
		ProvideTwoFactorService,
		ProvideUserService,
		ProvideAuthService,
		ProvideSessionService,
		ProvideTenantService,
		ProvideAuditService,

		// Event service (optional)
		ProvideEventService,

		// Error handler
		handlers.NewErrorHandler,

		// Handler layer
		ProvideAuthHandler,
	),
)

// NewAuthModule creates a new auth module instance with fx dependencies
func NewAuthModule(params AuthModuleParams) *Module {
	return NewModule(params.DatabaseManager.Pool, params.JWTSecret, params.JWTIssuer)
}

// UOW factory provider
func ProvideUOWFactory(authModule *Module) uow.UnitOfWorkFactory {
	return authModule.GetUnitOfWorkFactory()
}

// JWT Service provider
func ProvideJWTService(params AuthModuleParams) interfaces.JWTService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// Email Service provider
func ProvideEmailService() interfaces.EmailService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// SMS Service provider
func ProvideSMSService() interfaces.SMSService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// OTP Service provider
func ProvideOTPService(authModule *Module, emailService interfaces.EmailService, smsService interfaces.SMSService) interfaces.OTPService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// TwoFactor Service provider
func ProvideTwoFactorService(authModule *Module) interfaces.TwoFactorService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// User Service provider
func ProvideUserService(authModule *Module) interfaces.UserService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// Auth Service provider
func ProvideAuthService(
	uowFactory uow.UnitOfWorkFactory,
	jwtService interfaces.JWTService,
	eventService interfaces.EventService,
) interfaces.AuthService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// Session Service provider
func ProvideSessionService(uowFactory uow.UnitOfWorkFactory) interfaces.SessionService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// Tenant Service provider
func ProvideTenantService(authModule *Module) interfaces.TenantService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// Audit Service provider
func ProvideAuditService(authModule *Module) interfaces.AuditService {
	// Import services locally to avoid circular import
	return nil // TODO: implement factory without importing services package
}

// Event Service provider (optional)
func ProvideEventService(authModule *Module) interfaces.EventService {
	// Return nil for now - event service requires additional dependencies
	// that should be provided separately if needed
	return nil
}

// Auth Handler provider
func ProvideAuthHandler(
	authService interfaces.AuthService,
	sessionService interfaces.SessionService,
	jwtService interfaces.JWTService,
	twoFactorService interfaces.TwoFactorService,
	eventService interfaces.EventService,
	protobufMapper *mappers.ProtobufMapper,
	errorHandler *handlers.ErrorHandler,
) authconnect.AuthServiceHandler {
	return handlers.NewAuthHandler(
		authService,
		sessionService,
		jwtService,
		twoFactorService,
		eventService,
		protobufMapper,
		errorHandler,
	)
}
