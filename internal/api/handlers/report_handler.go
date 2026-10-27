package handlers

import (
	"errors"
	"net/http"

	"web-api/internal/api/response"
	"web-api/internal/services"
)

type ReportHandler struct {
	service *services.ReportService
}

func NewReportHandler(
	service *services.ReportService,
) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) HostsWithPolicies(
	w http.ResponseWriter,
	r *http.Request,
) {
	result, err := h.service.HostsWithPolicies(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError,
			"failed to get report")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *ReportHandler) HostsWithBaselines(
	w http.ResponseWriter,
	r *http.Request,
) {
	result, err := h.service.HostsWithBaselines(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError,
			"failed to get report")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *ReportHandler) AuditsWithHosts(
	w http.ResponseWriter,
	r *http.Request,
) {
	result, err := h.service.AuditsWithHosts(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError,
			"failed to get report")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *ReportHandler) PolicyRules(
	w http.ResponseWriter,
	r *http.Request,
) {
	severity := r.URL.Query().Get("severity")

	result, err := h.service.PolicyRules(
		r.Context(), severity,
	)
	if err != nil {
		if errors.Is(err, services.ErrInvalidSeverity) {
			response.Error(w, http.StatusBadRequest,
				"severity must be LOW, MEDIUM, HIGH or CRITICAL")
			return
		}

		response.Error(w, http.StatusInternalServerError,
			"failed to get report")
		return
	}

	response.JSON(w, http.StatusOK, result)
}
