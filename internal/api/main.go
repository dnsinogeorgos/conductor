package api

import (
	"github.com/dnsinogeorgos/conductor/internal/conductor"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// NewRouter creates a new chi router instance and initializes routes.
func NewRouter(cnd *conductor.Conductor) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)

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
