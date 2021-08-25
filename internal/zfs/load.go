package zfs

import (
	"fmt"
	"log"
)

// MustLoadAll recursively populates the provided *ZFS with the underlying dataset hierarchy or panics.
func (fs *ZFS) MustLoadAll() {
	err := fs.LoadCasts()
	if err != nil {
		log.Fatalf("%s", err)
	}
	for _, cast := range fs.Casts {
		err = cast.LoadReplicas()
		if err != nil {
			log.Fatalf("%s", err)
		}
	}
}

// LoadCasts populates the provided filesystem with the underlying casts.
func (fs *ZFS) LoadCasts() error {
	children, err := fs.Filesystem.Children(1)
	if err != nil {
		return LoadFilesystemChildrenError{s: fmt.Sprintf("could not get children for filesystem %s of pool %s\n", fs.FsName, fs.PoolName)}
	}

	for _, cast := range children {
		if cast.Type == "filesystem" {
			log.Printf("loading cast %s from filesystem %s.\n", cast.Name, fs.FsName)

			ncast := &Cast{
				Dataset:     cast,
				ReplicaPath: fs.ReplicaPath,
				Replicas:    make(map[string]*Replica),
				PortManager: fs.PortManager,
			}
			err = ncast.LoadState(fs.CastPath)
			if err != nil {
				return LoadCastStateError{s: fmt.Sprintf("could not load cast state for cast %s\n", cast.Name)}
			}

			fs.Casts[cast.Name] = ncast
		}
	}

	log.Printf("loaded %d casts from filesystem %s.\n", len(fs.Casts), fs.FsName)
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
			log.Printf("loading replica %s from cast %s.\n", replica.Name, cast.Name)

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

			cast.Replicas[replica.Name] = nreplica
		}
	}

	log.Printf("loaded %d replicas from cast %s\n", len(cast.Replicas), cast.Name)
	return nil
}
