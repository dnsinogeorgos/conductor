package api

import (
	"net/http"

	"github.com/dnsinogeorgos/conductor/internal/conductor"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// NewRouter creates a new chi router instance and initializes routes.
func NewRouter(cnd *conductor.Conductor) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)

	r.Get("/cnd", FilesystemResource{cnd}.FilesystemGet)

	r.Mount("/casts", CastsResource{cnd}.Routes())
	r.Mount("/replicas", ReplicasResource{cnd}.Routes())

	return r
}

// Routes creates a REST router for the casts resource.
func (cr CastsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", cr.CastsGet)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", cr.CastsIdGet)
		r.Post("/", cr.CastsIdPost)
		r.Delete("/", cr.CastsIdDelete)
	})

	return r
}

// Routes creates a REST router for the replicas resource.
func (rr ReplicasResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{castId}", rr.ReplicasCastIdGet)
	r.Get("/{castId}/", rr.ReplicasCastIdGet)
	r.Route("/{castId}/{id}", func(r chi.Router) {
		r.Get("/", rr.ReplicasCastIdIdGet)
		r.Post("/", rr.ReplicasCastIdIdPost)
		r.Delete("/", rr.ReplicasCastIdIdDelete)
	})

	return r
}

type FilesystemResource struct {
	*conductor.Conductor
}

// FilesystemGet returns the filesystem state object
func (fr FilesystemResource) FilesystemGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, fr)
}
