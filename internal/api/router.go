package api

import (
	"net/http"

	"web-api/internal/api/handlers"
	"web-api/internal/api/middleware"
	"web-api/internal/api/response"
)

func NewRouter(
	hosts *handlers.HostHandler,
	policies *handlers.PolicyHandler,
	audits *handlers.AuditHandler,
	baselines *handlers.BaselineHandler,
	reports *handlers.ReportHandler,
	auth *handlers.AuthHandler,
	access *middleware.AuthMiddleware,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// Authentication
	mux.HandleFunc("POST /api/v1/auth/login", auth.Login)
	mux.HandleFunc("GET /api/v1/auth/me",
		access.RequireAuth(middleware.Me))

	// Hosts
	mux.HandleFunc("POST /api/v1/hosts",
		access.RequireAdmin(hosts.Create))
	mux.HandleFunc("GET /api/v1/hosts", hosts.List)
	mux.HandleFunc("GET /api/v1/hosts/{id}", hosts.GetByID)
	mux.HandleFunc("PATCH /api/v1/hosts/{id}",
		access.RequireAdmin(hosts.Update))
	mux.HandleFunc("DELETE /api/v1/hosts/{id}",
		access.RequireAdmin(hosts.Delete))

	// Policies
	mux.HandleFunc("POST /api/v1/policies",
		access.RequireAdmin(policies.Create))
	mux.HandleFunc("GET /api/v1/policies", policies.List)
	mux.HandleFunc("GET /api/v1/policies/{id}", policies.GetByID)
	mux.HandleFunc("PATCH /api/v1/policies/{id}",
		access.RequireAdmin(policies.Update))
	mux.HandleFunc("DELETE /api/v1/policies/{id}",
		access.RequireAdmin(policies.Delete))

	// Policy rules
	mux.HandleFunc("POST /api/v1/policies/{id}/rules",
		access.RequireAdmin(policies.CreateRule))
	mux.HandleFunc("PATCH /api/v1/policies/{id}/rules/{ruleID}",
		access.RequireAdmin(policies.UpdateRule))
	mux.HandleFunc("DELETE /api/v1/policies/{id}/rules/{ruleID}",
		access.RequireAdmin(policies.DeleteRule))

	// Snapshots and audits
	mux.HandleFunc("POST /api/v1/hosts/{id}/snapshots",
		access.RequireAdmin(audits.Import))
	mux.HandleFunc("GET /api/v1/hosts/{id}/audits",
		audits.ListByHost)
	mux.HandleFunc("GET /api/v1/audits/{id}", audits.GetByID)
	mux.HandleFunc("GET /api/v1/audits/{id}/snapshot",
		audits.GetSnapshot)
	mux.HandleFunc("GET /api/v1/audits/{id}/summary",
		access.RequireAuth(audits.GetSummary))

	// Baseline
	mux.HandleFunc("PUT /api/v1/hosts/{id}/baseline",
		access.RequireAdmin(baselines.Set))
	mux.HandleFunc("GET /api/v1/hosts/{id}/baseline",
		baselines.Get)
	mux.HandleFunc("DELETE /api/v1/hosts/{id}/baseline",
		access.RequireAdmin(baselines.Clear))

	// Reports
	mux.HandleFunc("GET /api/v1/reports/hosts-with-policies",
		access.RequireAuth(reports.HostsWithPolicies))
	mux.HandleFunc("GET /api/v1/reports/hosts-with-baselines",
		access.RequireAuth(reports.HostsWithBaselines))
	mux.HandleFunc("GET /api/v1/reports/audits-with-hosts",
		access.RequireAuth(reports.AuditsWithHosts))
	mux.HandleFunc("GET /api/v1/reports/policy-rules",
		access.RequireAuth(reports.PolicyRules))

	return mux
}
