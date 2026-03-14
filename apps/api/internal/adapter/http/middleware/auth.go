package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	platformauth "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

type authUserContextKey string

const userContextKey authUserContextKey = "auth_user"

type TokenParser interface {
	ParseAccessToken(rawToken string) (platformauth.Claims, error)
}

type AuthUser struct {
	UserID string
	Role   identity.Role
}

type Authenticator struct {
	tokenParser TokenParser
}

func NewAuthenticator(tokenParser TokenParser) *Authenticator {
	return &Authenticator{tokenParser: tokenParser}
}

func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := observability.RequestIDFromContext(r.Context())
		tokenValue := bearerToken(r.Header.Get("Authorization"))
		if tokenValue == "" {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "missing bearer token",
			}})
			return
		}

		claims, err := a.tokenParser.ParseAccessToken(tokenValue)
		if err != nil {
			errorMessage := "invalid bearer token"
			if errors.Is(err, platformauth.ErrTokenExpired) {
				errorMessage = "token expired"
			}
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: errorMessage,
			}})
			return
		}

		authUser := AuthUser{
			UserID: claims.Subject,
			Role:   claims.Role,
		}
		next.ServeHTTP(w, r.WithContext(WithAuthUser(r.Context(), authUser)))
	})
}

func (a *Authenticator) RequireRole(requiredRole identity.Role, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := observability.RequestIDFromContext(r.Context())
		authUser, ok := AuthUserFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "authentication context missing",
			}})
			return
		}
		if authUser.Role != requiredRole {
			response.WriteError(w, http.StatusForbidden, "Forbidden", requestID, []response.ErrorItem{{
				Code:    errcode.Forbidden,
				Message: "insufficient role",
			}})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func WithAuthUser(ctx context.Context, authUser AuthUser) context.Context {
	return context.WithValue(ctx, userContextKey, authUser)
}

func AuthUserFromContext(ctx context.Context) (AuthUser, bool) {
	value, ok := ctx.Value(userContextKey).(AuthUser)
	if !ok {
		return AuthUser{}, false
	}
	return value, true
}

func bearerToken(rawAuthHeader string) string {
	parts := strings.Fields(strings.TrimSpace(rawAuthHeader))
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
