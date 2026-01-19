package middleware

import (
	"go.uber.org/fx"
)

// AuthModule provides the authentication service as an Fx module
var AuthModule = fx.Module("auth",
	fx.Provide(NewAuthService),
	fx.Invoke(registerAuthService),
)

// registerAuthService is called during fx lifecycle to initialize auth service
func registerAuthService(authService *AuthService) {
	authService.LogConfig()
}
