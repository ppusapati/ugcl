// auth/jwt.go
package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"p9e.in/ugcl/models"
)

// Grab your secret from env (or config)
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Claims are the custom payload in your JWT
type Claims struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// unexported type prevents collisions in context
type ctxKey int

const (
	userClaimsKey ctxKey = iota
)

// GenerateToken creates a signed JWT valid for 24 h
func GenerateToken(userID, role, name, phone string) (string, error) {
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
	return token.SignedString(jwtKey)
}

// JWTMiddleware validates the token and stashes the Claims in ctx
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		// attach the full Claims object to context
		ctx := context.WithValue(r.Context(), userClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole wraps a handler and ensures the JWT’s role matches
func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetRole(r) != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetClaims pulls the *Claims out of the request context (or nil)
func GetClaims(r *http.Request) *Claims {
	if c, ok := r.Context().Value(userClaimsKey).(*Claims); ok {
		return c
	}
	return nil
}

// Convenience methods:
func GetUserID(r *http.Request) string {
	if c := GetClaims(r); c != nil {
		return c.UserID
	}
	return ""
}

func GetUser(r *http.Request) models.User {
	if c := GetClaims(r); c != nil {
		User := models.User{
			Name:  c.Name,
			Phone: c.Phone,
			Role:  c.Role,
		}
		return User // or return c.Username if you have that field
	}
	return models.User{} // return zero value if no claims found
}
func GetRole(r *http.Request) string {
	if c := GetClaims(r); c != nil {
		return c.Role
	}
	return ""
}
