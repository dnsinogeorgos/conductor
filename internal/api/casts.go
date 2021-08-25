package api

import (
	"net/http"

	"github.com/go-chi/httplog"

	"github.com/dnsinogeorgos/conductor/internal/zfs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type CastsResource struct {
	*zfs.ZFS
}

type CastResponse struct {
	Id   string `json:"id"`
	Date string `json:"date"`
}

// CastsIdDelete deletes a cast from the filesystem.
func (cs CastsResource) CastsIdDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := cs.DeleteCast(id)
	if err != nil {
		switch e := err.(type) {
		case zfs.CastContainsReplicasError:
			w.WriteHeader(http.StatusConflict)
			return
		case zfs.CastNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			l := httplog.LogEntry(r.Context())
			l.Error().Msgf(e.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// CastsIdGet gets a cast from the filesystem.
func (cs CastsResource) CastsIdGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cast, err := cs.GetCast(id)
	if err != nil {
		switch e := err.(type) {
		case zfs.CastNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			l := httplog.LogEntry(r.Context())
			l.Error().Msgf(e.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	result := CastResponse{
		Id:   cast.Id,
		Date: cast.Date,
	}
	render.JSON(w, r, result)
}

// CastsIdPost creates a cast on the filesystem.
func (cs CastsResource) CastsIdPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cast, err := cs.CreateCast(id)
	if err != nil {
		switch e := err.(type) {
		case zfs.CastAlreadyExistsError:
			w.WriteHeader(http.StatusConflict)
			return
		default:
			l := httplog.LogEntry(r.Context())
			l.Error().Msgf(e.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	result := CastResponse{
		Id:   cast.Id,
		Date: cast.Date,
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, result)
}

// CastsGet returns a list of the casts on the filesystem.
func (cs CastsResource) CastsGet(w http.ResponseWriter, r *http.Request) {

	casts := cs.ListCasts()
	result := make([]CastResponse, 0)
	for _, cast := range casts {
		item := CastResponse{
			Id:   cast.Id,
			Date: cast.Date,
		}
		result = append(result, item)
	}
	render.JSON(w, r, result)
}
