package zfs

import (
	"fmt"
	"log"
	"os"

	"github.com/mistifyio/go-zfs"
)

const replicaStateFile = ".replica"

// Replica embeds a *zfs.Dataset.
type Replica struct {
	*zfs.Dataset
	Id   string
	Port uint16
}

// CreateReplica creates a replica in the provided cast.
func (cast *Cast) CreateReplica(name string) (*Replica, error) {
	cast.Lock()
	defer cast.Unlock()

	replicaName := cast.Name + "/" + name
	mountPoint := cast.ReplicaPath + "/" + cast.Id + "/" + name
	p := map[string]string{
		"mountpoint": mountPoint,
	}

	if replica, ok := cast.Replicas[replicaName]; ok {
		return &Replica{}, ReplicaAlreadyExistsError{s: fmt.Sprintf("replica %s already exists in cast %s\n", replica.Name, cast.Name)}
	}

	log.Printf("creating replica %s in cast %s\n", name, cast.Name)
	ss, err := cast.Snapshot(name, false)
	if err != nil {
		return &Replica{}, err
	}

	ds, err := ss.Clone(replicaName, p)
	if err != nil {
		return &Replica{}, err
	}

	// Get next available port from pool and bind to replica
	port, err := cast.PortManager.GetNextAvailable()
	if err != nil {
		return &Replica{}, err
	}
	err = cast.PortManager.Bind(port, name)
	if err != nil {
		return &Replica{}, err
	}

	replica := &Replica{
		ds,
		name,
		port,
	}

	err = replica.SaveState()
	if err != nil {
		return &Replica{}, err
	}

	cast.Replicas[replicaName] = replica

	return replica, nil
}

// DeleteReplica deletes a replica from the provided cast.
func (cast *Cast) DeleteReplica(name string) error {
	cast.Lock()
	defer cast.Unlock()

	replicaName := cast.Name + "/" + name

	if replica, ok := cast.Replicas[replicaName]; ok {
		mountPoint := replica.Mountpoint

		// Get parent snapshot
		origin, err := zfs.GetDataset(replica.Origin)
		if err != nil {
			return err
		}

		// Destroy deplica
		log.Printf("deleting replica %s from cast %s.\n", replica.Name, cast.Name)
		err = replica.Destroy(0)
		if err != nil {
			return err
		}

		// Destroy parent snapshot
		log.Printf("deleting origin snapshot %s for replica %s.\n", origin.Name, replica.Name)
		err = origin.Destroy(0)
		if err != nil {
			return err
		}

		// Release port back to the pool
		err = cast.PortManager.Release(replica.Port)
		if err != nil {
			return err
		}

		// Remove replica from map of current replicas
		delete(cast.Replicas, replicaName)

		// Cleanup mount directory
		err = os.Remove(mountPoint)
		if err != nil {
			return err
		}
	} else {
		return ReplicaNotFoundError{s: fmt.Sprintf("could not find replica %s in cast %s\n", name, cast.Name)}
	}
	return nil
}

// GetReplica gets a replica from the provided cast.
func (cast *Cast) GetReplica(name string) (*Replica, error) {
	cast.RLock()
	defer cast.RUnlock()

	replicaName := cast.Name + "/" + name

	if replica, ok := cast.Replicas[replicaName]; ok {
		return replica, nil
	} else {
		return &Replica{}, ReplicaNotFoundError{s: fmt.Sprintf("could not find replica %s in filesystem %s\n", replicaName, cast.Name)}
	}
}

// ListReplicas returns a slice of the replicas on a provided cast.
func (cast *Cast) ListReplicas() []*Replica {
	cast.RLock()
	defer cast.RUnlock()

	replicas := make([]*Replica, 0)
	for _, replica := range cast.Replicas {
		replicas = append(replicas, replica)
	}

	return replicas
}
