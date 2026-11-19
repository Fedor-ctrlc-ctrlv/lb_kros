package analysis

import (
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"

	"web-api/internal/api/middleware"
	"web-api/internal/api/response"
)

func Register(
	base http.Handler,
	service *Service,
	access *middleware.AuthMiddleware,
) http.Handler {
	h := &handler{service: service}
	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /api/v1/audits/{id}/analyses",
		access.RequireAdmin(h.create),
	)
	mux.HandleFunc(
		"GET /api/v1/analyses/{id}",
		access.RequireAuth(h.get),
	)
	mux.HandleFunc(
		"GET /api/v1/analyses/{id}/coverage",
		access.RequireAuth(h.coverage),
	)
	mux.HandleFunc(
		"GET /api/v1/analyses/{id}/changes",
		access.RequireAuth(h.changes),
	)
	mux.HandleFunc(
		"GET /api/v1/analyses/{id}/checks",
		access.RequireAuth(h.checks),
	)
	mux.HandleFunc(
		"PUT /api/v1/policies/{id}/engine",
		access.RequireAdmin(h.putPolicy),
	)
	mux.HandleFunc(
		"GET /api/v1/policies/{id}/engine",
		access.RequireAuth(h.getPolicy),
	)

	mux.Handle("/", base)
	return mux
}

type handler struct {
	service *Service
}

func requestID(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalid):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrConflict):
		response.Error(w, http.StatusConflict, err.Error())
	default:
		log.Printf("analysis request failed: %v", err)
		response.Error(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func (h *handler) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}

	run, err := h.service.Create(r.Context(), id)
	if err != nil {
		failure(w, err)
		return
	}

	w.Header().Set(
		"Location",
		"/api/v1/analyses/"+strconv.FormatInt(run.ID, 10),
	)
	response.JSON(w, http.StatusCreated, run)
}

func (h *handler) get(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}

	run, err := h.service.Get(r.Context(), id)
	if err != nil {
		failure(w, err)
		return
	}
	response.JSON(w, http.StatusOK, run)
}

func (h *handler) coverage(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}

	raw, err := h.service.Coverage(r.Context(), id)
	if err != nil {
		failure(w, err)
		return
	}
	response.JSON(w, http.StatusOK, raw)
}

func (h *handler) changes(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.details(w, r, false)
}

func (h *handler) checks(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.details(w, r, true)
}

func (h *handler) details(
	w http.ResponseWriter,
	r *http.Request,
	checks bool,
) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}

	query := r.URL.Query()
	filter := Filter{
		Section: query.Get("section"),
		Kind:    query.Get("kind"),
		Status:  query.Get("status"),
		Phase:   query.Get("phase"),
		Limit:   50,
	}

	var err error
	if query.Has("limit") {
		filter.Limit, err = strconv.Atoi(query.Get("limit"))
	}
	if err != nil || filter.Limit < 1 || filter.Limit > 200 {
		response.Error(w, http.StatusBadRequest, "limit must be 1..200")
		return
	}

	if query.Has("offset") {
		filter.Offset, err = strconv.Atoi(query.Get("offset"))
	}
	if err != nil || filter.Offset < 0 {
		response.Error(w, http.StatusBadRequest, "invalid offset")
		return
	}

	if (checks && filter.Kind != "") ||
		(!checks && (filter.Status != "" || filter.Phase != "")) {
		response.Error(
			w,
			http.StatusBadRequest,
			"filter is not applicable to this endpoint",
		)
		return
	}

	page, err := h.service.Details(r.Context(), id, checks, filter)
	if err != nil {
		failure(w, err)
		return
	}
	response.JSON(w, http.StatusOK, page)
}

func (h *handler) getPolicy(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}

	raw, err := h.service.GetPolicy(r.Context(), id)
	if err != nil {
		failure(w, err)
		return
	}
	response.JSON(w, http.StatusOK, raw)
}

func (h *handler) putPolicy(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := requestID(w, r)
	if !ok {
		return
	}

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		response.Error(
			w,
			http.StatusUnsupportedMediaType,
			"Content-Type must be application/json",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
	defer r.Body.Close()

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			response.Error(
				w,
				http.StatusRequestEntityTooLarge,
				"policy is too large",
			)
		} else {
			response.Error(w, http.StatusBadRequest, "cannot read body")
		}
		return
	}

	saved, err := h.service.PutPolicy(r.Context(), id, raw)
	if err != nil {
		failure(w, err)
		return
	}
	response.JSON(w, http.StatusOK, saved)
}
