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

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// Authentication
	mux.HandleFunc("POST /api/v1/auth/login", auth.Login)
	mux.HandleFunc(
		"GET /api/v1/auth/me",
		access.RequireAuth(middleware.Me),
	)

	// Hosts
	mux.HandleFunc(
		"POST /api/v1/hosts",
		access.RequireAdmin(hosts.Create),
	)
	mux.HandleFunc(
		"GET /api/v1/hosts",
		access.RequireAuth(hosts.List),
	)
	mux.HandleFunc(
		"GET /api/v1/hosts/{id}",
		access.RequireAuth(hosts.GetByID),
	)
	mux.HandleFunc(
		"PATCH /api/v1/hosts/{id}",
		access.RequireAdmin(hosts.Update),
	)
	mux.HandleFunc(
		"DELETE /api/v1/hosts/{id}",
		access.RequireAdmin(hosts.Delete),
	)

	// Security policies
	mux.HandleFunc(
		"POST /api/v1/policies",
		access.RequireAdmin(policies.Create),
	)
	mux.HandleFunc(
		"GET /api/v1/policies",
		access.RequireAuth(policies.List),
	)
	mux.HandleFunc(
		"GET /api/v1/policies/{id}",
		access.RequireAuth(policies.GetByID),
	)
	mux.HandleFunc(
		"PATCH /api/v1/policies/{id}",
		access.RequireAdmin(policies.Update),
	)
	mux.HandleFunc(
		"DELETE /api/v1/policies/{id}",
		access.RequireAdmin(policies.Delete),
	)

	// Policy rules
	mux.HandleFunc(
		"POST /api/v1/policies/{id}/rules",
		access.RequireAdmin(policies.CreateRule),
	)
	mux.HandleFunc(
		"PATCH /api/v1/policies/{id}/rules/{ruleID}",
		access.RequireAdmin(policies.UpdateRule),
	)
	mux.HandleFunc(
		"DELETE /api/v1/policies/{id}/rules/{ruleID}",
		access.RequireAdmin(policies.DeleteRule),
	)

	// Snapshots and audits
	mux.HandleFunc(
		"POST /api/v1/hosts/{id}/snapshots",
		access.RequireAdmin(audits.Import),
	)
	mux.HandleFunc(
		"GET /api/v1/hosts/{id}/audits",
		access.RequireAuth(audits.ListByHost),
	)
	mux.HandleFunc(
		"GET /api/v1/audits/{id}",
		access.RequireAuth(audits.GetByID),
	)
	mux.HandleFunc(
		"GET /api/v1/audits/{id}/snapshot",
		access.RequireAuth(audits.GetSnapshot),
	)
	mux.HandleFunc(
		"GET /api/v1/audits/{id}/summary",
		access.RequireAuth(audits.GetSummary),
	)

	// Baselines
	mux.HandleFunc(
		"PUT /api/v1/hosts/{id}/baseline",
		access.RequireAdmin(baselines.Set),
	)
	mux.HandleFunc(
		"GET /api/v1/hosts/{id}/baseline",
		access.RequireAuth(baselines.Get),
	)
	mux.HandleFunc(
		"DELETE /api/v1/hosts/{id}/baseline",
		access.RequireAdmin(baselines.Clear),
	)

	// Reports
	mux.HandleFunc(
		"GET /api/v1/reports/hosts-with-policies",
		access.RequireAuth(reports.HostsWithPolicies),
	)
	mux.HandleFunc(
		"GET /api/v1/reports/hosts-with-baselines",
		access.RequireAuth(reports.HostsWithBaselines),
	)
	mux.HandleFunc(
		"GET /api/v1/reports/audits-with-hosts",
		access.RequireAuth(reports.AuditsWithHosts),
	)
	mux.HandleFunc(
		"GET /api/v1/reports/policy-rules",
		access.RequireAuth(reports.PolicyRules),
	)

	return mux
}
