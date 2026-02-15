package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/1mb-dev/nivomoney/shared/errors"
	"github.com/1mb-dev/nivomoney/shared/response"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the context key for user ID.
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email.
	UserEmailKey ContextKey = "user_email"
	// UserRolesKey is the context key for user roles.
	UserRolesKey ContextKey = "user_roles"
	// UserPermissionsKey is the context key for user permissions.
	UserPermissionsKey ContextKey = "user_permissions"
)

// JWTClaims represents the JWT token claims.
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// JWTValidator validates JWT tokens locally without calling external services.
type JWTValidator struct {
	jwtSecret string
}

// NewJWTValidator creates a new JWT validator.
func NewJWTValidator(jwtSecret string) *JWTValidator {
	return &JWTValidator{
		jwtSecret: jwtSecret,
	}
}

// Authenticate is a middleware that validates JWT tokens.
func (v *JWTValidator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, errors.Unauthorized("missing authorization header"))
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(w, errors.Unauthorized("invalid authorization header format"))
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(v.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Error(w, errors.Unauthorized("invalid or expired token"))
			return
		}

		// Add user context to request
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
		ctx = context.WithValue(ctx, UserPermissionsKey, claims.Permissions)

		// Add user info to headers for backend services
		r.Header.Set("X-User-ID", claims.UserID)
		r.Header.Set("X-User-Email", claims.Email)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission is a middleware that checks if user has specific permission.
func (v *JWTValidator) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get permissions from context (set by Authenticate middleware)
			permissions, ok := r.Context().Value(UserPermissionsKey).([]string)
			if !ok {
				response.Error(w, errors.Forbidden("permissions not found in context"))
				return
			}

			// Check if user has required permission
			hasPermission := false
			for _, perm := range permissions {
				if perm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				err := errors.Forbidden("insufficient permissions")
				err.Details = map[string]interface{}{
					"required_permission": permission,
				}
				response.Error(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Optional is a middleware that attempts authentication but doesn't fail if token is missing.
// Useful for endpoints that work both with and without authentication.
func (v *JWTValidator) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Token provided, try to validate it
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(v.jwtSecret), nil
		})

		if err == nil && token.Valid {
			// Valid token, add to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
			ctx = context.WithValue(ctx, UserPermissionsKey, claims.Permissions)

			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Invalid token, continue without authentication
		next.ServeHTTP(w, r)
	})
}
