package api

import (
	"net/http"

	"github.com/dnsinogeorgos/conductor/internal/zfs"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/go-chi/render"
)

// NewRouter creates a new chi router instance and initializes routes.
func NewRouter(name string, fs *zfs.ZFS) *chi.Mux {

	r := chi.NewRouter()
	l := httplog.NewLogger(name, httplog.Options{
		LogLevel: "warn",
		JSON:     false,
		Concise:  false,
	})

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(httplog.RequestLogger(l))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)

	r.Get("/fs", FilesystemResource{fs}.FilesystemGet)

	r.Mount("/casts", CastsResource{fs}.Routes())
	r.Mount("/replicas", ReplicasResource{fs}.Routes())

	return r
}

// Routes creates a REST router for the casts resource.
func (cs CastsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", cs.CastsGet)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", cs.CastsIdGet)
		r.Post("/", cs.CastsIdPost)
		r.Delete("/", cs.CastsIdDelete)
	})

	return r
}

// Routes creates a REST router for the replicas resource.
func (rs ReplicasResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{castId}", rs.ReplicasCastIdGet)
	r.Get("/{castId}/", rs.ReplicasCastIdGet)
	r.Route("/{castId}/{id}", func(r chi.Router) {
		r.Get("/", rs.ReplicasCastIdIdGet)
		r.Post("/", rs.ReplicasCastIdIdPost)
		r.Delete("/", rs.ReplicasCastIdIdDelete)
	})

	return r
}

type FilesystemResource struct {
	*zfs.ZFS
}

// FilesystemGet returns the filesystem state object
func (fs FilesystemResource) FilesystemGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, fs)
}
