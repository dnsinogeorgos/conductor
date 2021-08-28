package zfsmanager

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dnsinogeorgos/conductor/internal/portmanager"
	"github.com/mistifyio/go-zfs"
	"go.uber.org/zap"
)

const castStateFile = ".cast"

// Cast embeds a *zfs.Dataset and replica relationship data.
type Cast struct {
	sync.RWMutex
	*zfs.Dataset
	l           *zap.Logger
	Id          string
	Date        string
	ReplicaPath string
	replicas    map[string]*Replica
	PortManager *portmanager.PortManager
}

// CreateCast creates a cast on the provided filesystem.
func (cnd *Conductor) CreateCast(name string) (*Cast, error) {
	cnd.Lock()
	defer cnd.Unlock()

	castName := cnd.poolName + "/" + cnd.fsName + "/" + name
	mountPoint := cnd.castPath + "/" + name
	p := map[string]string{
		"mountpoint": mountPoint,
	}

	if cast, ok := cnd.casts[castName]; ok {
		return &Cast{}, CastAlreadyExistsError{s: fmt.Sprintf("cast %s already exists in filesystem %s", cast.Name, cnd.fs.Name)}
	}

	cnd.l.Sugar().Debugf("stopping main unit %s for snapshotting", cnd.um.MainServiceName)
	err := cnd.um.StopMain()
	if err != nil {
		return &Cast{}, StopMainError{s: fmt.Sprintf("could not stop main unit %s", cnd.um.MainServiceName)}
	}

	cnd.l.Sugar().Debugf("taking snapshot for cast %s in filesystem %s", castName, cnd.fsName)
	ss, err := cnd.fs.Snapshot(name, false)
	if err != nil {
		return &Cast{}, err
	}

	time.Sleep(5 * time.Second)
	cnd.l.Sugar().Debugf("starting main unit %s after snapshotting", cnd.um.MainServiceName)
	err = cnd.um.StartMain()
	if err != nil {
		return &Cast{}, StopMainError{s: fmt.Sprintf("could not stop main unit %s", cnd.um.MainServiceName)}
	}

	cnd.l.Sugar().Debugf("cloning snapshot for cast %s in filesystem %s", castName, cnd.fsName)
	dsName := cnd.fs.Name + "/" + name
	dataset, err := ss.Clone(dsName, p)
	if err != nil {
		return &Cast{}, err
	}

	replicas := make(map[string]*Replica)
	timestamp := time.Now().UTC()
	cast := &Cast{
		l:           cnd.l,
		Dataset:     dataset,
		Id:          name,
		Date:        timestamp.Format(time.RFC3339),
		ReplicaPath: cnd.replicaPath,
		replicas:    replicas,
		PortManager: cnd.pm,
	}

	err = cast.SaveState(timestamp)
	if err != nil {
		return &Cast{}, err
	}

	cnd.casts[castName] = cast

	return cast, nil
}

// DeleteCast deletes a cast from the provided filesystem.
func (cnd *Conductor) DeleteCast(name string) error {
	cnd.Lock()
	defer cnd.Unlock()

	castName := cnd.poolName + "/" + cnd.fsName + "/" + name

	if cast, ok := cnd.casts[castName]; ok {
		mountPoint := cast.Mountpoint

		// Abort if replicas exist in the cast
		if len(cast.replicas) != 0 {
			return CastContainsReplicasError{s: fmt.Sprintf("cannot delete cast %s, it contains %d replica(s)", cast.Name, len(cast.replicas))}
		}

		// Get parent snapshot
		origin, err := zfs.GetDataset(cast.Origin)
		if err != nil {
			return err
		}

		// Destroy replica
		cnd.l.Sugar().Debugf("deleting cast %s from filesystem %s", cast.Name, cnd.fsName)
		err = cast.Destroy(0)
		if err != nil {
			return err
		}

		// Destroy parent snapshot
		cnd.l.Sugar().Debugf("deleting origin snapshot %s for cast %s", origin.Name, cast.Name)
		err = origin.Destroy(0)
		if err != nil {
			return err
		}

		// Remove cast from map of current casts
		delete(cnd.casts, castName)

		// Cleanup mount directory
		err = os.Remove(mountPoint)
		if err != nil {
			return err
		}
		err = os.Remove(cnd.replicaPath + "/" + name)
		if err != nil {
			_, ok = err.(*os.PathError)
			if !ok {
				return err
			}
		}
	} else {
		return CastNotFoundError{s: fmt.Sprintf("could not find cast %s in filesystem %s", name, cnd.fsName)}
	}
	return nil
}

// GetCast gets a cast from the provided filesystem.
func (cnd *Conductor) GetCast(name string) (*Cast, error) {
	cnd.RLock()
	defer cnd.RUnlock()

	castName := cnd.poolName + "/" + cnd.fsName + "/" + name

	if cast, ok := cnd.casts[castName]; ok {
		return cast, nil
	} else {
		return &Cast{}, CastNotFoundError{s: fmt.Sprintf("could not find cast %s in filesystem %s", castName, cnd.fsName)}
	}
}

// ListCasts returns a slice of the casts on the filesystem.
func (cnd *Conductor) ListCasts() []*Cast {
	cnd.RLock()
	defer cnd.RUnlock()

	casts := make([]*Cast, 0)
	for _, cast := range cnd.casts {
		casts = append(casts, cast)
	}

	return casts
}
