package api

import (
	"net/http"

	"web-api/back/api/handlers"
	"web-api/back/api/response"
)

func NewRouter(hosts *handlers.HostHandler) http.Handler {
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

	return mux
}
