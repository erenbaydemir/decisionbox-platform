package auth

import (
	"context"
	"net/http"
)

// UserPrincipal represents the authenticated user.
type UserPrincipal struct {
	ID    string
	AppID string
	OrgID string
	Roles []string
}

type contextKey string

const userKey contextKey = "user"

// FromContext extracts UserPrincipal from request context.
func FromContext(ctx context.Context) (*UserPrincipal, bool) {
	u, ok := ctx.Value(userKey).(*UserPrincipal)
	return u, ok
}

// WithUser stores UserPrincipal in context.
func WithUser(ctx context.Context, user *UserPrincipal) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// Provider is the interface for authentication backends.
type Provider interface {
	ValidateToken(ctx context.Context, token string) (*UserPrincipal, error)
	Middleware() func(http.Handler) http.Handler
}
