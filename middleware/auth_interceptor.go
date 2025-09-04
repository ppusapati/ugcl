package middleware

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

// AuthInterceptor returns a Connect interceptor for JWT authentication
func AuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			// Extract authorization header from Connect request
			authHeader := req.Header().Get("authorization")
			if authHeader == "" {
				// Also try Authorization with capital A
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
				return jwtKey, nil
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

			// Inject claims into context
			ctx = context.WithValue(ctx, ClaimsContextKey, claims)

			// Continue with the request
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
