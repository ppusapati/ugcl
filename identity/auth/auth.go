package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/ugcl/identity/auth/repository/interfaces"
	serviceInterfaces "p9e.in/ugcl/identity/auth/services/interfaces"
	"p9e.in/ugcl/identity/auth/uow"
)

// Module represents the auth module with all its dependencies
type Module struct {
	uowFactory       uow.UnitOfWorkFactory
	uow              uow.UnitOfWork
	jwtService       serviceInterfaces.JWTService
	twoFactorService serviceInterfaces.TwoFactorService
	otpService       serviceInterfaces.OTPService
	emailService     serviceInterfaces.EmailService
	smsService       serviceInterfaces.SMSService
}

// NewModule creates a new auth module instance
func NewModule(db *pgxpool.Pool, jwtSecret, issuer string) *Module {
	// Create UOW factory and instance
	uowFactory := uow.NewSQLCFactory(db)
	uowInstance := uowFactory.Create()

	module := &Module{
		uowFactory: uowFactory,
		uow:        uowInstance,
		// Services will be initialized by the service container
		// This avoids circular imports
	}

	return module
}

// GetUnitOfWorkFactory returns the unit of work factory
func (m *Module) GetUnitOfWorkFactory() uow.UnitOfWorkFactory {
	return m.uowFactory
}

// GetUnitOfWork creates a new unit of work instance
func (m *Module) GetUnitOfWork() uow.UnitOfWork {
	return m.uowFactory.Create()
}

// Direct repository access methods (for read-only operations)

func (m *Module) Users() interfaces.UserRepository {
	return m.uow.Users()
}

func (m *Module) Sessions() interfaces.SessionRepository {
	return m.uow.Sessions()
}

func (m *Module) ApiKeys() interfaces.ApiKeyRepository {
	return m.uow.ApiKeys()
}

func (m *Module) Tokens() interfaces.TokenRepository {
	return m.uow.Tokens()
}

func (m *Module) TwoFactor() interfaces.TwoFactorRepository {
	return m.uow.TwoFactor()
}

func (m *Module) Audit() interfaces.AuditRepository {
	return m.uow.Audit()
}

// Service access methods

func (m *Module) GetJWTService() serviceInterfaces.JWTService {
	return m.jwtService
}

func (m *Module) GetTwoFactorService() serviceInterfaces.TwoFactorService {
	return m.twoFactorService
}

func (m *Module) GetOTPService() serviceInterfaces.OTPService {
	return m.otpService
}

func (m *Module) GetEmailService() serviceInterfaces.EmailService {
	return m.emailService
}

func (m *Module) GetSMSService() serviceInterfaces.SMSService {
	return m.smsService
}

// Service setter methods (used by the service container to avoid circular imports)

func (m *Module) SetJWTService(service serviceInterfaces.JWTService) {
	m.jwtService = service
}

func (m *Module) SetTwoFactorService(service serviceInterfaces.TwoFactorService) {
	m.twoFactorService = service
}

func (m *Module) SetOTPService(service serviceInterfaces.OTPService) {
	m.otpService = service
}

func (m *Module) SetEmailService(service serviceInterfaces.EmailService) {
	m.emailService = service
}

func (m *Module) SetSMSService(service serviceInterfaces.SMSService) {
	m.smsService = service
}
