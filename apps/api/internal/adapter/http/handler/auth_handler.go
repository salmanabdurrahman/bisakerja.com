package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	domain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

// AuthHandler represents auth handler.
type AuthHandler struct {
	service *identityapp.Service
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// NewAuthHandler creates a new auth handler instance.
func NewAuthHandler(service *identityapp.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register handles register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	var request registerRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	user, err := h.service.Register(r.Context(), identityapp.RegisterInput{
		Email:    request.Email,
		Password: request.Password,
		Name:     request.Name,
	})
	if err != nil {
		switch {
		case errors.Is(err, identityapp.ErrInvalidEmail):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "email",
				Code:    errcode.InvalidEmail,
				Message: "email must be valid and non-empty",
			}})
		case errors.Is(err, identityapp.ErrInvalidPassword):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "password",
				Code:    errcode.InvalidPassword,
				Message: "password must be at least 8 chars and include uppercase + digit",
			}})
		case errors.Is(err, identityapp.ErrInvalidName):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "name",
				Code:    errcode.InvalidName,
				Message: "name must be between 2 and 100 characters",
			}})
		case errors.Is(err, domain.ErrEmailAlreadyRegistered):
			response.WriteError(w, http.StatusConflict, "Conflict", requestID, []response.ErrorItem{{
				Code:    errcode.EmailAlreadyRegistered,
				Message: "email is already registered",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to register user",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, "User registered", requestID, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

// Login handles login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	var request loginRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	tokens, err := h.service.Login(r.Context(), identityapp.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, identityapp.ErrInvalidEmail):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "email",
				Code:    errcode.InvalidEmail,
				Message: "email must be valid",
			}})
		case errors.Is(err, identityapp.ErrInvalidPassword):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "password",
				Code:    errcode.InvalidPassword,
				Message: "password is required",
			}})
		case errors.Is(err, domain.ErrInvalidCredentials):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.InvalidCredentials,
				Message: "invalid email or password",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to login",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Login successful", requestID, map[string]any{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    tokens.TokenType,
		"expires_in":    tokens.ExpiresIn,
	})
}

// Refresh handles refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	var request refreshRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}
	if strings.TrimSpace(request.RefreshToken) == "" {
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "refresh_token",
			Code:    errcode.BadRequest,
			Message: "refresh_token is required",
		}})
		return
	}

	token, err := h.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, identityapp.ErrInvalidToken):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "invalid refresh token",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to refresh token",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Token refreshed", requestID, map[string]any{
		"access_token": token.AccessToken,
		"token_type":   token.TokenType,
		"expires_in":   token.ExpiresIn,
	})
}

// Me handles me.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	profile, err := h.service.GetProfile(r.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to load user profile",
		}})
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Profile retrieved", requestID, map[string]any{
		"id":                 profile.ID,
		"email":              profile.Email,
		"name":               profile.Name,
		"role":               profile.Role,
		"is_premium":         profile.IsPremium,
		"premium_expired_at": profile.PremiumExpiredAt,
		"subscription_state": profile.SubscriptionState,
	})
}
