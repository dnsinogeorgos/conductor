package conductor

import (
	"go.uber.org/zap"
)

// Replica contains the state of a replica
type Replica struct {
	Id   string
	Port int32
}

// getUniqueReplicaName returns the unique replica name
func (cnd *Conductor) getUniqueReplicaName(castId, id string) string {
	return castId + "_" + id
}

// DeleteReplica orchestrates the deletion of a replica using the underlying managers
func (cnd *Conductor) DeleteReplica(castId, id string) error {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Debug("cannot delete replica, cast not found", zap.String("cast", castId), zap.String("replica", id))
		return CastNotFoundError{castId}
	}

	cast := cnd.casts[castId]
	if _, ok := cast.replicas[id]; !ok {
		cnd.l.Debug("cannot delete replica, replica not found", zap.String("cast", castId), zap.String("replica", id))
		return ReplicaNotFoundError{castId, id}
	}

	urn := cnd.getUniqueReplicaName(castId, id)
	err := cnd.um.StopUnit(urn)
	if err != nil {
		return err
	}

	cnd.l.Debug("deleting replica dataset", zap.String("cast", castId), zap.String("replica", id))
	err = cnd.zm.DeleteReplicaDataset(castId, id)
	if err != nil {
		return err
	}

	replica := cast.replicas[id]
	cnd.l.Debug("releasing port for replica", zap.String("cast", castId), zap.String("replica", id))
	err = cnd.pm.Release(replica.Port)
	if err != nil {
		return err
	}

	cnd.l.Info("deleting replica object", zap.String("cast", castId), zap.String("replica", id))
	delete(cast.replicas, id)

	return nil
}

// GetReplica retrieves the replica object from the state
func (cnd *Conductor) GetReplica(castId, id string) (*Replica, error) {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Debug("cannot get replica, cast not found", zap.String("cast", castId), zap.String("replica", id))
		return &Replica{}, CastNotFoundError{castId}
	}

	cast := cnd.casts[castId]
	if _, ok := cast.replicas[id]; !ok {
		cnd.l.Debug("cannot get replica, replica not found", zap.String("cast", castId), zap.String("replica", id))
		return &Replica{}, ReplicaNotFoundError{castId, id}
	}

	cnd.l.Debug("getting replica object", zap.String("cast", castId), zap.String("replica", id))
	replica := cast.replicas[id]

	return replica, nil
}

// CreateReplica orchestrates the creation of a replica using the underlying managers
func (cnd *Conductor) CreateReplica(castId, id string) (*Replica, error) {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	port, portErr := cnd.pm.GetNextAvailable()

	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Debug("cannot create replica, cast not found", zap.String("cast", castId), zap.String("replica", id))
		return &Replica{}, CastNotFoundError{castId}
	}

	cast := cnd.casts[castId]
	if _, ok := cast.replicas[id]; ok {
		cnd.l.Debug("cannot create replica, already exists", zap.String("cast", castId), zap.String("replica", id))
		return &Replica{}, ReplicaAlreadyExistsError{castId, id}
	}

	urn := cnd.getUniqueReplicaName(castId, id)
	cnd.l.Debug("binding port for replica", zap.String("cast", castId), zap.String("replica", id))
	if portErr != nil {
		cnd.l.Error("configured range of ports is exhausted", zap.Error(portErr))
		return &Replica{}, PortsExhaustedError{s: portErr.Error()}
	}
	err := cnd.pm.Bind(port, urn)
	if err != nil {
		return &Replica{}, err
	}

	cnd.l.Debug("creating replica dataset", zap.String("cast", castId), zap.String("replica", id))
	err = cnd.zm.CreateReplicaDataset(castId, id, port)
	if err != nil {
		return &Replica{}, err
	}

	cnd.l.Info("creating replica object", zap.String("cast", castId), zap.String("replica", id))
	replica := &Replica{
		Id:   id,
		Port: port,
	}
	cast.replicas[id] = replica

	err = cnd.um.StartUnit(urn, cnd.zm.ReplicaMountPoint(castId, id), port)
	if err != nil {
		return &Replica{}, err
	}

	return replica, nil
}

// ListReplicas returns a slice of the existing replicas
func (cnd *Conductor) ListReplicas(castId string) ([]*Replica, error) {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	replicas := make([]*Replica, 0)
	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Debug("cannot list replica objects, cast not found", zap.String("cast", castId))
		return replicas, CastNotFoundError{castId}
	}

	cnd.l.Debug("listing replica objects", zap.String("cast", castId))
	cast := cnd.casts[castId]
	for _, replica := range cast.replicas {
		replicas = append(replicas, replica)
	}

	return replicas, nil
}
