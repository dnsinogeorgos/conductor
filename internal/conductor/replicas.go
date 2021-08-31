package conductor

// Replica contains the state of a replica
type Replica struct {
	Id   string
	Port int32
}

// DeleteReplica orchestrates the deletion of a replica using the underlying managers
func (cnd *Conductor) DeleteReplica(castId, id string) error {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Sugar().Debugf("cannot delete replica '%s' in cast '%s', cast not found", id, castId)
		return CastNotFoundError{castId}
	}

	cast := cnd.casts[castId]
	if _, ok := cast.replicas[id]; !ok {
		cnd.l.Sugar().Debugf("cannot delete replica '%s' in cast '%s', replica not found", id, castId)
		return ReplicaNotFoundError{castId, id}
	}

	cnd.l.Sugar().Debugf("deleting replica '%s' dataset in cast '%s'", id, castId)
	err := cnd.zm.DeleteReplicaDataset(castId, id)
	if err != nil {
		return err
	}

	cnd.l.Sugar().Debugf("deleting replica '%s' in cast '%s'", id, castId)
	delete(cast.replicas, id)

	return nil
}

// GetReplica retrieves the replica object from the state
func (cnd *Conductor) GetReplica(castId, id string) (*Replica, error) {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Sugar().Debugf("cannot get replica '%s' in cast '%s', cast not found", id, castId)
		return &Replica{}, CastNotFoundError{castId}
	}

	cast := cnd.casts[castId]
	if _, ok := cast.replicas[id]; !ok {
		cnd.l.Sugar().Debugf("cannot get replica '%s' in cast '%s', replica not found", id, castId)
		return &Replica{}, ReplicaNotFoundError{castId, id}
	}

	cnd.l.Sugar().Debugf("getting replica '%s' in cast '%s'", id, castId)
	replica := cast.replicas[id]

	return replica, nil
}

// CreateReplica orchestrates the creation of a replica using the underlying managers
func (cnd *Conductor) CreateReplica(castId, id string) (*Replica, error) {
	cnd.mu.Lock()
	defer cnd.mu.Unlock()

	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Sugar().Debugf("cannot create replica '%s' in cast '%s', cast not found", id, castId)
		return &Replica{}, CastNotFoundError{castId}
	}

	cast := cnd.casts[castId]
	if _, ok := cast.replicas[id]; ok {
		cnd.l.Sugar().Debugf("cannot create replica '%s' in cast '%s', already exists", id, castId)
		return &Replica{}, ReplicaAlreadyExistsError{castId, id}
	}

	cnd.l.Sugar().Debugf("creating replica '%s' dataset in cast '%s'", id, castId)
	err := cnd.zm.CreateReplicaDataset(castId, id, 0)
	if err != nil {
		return &Replica{}, err
	}

	cnd.l.Sugar().Debugf("creating replica '%s' in cast '%s'", id, castId)
	replica := &Replica{
		Id:   id,
		Port: 0,
	}
	cast.replicas[id] = replica

	return replica, nil
}

// ListReplicas returns a slice of the existing replicas
func (cnd *Conductor) ListReplicas(castId string) ([]*Replica, error) {
	cnd.mu.RLock()
	defer cnd.mu.RUnlock()

	replicas := make([]*Replica, 0)
	if _, ok := cnd.casts[castId]; !ok {
		cnd.l.Sugar().Debugf("cannot get replicas in cast '%s', cast not found", castId)
		return replicas, CastNotFoundError{castId}
	}

	cnd.l.Sugar().Debugf("listing replicas in cast '%s'", castId)
	cast := cnd.casts[castId]
	for _, replica := range cast.replicas {
		replicas = append(replicas, replica)
	}

	return replicas, nil
}
