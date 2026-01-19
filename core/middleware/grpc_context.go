package middleware

import (
	"context"

	"p9e.in/ugcl/identity/user/models"
)

type grpcCtxKey string

// const ClaimsContextKey grpcCtxKey = "claims"

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}

func UserFromContext(ctx context.Context) models.User {
	if c, ok := ClaimsFromContext(ctx); ok {
		return models.User{
			Uuid:        c.UserID,
			Username:    &c.Name,
			Phone:       &c.Phone,
			TenantRoles: []string{c.Role},
		}
	}
	return models.User{}
}
