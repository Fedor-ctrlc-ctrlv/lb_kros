package api

import (
	"net/http"

	"web-api/back/api/handlers"
	"web-api/back/api/response"
)

func NewRouter(
	hosts *handlers.HostHandler,
	policies *handlers.PolicyHandler,
	audits *handlers.AuditHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// Hosts
	mux.HandleFunc("POST /api/v1/hosts", hosts.Create)
	mux.HandleFunc("GET /api/v1/hosts", hosts.List)
	mux.HandleFunc("GET /api/v1/hosts/{id}", hosts.GetByID)
	mux.HandleFunc("PATCH /api/v1/hosts/{id}", hosts.Update)
	mux.HandleFunc("DELETE /api/v1/hosts/{id}", hosts.Delete)

	// Policies
	mux.HandleFunc("POST /api/v1/policies", policies.Create)
	mux.HandleFunc("GET /api/v1/policies", policies.List)
	mux.HandleFunc("GET /api/v1/policies/{id}", policies.GetByID)
	mux.HandleFunc("PATCH /api/v1/policies/{id}", policies.Update)
	mux.HandleFunc("DELETE /api/v1/policies/{id}", policies.Delete)

	// Policy rules
	mux.HandleFunc("POST /api/v1/policies/{id}/rules", policies.CreateRule)
	mux.HandleFunc("PATCH /api/v1/policies/{id}/rules/{ruleID}", policies.UpdateRule)
	mux.HandleFunc("DELETE /api/v1/policies/{id}/rules/{ruleID}", policies.DeleteRule)

	// Snapshots and audits
	mux.HandleFunc("POST /api/v1/hosts/{id}/snapshots", audits.Import)
	mux.HandleFunc("GET /api/v1/hosts/{id}/audits", audits.ListByHost)
	mux.HandleFunc("GET /api/v1/audits/{id}", audits.GetByID)
	mux.HandleFunc("GET /api/v1/audits/{id}/snapshot", audits.GetSnapshot)
	mux.HandleFunc("GET /api/v1/audits/{id}/summary", audits.GetSummary)

	return mux
}
