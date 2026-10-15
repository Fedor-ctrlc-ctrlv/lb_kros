package handlers

import (
	"errors"
	"net/http"

	"web-api/back/api/response"
	"web-api/back/repository"
	"web-api/back/services"
)

type BaselineHandler struct {
	service *services.BaselineService
}

func NewBaselineHandler(
	service *services.BaselineService,
) *BaselineHandler {
	return &BaselineHandler{service: service}
}

func writeBaselineError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidBaseline):
		response.Error(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, repository.ErrHostNotFound),
		errors.Is(err, repository.ErrAuditNotFound),
		errors.Is(err, repository.ErrBaselineNotSet):
		response.Error(w, http.StatusNotFound, err.Error())

	case errors.Is(err, repository.ErrAuditHostMismatch),
		errors.Is(err, repository.ErrAuditNotCompleted):
		response.Error(w, http.StatusConflict, err.Error())

	default:
		response.Error(w, http.StatusInternalServerError,
			"internal server error")
	}
}

func (h *BaselineHandler) Set(
	w http.ResponseWriter,
	r *http.Request,
) {
	hostID, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid host id")
		return
	}

	var input services.SetBaselineInput
	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.service.Set(r.Context(), hostID, input)
	if err != nil {
		writeBaselineError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *BaselineHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	hostID, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid host id")
		return
	}

	result, err := h.service.Get(r.Context(), hostID)
	if err != nil {
		writeBaselineError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *BaselineHandler) Clear(
	w http.ResponseWriter,
	r *http.Request,
) {
	hostID, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid host id")
		return
	}

	if err := h.service.Clear(r.Context(), hostID); err != nil {
		writeBaselineError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
