package api

import (
	"net/http"

	"github.com/go-chi/httplog"

	"github.com/dnsinogeorgos/conductor/internal/zfs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ReplicasResource struct {
	*zfs.ZFS
}

type ReplicaResponse struct {
	Id     string `json:"id"`
	CastId string `json:"castId"`
	Port   uint16 `json:"port"`
}

// ReplicasCastIdIdDelete deletes a replica from the provided cast.
func (rs ReplicasResource) ReplicasCastIdIdDelete(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")
	id := chi.URLParam(r, "id")

	cast, err := rs.GetCast(castId)
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

	err = cast.DeleteReplica(id)
	if err != nil {
		switch e := err.(type) {
		case zfs.ReplicaNotFoundError:
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

// ReplicasCastIdIdGet gets a replica from the provided cast.
func (rs ReplicasResource) ReplicasCastIdIdGet(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")
	id := chi.URLParam(r, "id")

	cast, err := rs.GetCast(castId)
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

	replica, err := cast.GetReplica(id)
	if err != nil {
		switch e := err.(type) {
		case zfs.ReplicaNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			l := httplog.LogEntry(r.Context())
			l.Error().Msgf(e.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	result := ReplicaResponse{
		CastId: castId,
		Id:     replica.Id,
		Port:   replica.Port,
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, result)
}

// ReplicasCastIdIdPost creates a replica in the provided cast.
func (rs ReplicasResource) ReplicasCastIdIdPost(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")
	id := chi.URLParam(r, "id")

	cast, err := rs.GetCast(castId)
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

	replica, err := cast.CreateReplica(id)
	if err != nil {
		switch e := err.(type) {
		case zfs.ReplicaAlreadyExistsError:
			w.WriteHeader(http.StatusConflict)
			return
		default:
			l := httplog.LogEntry(r.Context())
			l.Error().Msgf(e.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	result := ReplicaResponse{
		CastId: castId,
		Id:     replica.Id,
		Port:   replica.Port,
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, result)
}

// ReplicasCastIdGet returns a list of the replicas on a provided cast.
func (rs ReplicasResource) ReplicasCastIdGet(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")

	cast, err := rs.GetCast(castId)
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

	replicas := cast.ListReplicas()
	result := make([]ReplicaResponse, 0)
	for _, replica := range replicas {
		item := ReplicaResponse{
			CastId: castId,
			Id:     replica.Id,
			Port:   replica.Port,
		}
		result = append(result, item)
	}
	render.JSON(w, r, result)
}
