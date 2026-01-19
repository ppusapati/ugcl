package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
)

type jwtService struct {
	secretKey  []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTService(secretKey string, issuer string) interfaces.JWTService {
	return &jwtService{
		secretKey:  []byte(secretKey),
		issuer:     issuer,
		accessTTL:  15 * time.Minute,   // Access tokens expire in 15 minutes
		refreshTTL: 7 * 24 * time.Hour, // Refresh tokens expire in 7 days
	}
}

type CustomClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	TenantID    string   `json:"tenant_id,omitempty"`
	SessionID   string   `json:"session_id"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	TokenType   string   `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func (j *jwtService) GenerateTokens(user *models.User, session *models.Session, tenantID string, roles []string) (*interfaces.TokenPair, error) {
	now := time.Now()
	jti := generateJTI()

	// Generate access token
	accessClaims := &CustomClaims{
		UserID:    user.UserID,
		Username:  user.Username,
		Email:     user.Email,
		TenantID:  tenantID,
		SessionID: session.SessionID,
		Roles:     roles,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   user.UserID,
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{"ugcl-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshJTI := generateJTI()
	refreshClaims := &CustomClaims{
		UserID:    user.UserID,
		SessionID: session.SessionID,
		TenantID:  tenantID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   user.UserID,
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{"ugcl-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &interfaces.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(j.accessTTL),
		TokenType:    "Bearer",
	}, nil
}

func (j *jwtService) ValidateToken(tokenString string) (*interfaces.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &interfaces.TokenClaims{
		UserID:      claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		TenantID:    claims.TenantID,
		SessionID:   claims.SessionID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		TokenType:   claims.TokenType,
		JTI:         claims.ID,
		ExpiresAt:   claims.ExpiresAt.Time,
		IssuedAt:    claims.IssuedAt.Time,
	}, nil
}

func (j *jwtService) RefreshToken(refreshTokenString string) (*interfaces.TokenPair, error) {
	// Validate the refresh token
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type, expected refresh token")
	}

	// Check if token is not expired
	if time.Now().After(claims.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Generate new access token
	now := time.Now()
	jti := generateJTI()

	accessClaims := &CustomClaims{
		UserID:    claims.UserID,
		Username:  claims.Username,
		Email:     claims.Email,
		TenantID:  claims.TenantID,
		SessionID: claims.SessionID,
		Roles:     claims.Roles,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   claims.UserID,
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{"ugcl-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign new access token: %w", err)
	}

	// Generate new refresh token with extended expiry
	refreshJTI := generateJTI()
	refreshClaims := &CustomClaims{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		TenantID:  claims.TenantID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   claims.UserID,
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{"ugcl-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	newRefreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign new refresh token: %w", err)
	}

	return &interfaces.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: newRefreshTokenString,
		ExpiresAt:    now.Add(j.accessTTL),
		TokenType:    "Bearer",
	}, nil
}

func (j *jwtService) ExtractJTI(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.ID, nil
}

func generateJTI() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
