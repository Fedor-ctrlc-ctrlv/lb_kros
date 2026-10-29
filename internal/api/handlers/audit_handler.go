package handlers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"web-api/internal/api/response"
	"web-api/internal/models"
	"web-api/internal/services"
)

type AuditHandler struct {
	service *services.AuditService
}

func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func writeAuditError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidSnapshot),
		errors.Is(err, services.ErrInvalidHostID):
		response.Error(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, models.ErrSnapshotHostMismatch):
		response.Error(w, http.StatusConflict, err.Error())

	case errors.Is(err, models.ErrHostNotFound),
		errors.Is(err, models.ErrAuditNotFound):
		response.Error(w, http.StatusNotFound, err.Error())

	default:
		response.Error(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func (h *AuditHandler) Import(w http.ResponseWriter, r *http.Request) {
	hostID, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid host id")
		return
	}

	const maxSnapshotSize = 32 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxSnapshotSize)
	defer r.Body.Close()

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			response.Error(w, http.StatusRequestEntityTooLarge,
				"snapshot is too large")
		} else {
			response.Error(w, http.StatusBadRequest,
				"cannot read snapshot")
		}
		return
	}

	result, err := h.service.Import(r.Context(), hostID, raw)
	if err != nil {
		writeAuditError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

/*
func (h *AuditHandler) ListByHost(
	w http.ResponseWriter, r *http.Request,
) {
	hostID, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid host id")
		return
	}

	result, err := h.service.ListByHost(r.Context(), hostID)
	if err != nil {
		writeAuditError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}*/

func (h *AuditHandler) ListByHost(
	w http.ResponseWriter, r *http.Request,
) {
	hostID, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid host id")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	// Если даты не переданы — возвращаем все аудиты хоста.
	if fromStr == "" && toStr == "" {
		result, err := h.service.ListByHost(r.Context(), hostID)
		if err != nil {
			writeAuditError(w, err)
			return
		}

		response.JSON(w, http.StatusOK, result)
		return
	}

	if fromStr == "" || toStr == "" {
		response.Error(w, http.StatusBadRequest,
			"both from and to are required")
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest,
			"invalid from date")
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest,
			"invalid to date")
		return
	}

	if !from.Before(to) {
		response.Error(w, http.StatusBadRequest,
			"from must be before to")
		return
	}

	result, err := h.service.ListByHostPeriod(
		r.Context(),
		hostID,
		from,
		to,
	)
	if err != nil {
		writeAuditError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *AuditHandler) GetByID(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid audit id")
		return
	}

	result, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeAuditError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *AuditHandler) GetSnapshot(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid audit id")
		return
	}

	raw, err := h.service.GetSnapshot(r.Context(), id)
	if err != nil {
		writeAuditError(w, err)
		return
	}
	if len(raw) == 0 {
		response.Error(w, http.StatusNotFound,
			"snapshot not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (h *AuditHandler) GetSummary(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid audit id")
		return
	}

	result, err := h.service.GetSummary(r.Context(), id)
	if err != nil {
		writeAuditError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}
