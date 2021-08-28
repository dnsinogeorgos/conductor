package conductor

import (
	"fmt"

	"go.uber.org/zap"
)

// MustLoadAll recursively populates the provided *Conductor with the underlying dataset hierarchy or panics.
func (cnd *Conductor) MustLoadAll() {
	err := cnd.LoadCasts()
	if err != nil {
		cnd.l.Fatal("failed to load casts", zap.Error(err))
	}

	numReplicas := 0
	for _, cast := range cnd.casts {
		err = cast.LoadReplicas()
		if err != nil {
			cnd.l.Fatal("failed to load replicas", zap.String("cast", cast.Name), zap.Error(err))
		}
		numReplicas += len(cast.replicas)
	}
	cnd.l.Sugar().Infof("loaded %d replicas from %d casts", numReplicas, len(cnd.casts))
}

// LoadCasts populates the provided filesystem with the underlying casts.
func (cnd *Conductor) LoadCasts() error {
	children, err := cnd.fs.Children(1)
	if err != nil {
		return LoadFilesystemChildrenError{s: fmt.Sprintf("could not get children for filesystem %s of pool %s\n", cnd.fsName, cnd.poolName)}
	}

	for _, cast := range children {
		if cast.Type == "filesystem" {
			cnd.l.Sugar().Debugf("loading cast %s from filesystem %s", cast.Name, cnd.fsName)

			ncast := &Cast{
				l:           cnd.l,
				Dataset:     cast,
				ReplicaPath: cnd.replicaPath,
				replicas:    make(map[string]*Replica),
				PortManager: cnd.pm,
			}
			err = ncast.LoadState(cnd.castPath)
			if err != nil {
				return LoadCastStateError{s: fmt.Sprintf("could not load cast state for cast %s\n", cast.Name)}
			}

			cnd.casts[cast.Name] = ncast
		}
	}

	cnd.l.Sugar().Infof("loaded %d casts from filesystem %s", len(cnd.casts), cnd.fsName)
	return nil
}

// LoadReplicas populates the provided cast with the underlying replicas.
func (cast *Cast) LoadReplicas() error {
	children, err := cast.Children(1)
	if err != nil {
		return fmt.Errorf("could not get children for cast %s.\n", cast.Name)
	}

	for _, replica := range children {
		if replica.Type == "filesystem" {
			cast.l.Sugar().Debugf("loading replica %s from cast %s", replica.Name, cast.Name)

			nreplica := &Replica{
				Dataset: replica,
			}
			err = nreplica.LoadState(cast.ReplicaPath, cast.Id)
			if err != nil {
				return LoadReplicaStateError{s: fmt.Sprintf("could not load replica state for replica %s\n", replica.Name)}
			}

			err = cast.PortManager.Bind(nreplica.Port, nreplica.Id)
			if err != nil {
				return err
			}

			cast.replicas[replica.Name] = nreplica
		}
	}

	cast.l.Sugar().Debugf("loaded %d replicas from cast %s", len(cast.replicas), cast.Name)
	return nil
}
