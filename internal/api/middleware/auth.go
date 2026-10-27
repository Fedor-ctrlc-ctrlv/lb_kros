package middleware

import (
	"context"
	"net/http"
	"strings"

	"web-api/internal/api/response"
	"web-api/internal/models"
	"web-api/internal/services"
)

type userContextKey struct{}

type AuthMiddleware struct {
	service *services.AuthService
}

func NewAuthMiddleware(
	service *services.AuthService,
) *AuthMiddleware {
	return &AuthMiddleware{service: service}
}

func CurrentUser(ctx context.Context) (*models.CurrentUser, bool) {
	user, ok := ctx.Value(userContextKey{}).(*models.CurrentUser)
	return user, ok
}

func (m *AuthMiddleware) RequireAuth(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		parts := strings.Fields(header)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			response.Error(w, http.StatusUnauthorized,
				"missing or invalid bearer token")
			return
		}

		user, err := m.service.Authenticate(
			r.Context(), parts[1],
		)
		if err != nil {
			if err == services.ErrInvalidToken {
				response.Error(w, http.StatusUnauthorized,
					"invalid or expired token")
			} else {
				response.Error(w, http.StatusInternalServerError,
					"internal server error")
			}
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userContextKey{},
			user,
		)

		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) RequireAdmin(
	next http.HandlerFunc,
) http.HandlerFunc {
	return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUser(r.Context())
		if !ok {
			response.Error(w, http.StatusUnauthorized,
				"authentication required")
			return
		}

		if user.Role != models.RoleAdmin {
			response.Error(w, http.StatusForbidden,
				"admin role required")
			return
		}

		next(w, r)
	})
}

func Me(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized,
			"authentication required")
		return
	}

	response.JSON(w, http.StatusOK, user)
}
