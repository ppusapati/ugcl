// core/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Config represents the authentication configuration
type Config struct {
	JWT struct {
		Secret string `yaml:"secret" json:"secret" mapstructure:"secret"`
	} `yaml:"jwt" json:"jwt" mapstructure:"jwt"`

	APIKeys map[string]APIKeyConfig `yaml:"api_keys" json:"api_keys" mapstructure:"api_keys"`

	Security struct {
		WhitelistedIPs []string `yaml:"whitelisted_ips" json:"whitelisted_ips" mapstructure:"whitelisted_ips"`
	} `yaml:"security" json:"security" mapstructure:"security"`
}

// APIKeyConfig represents configuration for each API key
type APIKeyConfig struct {
	AppName        string          `yaml:"app_name" json:"app_name" mapstructure:"app_name"`
	AllowedPaths   []string        `yaml:"allowed_paths" json:"allowed_paths" mapstructure:"allowed_paths"`
	AllowedMethods map[string]bool `yaml:"allowed_methods" json:"allowed_methods" mapstructure:"allowed_methods"`
	SkipIPCheck    bool            `yaml:"skip_ip_check" json:"skip_ip_check" mapstructure:"skip_ip_check"`
}

// AuthService handles authentication with dependency injection
type AuthService struct {
	config         *Config
	jwtKey         []byte
	apiKeyConfigs  map[string]APIKeyInfo
	whitelistedIPs map[string]bool
}

type Claims struct {
	UserID uuid.UUID `json:"userId"`
	Name   string    `json:"name"`
	Phone  string    `json:"phone"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// AuthType represents the type of authentication used
type AuthType string

const (
	AuthTypeJWT    AuthType = "jwt"
	AuthTypeAPIKey AuthType = "api_key"
)

// AuthContext contains authentication information
type AuthContext struct {
	Type        AuthType
	Claims      *Claims
	APIKeyInfo  *APIKeyInfo
	ClientIP    string
	RequestTime time.Time
}

// APIKeyInfo contains information about the API key
type APIKeyInfo struct {
	AppName        string
	AllowedPaths   []string
	AllowedMethods map[string]bool
	SkipIPCheck    bool
}

// Context keys
type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
	AuthContextKey   contextKey = "auth_context"
)

// LoadAuthConfigFromFile loads authentication configuration from a YAML file
func LoadAuthConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth config file %s: %w", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse auth config: %w", err)
	}

	// Validate required fields
	if err := validateAuthConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid auth config: %w", err)
	}

	return &config, nil
}

// validateAuthConfig validates the authentication configuration
func validateAuthConfig(config *Config) error {
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	if len(config.APIKeys) == 0 {
		return fmt.Errorf("at least one API key must be configured")
	}

	for apiKey, apiKeyConfig := range config.APIKeys {
		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		if apiKeyConfig.AppName == "" {
			return fmt.Errorf("app_name is required for API key %s", maskAPIKey(apiKey))
		}

		if len(apiKeyConfig.AllowedPaths) == 0 {
			return fmt.Errorf("allowed_paths is required for API key %s", maskAPIKey(apiKey))
		}

		if len(apiKeyConfig.AllowedMethods) == 0 {
			return fmt.Errorf("allowed_methods is required for API key %s", maskAPIKey(apiKey))
		}
	}

	return nil
}

// maskAPIKey masks API key for logging
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// NewAuthService creates a new AuthService with the provided config
func NewAuthService(config *Config) (*AuthService, error) {
	if config.JWT.Secret == "" {
		return nil, errors.New("JWT secret is required in config")
	}

	authService := &AuthService{
		config:         config,
		jwtKey:         []byte(config.JWT.Secret),
		apiKeyConfigs:  make(map[string]APIKeyInfo),
		whitelistedIPs: make(map[string]bool),
	}

	// Convert config format to internal format
	for apiKey, apiKeyConfig := range config.APIKeys {
		authService.apiKeyConfigs[apiKey] = APIKeyInfo{
			AppName:        apiKeyConfig.AppName,
			AllowedPaths:   apiKeyConfig.AllowedPaths,
			AllowedMethods: apiKeyConfig.AllowedMethods,
			SkipIPCheck:    apiKeyConfig.SkipIPCheck,
		}
	}

	// Set up whitelisted IPs
	for _, ip := range config.Security.WhitelistedIPs {
		authService.whitelistedIPs[ip] = true
	}

	return authService, nil
}

// GenerateToken creates a signed JWT valid for 24 hours
func (as *AuthService) GenerateToken(userID uuid.UUID, role, name, phone string) (string, error) {
	claims := Claims{
		UserID: userID,
		Name:   name,
		Phone:  phone,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(as.jwtKey)
}

// AuthInterceptor returns a Connect interceptor for authentication
func (as *AuthService) AuthInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Extract client IP
			clientIP := as.extractClientIP(req)

			// Create auth context
			authCtx := &AuthContext{
				ClientIP:    clientIP,
				RequestTime: time.Now(),
			}

			// Check for API key first
			apiKey := req.Header().Get("x-api-key")
			if apiKey != "" {
				return as.handleAPIKeyAuth(ctx, req, next, authCtx, apiKey)
			}

			// Fall back to JWT authentication
			return as.handleJWTAuth(ctx, req, next, authCtx)
		}
	}
}

// handleAPIKeyAuth processes API key-based authentication
func (as *AuthService) handleAPIKeyAuth(
	ctx context.Context,
	req connect.AnyRequest,
	next connect.UnaryFunc,
	authCtx *AuthContext,
	apiKey string,
) (connect.AnyResponse, error) {
	// Validate API key
	apiKeyInfo, exists := as.apiKeyConfigs[apiKey]
	if !exists {
		log.Printf("[SECURITY] 🔒 Blocked - Invalid API key. IP=%s Path=%s",
			authCtx.ClientIP, req.Spec().Procedure)
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("invalid or missing API key"),
		)
	}

	// IP validation for non-mobile apps
	if !apiKeyInfo.SkipIPCheck && !as.whitelistedIPs[authCtx.ClientIP] {
		log.Printf("[SECURITY] 🚫 Blocked - IP not whitelisted. App=%s IP=%s Path=%s",
			apiKeyInfo.AppName, authCtx.ClientIP, req.Spec().Procedure)
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			errors.New("access from this IP is not allowed"),
		)
	}

	// Path validation
	if !as.isPathAllowed(req.Spec().Procedure, apiKeyInfo.AllowedPaths) {
		log.Printf("[SECURITY] ⛔️ Denied - Path not allowed. App=%s IP=%s Path=%s",
			apiKeyInfo.AppName, authCtx.ClientIP, req.Spec().Procedure)
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			errors.New("access to this endpoint is not allowed for this app"),
		)
	}

	// Method validation (Connect RPC is always POST, but we can extend this)
	httpMethod := "POST" // Connect RPC uses POST
	if !apiKeyInfo.AllowedMethods[httpMethod] {
		log.Printf("[SECURITY] ⛔️ Denied - Method not allowed. App=%s Method=%s Path=%s",
			apiKeyInfo.AppName, httpMethod, req.Spec().Procedure)
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			errors.New("this method is not allowed for this app"),
		)
	}

	// Set auth context
	authCtx.Type = AuthTypeAPIKey
	authCtx.APIKeyInfo = &apiKeyInfo

	// Log successful API key authentication
	log.Printf("[SECURITY] ✅ API Key Auth - App=%s IP=%s Path=%s Time=%s",
		apiKeyInfo.AppName, authCtx.ClientIP, req.Spec().Procedure,
		authCtx.RequestTime.Format(time.RFC3339))

	// Add auth context to request context
	ctx = context.WithValue(ctx, AuthContextKey, authCtx)

	return next(ctx, req)
}

// handleJWTAuth processes JWT-based authentication
func (as *AuthService) handleJWTAuth(
	ctx context.Context,
	req connect.AnyRequest,
	next connect.UnaryFunc,
	authCtx *AuthContext,
) (connect.AnyResponse, error) {
	// Extract authorization header
	authHeader := req.Header().Get("authorization")
	if authHeader == "" {
		authHeader = req.Header().Get("Authorization")
	}

	if authHeader == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("authorization header required"),
		)
	}

	// Validate Bearer token format
	token := strings.TrimSpace(authHeader)
	if !strings.HasPrefix(token, "Bearer ") {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("invalid authorization format"),
		)
	}

	// Extract raw token
	rawToken := strings.TrimPrefix(token, "Bearer ")

	// Parse and validate JWT token
	parsedToken, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(t *jwt.Token) (any, error) {
		return as.jwtKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("invalid or expired token"),
		)
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("could not parse claims"),
		)
	}

	// Set auth context
	authCtx.Type = AuthTypeJWT
	authCtx.Claims = claims

	// Log successful JWT authentication
	log.Printf("[SECURITY] ✅ JWT Auth - UserID=%s Name=%s Role=%s IP=%s Path=%s Time=%s",
		claims.UserID.String(), claims.Name, claims.Role,
		authCtx.ClientIP, req.Spec().Procedure, authCtx.RequestTime.Format(time.RFC3339))

	// Add both claims and auth context to request context
	ctx = context.WithValue(ctx, ClaimsContextKey, claims)
	ctx = context.WithValue(ctx, AuthContextKey, authCtx)

	return next(ctx, req)
}

// isPathAllowed checks if the requested path is allowed for the API key
func (as *AuthService) isPathAllowed(requestPath string, allowedPaths []string) bool {
	for _, allowedPath := range allowedPaths {
		if strings.HasSuffix(allowedPath, "*") {
			// Wildcard match
			prefix := strings.TrimSuffix(allowedPath, "*")
			if strings.HasPrefix(requestPath, prefix) {
				return true
			}
		} else if requestPath == allowedPath {
			// Exact match
			return true
		}
	}
	return false
}

// extractClientIP extracts the client IP from Connect request headers
func (as *AuthService) extractClientIP(req connect.AnyRequest) string {
	// Try to get IP from various headers
	if ip := req.Header().Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := req.Header().Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := req.Header().Get("CF-Connecting-IP"); ip != "" { // Cloudflare
		return ip
	}

	// Fallback - this might not be available in Connect
	return "unknown"
}

// RequireRole creates an interceptor that requires specific roles (JWT only)
func (as *AuthService) RequireRole(allowedRoles []string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip role check for API key authentication
			if IsAPIKeyAuth(ctx) {
				return next(ctx, req)
			}

			// For JWT authentication, check roles
			claims := GetClaimsFromContext(ctx)
			if claims == nil {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("authentication required"),
				)
			}

			// Check if user has required role
			hasRole := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				log.Printf("[SECURITY] ⛔️ Role denied - UserID=%s Role=%s Required=%v Path=%s",
					claims.UserID.String(), claims.Role, allowedRoles, req.Spec().Procedure)
				return nil, connect.NewError(
					connect.CodePermissionDenied,
					errors.New("insufficient permissions"),
				)
			}

			return next(ctx, req)
		}
	}
}

// RequireApp creates an interceptor that requires specific app (API key only)
func (as *AuthService) RequireApp(allowedApps []string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip app check for JWT authentication
			if IsJWTAuth(ctx) {
				return next(ctx, req)
			}

			appName := GetAppName(ctx)
			if appName == "" {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("app authentication required"),
				)
			}

			// Check if app is allowed
			hasAccess := false
			for _, app := range allowedApps {
				if appName == app {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				log.Printf("[SECURITY] ⛔️ App denied - App=%s Allowed=%v Path=%s",
					appName, allowedApps, req.Spec().Procedure)
				return nil, connect.NewError(
					connect.CodePermissionDenied,
					errors.New("app not authorized for this endpoint"),
				)
			}

			return next(ctx, req)
		}
	}
}

// Helper functions to extract auth information from context

// GetAuthContext retrieves the authentication context from request context
func GetAuthContext(ctx context.Context) *AuthContext {
	if authCtx, ok := ctx.Value(AuthContextKey).(*AuthContext); ok {
		return authCtx
	}
	return nil
}

// GetClaimsFromContext retrieves JWT claims from context
func GetClaimsFromContext(ctx context.Context) *Claims {
	if claims, ok := ctx.Value(ClaimsContextKey).(*Claims); ok {
		return claims
	}
	return nil
}

// IsAPIKeyAuth checks if the request was authenticated using API key
func IsAPIKeyAuth(ctx context.Context) bool {
	authCtx := GetAuthContext(ctx)
	return authCtx != nil && authCtx.Type == AuthTypeAPIKey
}

// IsJWTAuth checks if the request was authenticated using JWT
func IsJWTAuth(ctx context.Context) bool {
	authCtx := GetAuthContext(ctx)
	return authCtx != nil && authCtx.Type == AuthTypeJWT
}

// GetAppName returns the app name for API key authenticated requests
func GetAppName(ctx context.Context) string {
	authCtx := GetAuthContext(ctx)
	if authCtx != nil && authCtx.APIKeyInfo != nil {
		return authCtx.APIKeyInfo.AppName
	}
	return ""
}

// LogConfig logs the current authentication configuration (for debugging)
func (as *AuthService) LogConfig() {
	log.Println("🔐 Authentication Configuration Loaded:")
	log.Printf("   JWT Secret: %s", maskSecret(as.config.JWT.Secret))

	log.Println("   API Keys configured:")
	for _, apiKeyInfo := range as.apiKeyConfigs {
		log.Printf("     • %s (IP Check: %t)",
			apiKeyInfo.AppName, !apiKeyInfo.SkipIPCheck)
	}

	log.Println("   Whitelisted IPs:")
	for ip := range as.whitelistedIPs {
		log.Printf("     • %s", ip)
	}
}

// maskSecret masks sensitive values for logging
func maskSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:2] + "****" + secret[len(secret)-2:]
}
