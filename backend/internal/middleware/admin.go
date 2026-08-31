package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/alumasinde/budget254-paye-api/internal/response"
)

type adminKey string

const (
	adminID          adminKey = "admin_id"
	adminDBID        adminKey = "admin_db_id"
	adminPermissions adminKey = "admin_permissions"
)

// AdminID returns the authenticated admin's public ID, or "" if the
// request context has none (i.e. RequireAdmin was not applied or the
// token was invalid).
func AdminID(ctx context.Context) string { v, _ := ctx.Value(adminID).(string); return v }

// AdminDBID returns the authenticated admin's numeric database ID (the
// JWT's "uid" claim), or 0 if unavailable. Use this for foreign keys like
// payroll_rule_sets.created_by - AdminID's public ID is not the right type
// for those columns.
func AdminDBID(ctx context.Context) uint64 { v, _ := ctx.Value(adminDBID).(uint64); return v }

// AdminPermissions returns the permission codes embedded in the admin's
// token at login time (see internal/admin/service.Service.issue).
func AdminPermissions(ctx context.Context) []string {
	v, _ := ctx.Value(adminPermissions).([]string)
	return v
}

// RequireAdmin verifies a Bearer JWT issued by internal/admin/service and
// carrying "typ":"admin". A customer-facing token (internal/auth) has no
// such claim, so it can never satisfy this check even though both token
// types are signed with the same secret.
func RequireAdmin(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(w, 401, "UNAUTHORIZED", "unauthorized", RequestIDFromContext(r.Context()), nil)
			return
		}
		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) { return secret, nil })
		if err != nil || !token.Valid {
			response.Fail(w, 401, "UNAUTHORIZED", "unauthorized", RequestIDFromContext(r.Context()), nil)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["typ"] != "admin" {
			response.Fail(w, 401, "UNAUTHORIZED", "unauthorized", RequestIDFromContext(r.Context()), nil)
			return
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			response.Fail(w, 401, "UNAUTHORIZED", "unauthorized", RequestIDFromContext(r.Context()), nil)
			return
		}
		var dbID uint64
		if uidF, ok := claims["uid"].(float64); ok {
			dbID = uint64(uidF)
		}
		var perms []string
		if raw, ok := claims["permissions"].([]any); ok {
			for _, p := range raw {
				if s, ok := p.(string); ok {
					perms = append(perms, s)
				}
			}
		}
		ctx := context.WithValue(r.Context(), adminID, sub)
		ctx = context.WithValue(ctx, adminDBID, dbID)
		ctx = context.WithValue(ctx, adminPermissions, perms)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission wraps a handler that RequireAdmin has already run for,
// rejecting the request with 403 unless the admin's token carries code.
func RequirePermission(code string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range AdminPermissions(r.Context()) {
			if p == code {
				next.ServeHTTP(w, r)
				return
			}
		}
		response.Fail(w, 403, "FORBIDDEN", "forbidden", RequestIDFromContext(r.Context()), nil)
	})
}
