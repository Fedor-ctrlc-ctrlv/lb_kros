package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"web-api/internal/api/response"
	"web-api/internal/services"

	"github.com/jackc/pgx/v5/pgconn"
)

type HostHandler struct {
	service *services.HostService
}

func NewHostHandler(service *services.HostService) *HostHandler {
	return &HostHandler{service: service}
}

func hostID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, services.ErrInvalidHostID
	}
	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(
		io.LimitReader(r.Body, 1<<20),
	)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("expected one JSON object")
	}

	return nil
}

func writeHostError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, services.ErrInvalidHost),
		errors.Is(err, services.ErrInvalidHostID):
		response.Error(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, services.ErrHostNotFound):
		response.Error(w, http.StatusNotFound, "host not found")

	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		response.Error(w, http.StatusConflict, "host name already exists")

	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *HostHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input services.CreateHostInput

	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	host, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeHostError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, host)
}

func (h *HostHandler) List(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.service.List(r.Context())
	if err != nil {
		writeHostError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, hosts)
}

func (h *HostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := hostID(r)
	if err != nil {
		writeHostError(w, err)
		return
	}

	host, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeHostError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, host)
}

func (h *HostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := hostID(r)
	if err != nil {
		writeHostError(w, err)
		return
	}

	var input services.UpdateHostInput
	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	host, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		writeHostError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, host)
}

func (h *HostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := hostID(r)
	if err != nil {
		writeHostError(w, err)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeHostError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
