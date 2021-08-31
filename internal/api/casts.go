package api

import (
	"net/http"

	"github.com/dnsinogeorgos/conductor/internal/conductor"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// CastsResource embeds the conductor type to allow extending it's interface with
// handlers
type CastsResource struct {
	*conductor.Conductor
}

// CastResponse describes the API cast response object
type CastResponse struct {
	Id        string `json:"id"`
	Timestamp string `json:"timestamp"`
}

// CastsIdDelete deletes a cast from the filesystem.
func (cr CastsResource) CastsIdDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := cr.DeleteCast(id)
	if err != nil {
		switch e := err.(type) {
		case conductor.CastNotEmpty:
			w.WriteHeader(http.StatusConflict)
			return
		case conductor.CastNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			_ = e
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// CastsIdGet gets a cast from the filesystem.
func (cr CastsResource) CastsIdGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cast, err := cr.GetCast(id)
	if err != nil {
		switch e := err.(type) {
		case conductor.CastNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			_ = e
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	result := CastResponse{
		Id:        cast.Id,
		Timestamp: cast.Timestamp,
	}
	render.JSON(w, r, result)
}

// CastsIdPost creates a cast on the filesystem.
func (cr CastsResource) CastsIdPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cast, err := cr.CreateCast(id)
	if err != nil {
		switch e := err.(type) {
		case conductor.CastAlreadyExistsError:
			w.WriteHeader(http.StatusConflict)
			return
		default:
			_ = e
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	result := CastResponse{
		Id:        cast.Id,
		Timestamp: cast.Timestamp,
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, result)
}

// CastsGet returns a list of the casts on the filesystem.
func (cr CastsResource) CastsGet(w http.ResponseWriter, r *http.Request) {

	casts := cr.ListCasts()
	result := make([]CastResponse, 0)
	for _, cast := range casts {
		item := CastResponse{
			Id:        cast.Id,
			Timestamp: cast.Timestamp,
		}
		result = append(result, item)
	}
	render.JSON(w, r, result)
}
