package handlers

import (
	"errors"
	"net/http"

	"web-api/back/api/response"
	"web-api/back/models"
	"web-api/back/services"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input models.LoginInput

	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.service.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized,
				"invalid username or password")
			return
		}

		response.Error(w, http.StatusInternalServerError,
			"internal server error")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	response.JSON(w, http.StatusOK, result)
}
