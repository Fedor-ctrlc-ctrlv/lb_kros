package api

import (
	"net/http"

	"web-api/back/api/handlers"
	"web-api/back/api/response"
)

func NewRouter(
	hosts *handlers.HostHandler,
	policies *handlers.PolicyHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	mux.HandleFunc("POST /api/v1/hosts", hosts.Create)
	mux.HandleFunc("GET /api/v1/hosts", hosts.List)
	mux.HandleFunc("GET /api/v1/hosts/{id}", hosts.GetByID)
	mux.HandleFunc("PATCH /api/v1/hosts/{id}", hosts.Update)
	mux.HandleFunc("DELETE /api/v1/hosts/{id}", hosts.Delete)

	mux.HandleFunc("POST /api/v1/policies", policies.Create)
	mux.HandleFunc("GET /api/v1/policies", policies.List)
	mux.HandleFunc("GET /api/v1/policies/{id}", policies.GetByID)
	mux.HandleFunc("PATCH /api/v1/policies/{id}", policies.Update)
	mux.HandleFunc("DELETE /api/v1/policies/{id}", policies.Delete)

	mux.HandleFunc("POST /api/v1/policies/{id}/rules", policies.CreateRule)
	mux.HandleFunc("PATCH /api/v1/policies/{id}/rules/{ruleID}", policies.UpdateRule)
	mux.HandleFunc("DELETE /api/v1/policies/{id}/rules/{ruleID}", policies.DeleteRule)

	return mux
}
