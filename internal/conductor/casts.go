package conductor

import (
	"time"

	"go.uber.org/zap"
)

// Cast contains the state of a cast and it's child relationships
type Cast struct {
	Id        string
	Timestamp string
	replicas  map[string]*Replica
}

// GetCast retrieves the cast object from the state
func (cnd *Conductor) GetCast(id string) (*Cast, error) {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	if _, ok := cnd.casts[id]; !ok {
		cnd.l.Debug("cannot get cast, not found", zap.String("cast", id))
		return &Cast{}, CastNotFoundError{id}
	}

	cnd.l.Debug("getting cast object", zap.String("cast", id))
	return cnd.casts[id], nil
}

// ListCasts returns a slice of the existing casts
func (cnd *Conductor) ListCasts() []*Cast {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	cnd.l.Debug("listing cast objects")
	casts := make([]*Cast, 0)
	for _, cast := range cnd.casts {
		casts = append(casts, cast)
	}

	return casts
}

// TODO: Casts must bind ports and create units just like replicas

// CreateCast orchestrates the creation of a cast using the underlying managers
func (cnd *Conductor) CreateCast(id string) (*Cast, error) {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	if _, ok := cnd.casts[id]; ok {
		cnd.l.Debug("cannot create cast, already exists", zap.String("cast", id))
		return &Cast{}, CastAlreadyExistsError{id}
	}

	cnd.l.Debug("creating cast dataset", zap.String("cast", id))
	timestamp, err := cnd.zm.CreateCastDataset(id, cnd.um.StopMainUnit, cnd.um.StartMainUnit)
	if err != nil {
		return &Cast{}, err
	}

	cnd.l.Info("creating cast object", zap.String("cast", id))
	cast := &Cast{
		Id:        id,
		Timestamp: timestamp.Format(time.RFC3339),
		replicas:  make(map[string]*Replica),
	}
	cnd.casts[id] = cast

	return cast, nil
}

// DeleteCast orchestrates the deletion of a cast using the underlying managers
func (cnd *Conductor) DeleteCast(id string) error {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	if _, ok := cnd.casts[id]; !ok {
		cnd.l.Debug("cannot delete cast, not found", zap.String("cast", id))
		return CastNotFoundError{id}
	}

	if len(cnd.casts[id].replicas) != 0 {
		cnd.l.Debug("cannot delete cast, not empty", zap.String("cast", id))
		return CastNotEmpty{id}
	}

	cnd.l.Debug("deleting cast dataset", zap.String("cast", id))
	err := cnd.zm.DeleteCastDataset(id)
	if err != nil {
		return err
	}

	cnd.l.Info("deleting cast object", zap.String("cast", id))
	delete(cnd.casts, id)

	return nil
}
