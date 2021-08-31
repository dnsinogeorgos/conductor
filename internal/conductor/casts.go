package conductor

import (
	"time"
)

// Cast contains the state of a cast and it's child relationships
type Cast struct {
	Id        string
	Timestamp string
	replicas  map[string]*Replica
}

// DeleteCast orchestrates the deletion of a cast using the underlying managers
func (cnd *Conductor) DeleteCast(id string) error {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	if _, ok := cnd.casts[id]; !ok {
		cnd.l.Sugar().Debugf("cannot delete cast '%s', not found", id)
		return CastNotFoundError{id}
	}

	if len(cnd.casts[id].replicas) != 0 {
		cnd.l.Sugar().Debugf("cannot delete cast '%s', not empty", id)
		return CastNotEmpty{id}
	}

	cnd.l.Sugar().Debugf("deleting cast '%s' dataset", id)
	err := cnd.zm.DeleteCastDataset(id)
	if err != nil {
		return err
	}

	cnd.l.Sugar().Debugf("deleting cast '%s'", id)
	delete(cnd.casts, id)

	return nil
}

// GetCast retrieves the cast object from the state
func (cnd *Conductor) GetCast(id string) (*Cast, error) {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	if _, ok := cnd.casts[id]; !ok {
		cnd.l.Sugar().Debugf("cannot get cast '%s', not found", id)
		return &Cast{}, CastNotFoundError{id}
	}

	cnd.l.Sugar().Debugf("getting cast '%s'", id)
	return cnd.casts[id], nil
}

// CreateCast orchestrates the creation of a cast using the underlying managers
func (cnd *Conductor) CreateCast(id string) (*Cast, error) {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	if _, ok := cnd.casts[id]; ok {
		cnd.l.Sugar().Debugf("cannot create cast '%s', already exists", id)
		return &Cast{}, CastAlreadyExistsError{id}
	}

	cnd.l.Sugar().Debugf("creating cast '%s' dataset", id)
	timestamp, err := cnd.zm.CreateCastDataset(id)
	if err != nil {
		return &Cast{}, err
	}

	cnd.l.Sugar().Debugf("creating cast '%s'", id)
	cast := &Cast{
		Id:        id,
		Timestamp: timestamp.Format(time.RFC3339),
		replicas:  make(map[string]*Replica),
	}
	cnd.casts[id] = cast

	return cast, nil
}

// ListCasts returns a slice of the existing casts
func (cnd *Conductor) ListCasts() []*Cast {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	cnd.l.Sugar().Debugf("listing casts")
	casts := make([]*Cast, 0)
	for _, cast := range cnd.casts {
		casts = append(casts, cast)
	}

	return casts
}
