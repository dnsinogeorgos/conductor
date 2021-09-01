package api

import (
	"net/http"

	"github.com/dnsinogeorgos/conductor/internal/conductor"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// ReplicasResource embeds the conductor type to allow the use of its exported methods
type ReplicasResource struct {
	*conductor.Conductor
}

// ReplicaResponse describes the API replica response object
type ReplicaResponse struct {
	Id     string `json:"id"`
	CastId string `json:"castId"`
	Port   int32  `json:"port"`
	Error  string `json:"error,omitempty"`
}

// ReplicasCastIdIdDelete deletes a replica from the provided cast.
func (rr ReplicasResource) ReplicasCastIdIdDelete(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")
	id := chi.URLParam(r, "id")

	err := rr.DeleteReplica(castId, id)
	if err != nil {
		switch e := err.(type) {
		case conductor.CastNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		case conductor.ReplicaNotFoundError:
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

// ReplicasCastIdIdGet gets a replica from the provided cast.
func (rr ReplicasResource) ReplicasCastIdIdGet(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")
	id := chi.URLParam(r, "id")

	replica, err := rr.GetReplica(castId, id)
	if err != nil {
		switch e := err.(type) {
		case conductor.CastNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		case conductor.ReplicaNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			_ = e
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
func (rr ReplicasResource) ReplicasCastIdIdPost(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")
	id := chi.URLParam(r, "id")

	replica, err := rr.CreateReplica(castId, id)
	if err != nil {
		switch e := err.(type) {
		case conductor.CastNotFoundError:
			w.WriteHeader(http.StatusNotFound)
			return
		case conductor.ReplicaAlreadyExistsError:
			w.WriteHeader(http.StatusConflict)
			return
		case conductor.PortsExhaustedError:
			result := ReplicaResponse{
				CastId: castId,
				Id:     id,
				Port:   0,
				Error:  e.Error(),
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			render.JSON(w, r, result)
			return
		default:
			_ = e
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
func (rr ReplicasResource) ReplicasCastIdGet(w http.ResponseWriter, r *http.Request) {
	castId := chi.URLParam(r, "castId")

	replicas, err := rr.ListReplicas(castId)
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
