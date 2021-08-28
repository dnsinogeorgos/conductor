package zfsmanager

import (
	"fmt"

	"github.com/dnsinogeorgos/conductor/internal/portmanager"
	"github.com/dnsinogeorgos/conductor/internal/unitmanager"
	"go.uber.org/zap"
)

// MustLoadAll recursively populates the provided *Conductor with the underlying dataset hierarchy or panics.
func (zm *ZFSManager) MustLoadAll(um *unitmanager.UnitManager, pm *portmanager.PortManager) {
	err := zm.LoadCasts()
	if err != nil {
		zm.l.Fatal("failed to load casts", zap.Error(err))
	}

	numReplicas := 0
	for _, cast := range zm.casts {
		err = cast.LoadReplicas(um, pm)
		if err != nil {
			zm.l.Fatal("failed to load replicas", zap.String("cast", cast.Name), zap.Error(err))
		}
		numReplicas += len(cast.replicas)
	}
	zm.l.Sugar().Infof("loaded %d replicas from %d casts", numReplicas, len(zm.casts))
}

// LoadCasts populates the provided filesystem with the underlying casts.
func (zm *ZFSManager) LoadCasts() error {
	children, err := zm.fs.Children(1)
	if err != nil {
		return LoadFilesystemChildrenError{s: fmt.Sprintf("could not get children for filesystem %s of pool %s\n", zm.fsName, zm.poolName)}
	}

	for _, cast := range children {
		if cast.Type == "filesystem" {
			zm.l.Sugar().Debugf("loading cast %s from filesystem %s", cast.Name, zm.fsName)

			ncast := &Cast{
				l:           zm.l,
				Dataset:     cast,
				ReplicaPath: zm.replicaPath,
				replicas:    make(map[string]*Replica),
			}
			err = ncast.LoadState(zm.castPath)
			if err != nil {
				return LoadCastStateError{s: fmt.Sprintf("could not load cast state for cast %s\n", cast.Name)}
			}

			zm.casts[cast.Name] = ncast
		}
	}

	zm.l.Sugar().Infof("loaded %d casts from filesystem %s", len(zm.casts), zm.fsName)
	return nil
}

// LoadReplicas populates the provided cast with the underlying replicas.
func (cast *Cast) LoadReplicas(um *unitmanager.UnitManager, pm *portmanager.PortManager) error {
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

			err = pm.Bind(nreplica.Port, nreplica.Id)
			if err != nil {
				return err
			}

			cast.replicas[replica.Name] = nreplica
		}
	}

	cast.l.Sugar().Debugf("loaded %d replicas from cast %s", len(cast.replicas), cast.Name)
	return nil
}
