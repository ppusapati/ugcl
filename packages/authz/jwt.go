package authz

import (
	"errors"
	"time"

	pb "p9e.in/ugcl/identity/user/api/v2/permission"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines the structure embedded in JWT
type CustomClaims struct {
	UserID      string          `json:"sub"`
	TenantID    string          `json:"tenant_id"`
	Role        string          `json:"role"`
	Permissions []pb.Permission `json:"permissions"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("your-secret-key") // Replace with env var or config

// ParseJWT parses the token and returns claims
func ParseJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token or claims")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
