package zfs

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/dnsinogeorgos/conductor/internal/portmanager"

	"github.com/mistifyio/go-zfs"
)

const castStateFile = ".cast"

// Cast embeds a *zfs.Dataset and replica relationship data.
type Cast struct {
	sync.RWMutex
	*zfs.Dataset
	Id          string
	Date        string
	ReplicaPath string
	Replicas    map[string]*Replica
	PortManager *portmanager.PortManager
}

// CreateCast creates a cast on the provided filesystem.
func (fs *ZFS) CreateCast(name string) (*Cast, error) {
	fs.Lock()
	defer fs.Unlock()

	castName := fs.PoolName + "/" + fs.FsName + "/" + name
	mountPoint := fs.CastPath + "/" + name
	p := map[string]string{
		"mountpoint": mountPoint,
	}

	if cast, ok := fs.Casts[castName]; ok {
		return &Cast{}, CastAlreadyExistsError{s: fmt.Sprintf("cast %s already exists in filesystem %s\n", cast.Name, fs.Filesystem.Name)}
	}

	log.Printf("creating cast %s in filesystem %s\n", castName, fs.FsName)
	ss, err := fs.Filesystem.Snapshot(name, false)
	if err != nil {
		return &Cast{}, err
	}

	dsName := fs.Filesystem.Name + "/" + name
	dataset, err := ss.Clone(dsName, p)
	if err != nil {
		return &Cast{}, err
	}

	replicas := make(map[string]*Replica)
	timestamp := time.Now().UTC()
	cast := &Cast{
		Dataset:     dataset,
		Id:          name,
		Date:        timestamp.Format(time.RFC3339),
		ReplicaPath: fs.ReplicaPath,
		Replicas:    replicas,
		PortManager: fs.PortManager,
	}

	err = cast.SaveState(timestamp)
	if err != nil {
		return &Cast{}, err
	}

	fs.Casts[castName] = cast

	return cast, nil
}

// DeleteCast deletes a cast from the provided filesystem.
func (fs *ZFS) DeleteCast(name string) error {
	fs.Lock()
	defer fs.Unlock()

	castName := fs.PoolName + "/" + fs.FsName + "/" + name

	if cast, ok := fs.Casts[castName]; ok {
		mountPoint := cast.Mountpoint

		// Abort if replicas exist in the cast
		if len(cast.Replicas) != 0 {
			return CastContainsReplicasError{s: fmt.Sprintf("cannot delete cast %s, it contains %d replica(s)\n", cast.Name, len(cast.Replicas))}
		}

		// Get parent snapshot
		origin, err := zfs.GetDataset(cast.Origin)
		if err != nil {
			return err
		}

		// Destroy replica
		log.Printf("deleting cast %s from filesystem %s\n", cast.Name, fs.FsName)
		err = cast.Destroy(0)
		if err != nil {
			return err
		}

		// Destroy parent snapshot
		log.Printf("deleting origin snapshot %s for cast %s\n", origin.Name, cast.Name)
		err = origin.Destroy(0)
		if err != nil {
			return err
		}

		// Remove cast from map of current casts
		delete(fs.Casts, castName)

		// Cleanup mount directory
		err = os.Remove(mountPoint)
		if err != nil {
			return err
		}
		err = os.Remove(fs.ReplicaPath + "/" + name)
		if err != nil {
			return err
		}
	} else {
		return CastNotFoundError{s: fmt.Sprintf("could not find cast %s in filesystem %s\n", name, fs.FsName)}
	}
	return nil
}

// GetCast gets a cast from the provided filesystem.
func (fs *ZFS) GetCast(name string) (*Cast, error) {
	fs.RLock()
	defer fs.RUnlock()

	castName := fs.PoolName + "/" + fs.FsName + "/" + name

	if cast, ok := fs.Casts[castName]; ok {
		return cast, nil
	} else {
		return &Cast{}, CastNotFoundError{s: fmt.Sprintf("could not find cast %s in filesystem %s\n", castName, fs.FsName)}
	}
}

// ListCasts returns a slice of the casts on the filesystem.
func (fs *ZFS) ListCasts() []*Cast {
	fs.RLock()
	defer fs.RUnlock()

	casts := make([]*Cast, 0)
	for _, cast := range fs.Casts {
		casts = append(casts, cast)
	}

	return casts
}
