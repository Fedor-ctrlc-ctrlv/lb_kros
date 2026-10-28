package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"web-api/internal/api/response"
	"web-api/internal/models"
	"web-api/internal/services"

	"github.com/jackc/pgx/v5/pgconn"
)

type PolicyHandler struct {
	service *services.PolicyService
}

func NewPolicyHandler(service *services.PolicyService) *PolicyHandler {
	return &PolicyHandler{service: service}
}

func parseID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, services.ErrInvalidPolicy
	}
	return id, nil
}

func writePolicyError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, services.ErrInvalidPolicy),
		errors.Is(err, services.ErrInvalidRule):
		response.Error(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, models.ErrPolicyNotFound),
		errors.Is(err, models.ErrRuleNotFound):
		response.Error(w, http.StatusNotFound, err.Error())

	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		response.Error(w, http.StatusConflict, "policy or rule already exists")

	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		response.Error(w, http.StatusConflict, "policy is referenced by another entity")

	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *PolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input services.CreatePolicyInput
	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.service.Create(r.Context(), input)
	if err != nil {
		writePolicyError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *PolicyHandler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.List(r.Context())
	if err != nil {
		writePolicyError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *PolicyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writePolicyError(w, err)
		return
	}

	result, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writePolicyError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *PolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writePolicyError(w, err)
		return
	}

	var input services.UpdatePolicyInput
	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		writePolicyError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *PolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writePolicyError(w, err)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writePolicyError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PolicyHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	policyID, err := parseID(r, "id")
	if err != nil {
		writePolicyError(w, err)
		return
	}

	var input services.CreateRuleInput
	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.service.CreateRule(r.Context(), policyID, input)
	if err != nil {
		writePolicyError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, result)
}

func (h *PolicyHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	policyID, err := parseID(r, "id")
	if err != nil {
		writePolicyError(w, err)
		return
	}
	ruleID, err := parseID(r, "ruleID")
	if err != nil {
		writePolicyError(w, err)
		return
	}

	var input services.UpdateRuleInput
	if err := decodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.service.UpdateRule(
		r.Context(), policyID, ruleID, input,
	)
	if err != nil {
		writePolicyError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *PolicyHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	policyID, err := parseID(r, "id")
	if err != nil {
		writePolicyError(w, err)
		return
	}
	ruleID, err := parseID(r, "ruleID")
	if err != nil {
		writePolicyError(w, err)
		return
	}

	if err := h.service.DeleteRule(
		r.Context(), policyID, ruleID,
	); err != nil {
		writePolicyError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
