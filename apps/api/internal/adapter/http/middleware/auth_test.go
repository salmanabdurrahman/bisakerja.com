package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	platformauth "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

type fakeTokenParser struct {
	parse func(rawToken string) (platformauth.Claims, error)
}

func (f fakeTokenParser) ParseAccessToken(rawToken string) (platformauth.Claims, error) {
	if f.parse != nil {
		return f.parse(rawToken)
	}
	return platformauth.Claims{}, errors.New("token parser is not configured")
}

func TestRequireAuth_MissingBearerToken_ReturnsUnauthorized(t *testing.T) {
	authenticator := NewAuthenticator(fakeTokenParser{})
	handler := observability.RequestID(authenticator.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "UNAUTHORIZED") {
		t.Fatalf("expected UNAUTHORIZED code, got body: %s", response.Body.String())
	}
}

func TestRequireAuth_ValidToken_SetsContextUser(t *testing.T) {
	authenticator := NewAuthenticator(fakeTokenParser{
		parse: func(rawToken string) (platformauth.Claims, error) {
			if rawToken != "token-1" {
				return platformauth.Claims{}, platformauth.ErrInvalidToken
			}
			return platformauth.Claims{
				Role: identity.RoleUser,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "usr_1",
				},
			}, nil
		},
	})

	handler := observability.RequestID(authenticator.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := AuthUserFromContext(r.Context())
		if !ok {
			t.Fatal("expected auth user in context")
		}
		if authUser.Role != identity.RoleUser {
			t.Fatalf("expected role user, got %s", authUser.Role)
		}
		if authUser.UserID != "usr_1" {
			t.Fatalf("expected user id usr_1, got %s", authUser.UserID)
		}
		_, _ = io.WriteString(w, authUser.UserID)
	})))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer token-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestRequireRole_EnforcesAdminRole(t *testing.T) {
	authenticator := NewAuthenticator(fakeTokenParser{})
	protected := authenticator.RequireRole(identity.RoleAdmin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
	request = request.WithContext(WithAuthUser(request.Context(), AuthUser{
		UserID: "usr_1",
		Role:   identity.RoleUser,
	}))
	response := httptest.NewRecorder()
	observability.RequestID(protected).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "FORBIDDEN") {
		t.Fatalf("expected FORBIDDEN code, got body: %s", response.Body.String())
	}
}
