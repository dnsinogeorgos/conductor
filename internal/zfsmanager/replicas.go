package zfsmanager

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mistifyio/go-zfs"
	"go.uber.org/zap"
)

const replicaStateFile = ".replica"

// ReplicaState describes the replica state stored on the dataset
type ReplicaState struct {
	Id   string `json:"id"`
	Port int32  `json:"port"`
}

// replica contains the state of a replica and it's parent relationship
type replica struct {
	id     string
	ds     *zfs.Dataset
	parent *cast
	port   int32
}

// replicaFullName returns the full dataset name of the replica
func (zm *ZFSManager) replicaFullName(castId, id string) string {
	return zm.poolName + "/" + zm.fsName + "/" + castId + "/" + id
}

// CreateReplicaDataset orchestrates the creation of a replica dataset onto the underlying
// ZFS filesystem
func (zm *ZFSManager) CreateReplicaDataset(castId, id string, port int32) error {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	castName := zm.castFullName(castId)
	name := zm.replicaFullName(castId, id)

	if _, ok := zm.casts[castName]; !ok {
		zm.l.Error("cannot create replica, cast not found", zap.String("cast", castId), zap.String("replica", id))
		return CastNotFoundError{castId}
	}

	cast := zm.casts[castName]

	if _, ok := cast.replicas[name]; ok {
		zm.l.Error("cannot create replica, already exists", zap.String("cast", castId), zap.String("replica", id))
		return ReplicaAlreadyExistsError{castId, id}
	}

	zm.l.Debug("snapshotting replica", zap.String("cast", castId), zap.String("replica", id))
	snapshot, err := cast.ds.Snapshot(id, false)
	if err != nil {
		zm.l.Fatal("failed to snapshot replica ", zap.String("cast", castId), zap.String("replica", id), zap.Error(err))
		return err
	}

	mountPoint := zm.replicaPath + "/" + castId + "/" + id
	p := map[string]string{
		"mountpoint": mountPoint,
	}

	zm.l.Debug("cloning snapshot for replica", zap.String("cast", castId), zap.String("replica", id))
	dsName := zm.fs.Name + "/" + castId + "/" + id
	ds, err := snapshot.Clone(dsName, p)
	if err != nil {
		zm.l.Fatal("failed to clone snapshot", zap.String("cast", castId), zap.String("replica", id), zap.Error(err))
		return err
	}

	zm.l.Debug("preparing replica", zap.String("cast", castId), zap.String("replica", id))
	replica := &replica{
		ds:   ds,
		id:   id,
		port: port,
	}

	ss := strings.Split(replica.ds.Name, "/")
	s := ss[len(ss)-1] + "/" + replicaStateFile

	zm.l.Debug("marshaling replica state to json", zap.String("cast", castId), zap.String("replica", id))
	b, err := json.MarshalIndent(&ReplicaState{
		Id:   replica.id,
		Port: replica.port,
	}, "", "  ")
	if err != nil {
		zm.l.Error("failed to marshal replica state json", zap.String("path", zm.replicaPath+"/"+s))
		return err
	}

	zm.l.Debug("writing cast state file", zap.String("path", replica.ds.Mountpoint+"/"+replicaStateFile))
	err = ioutil.WriteFile(replica.ds.Mountpoint+"/"+replicaStateFile, b, 0644)
	if err != nil {
		zm.l.Error("failed to write replica state file", zap.String("path", replica.ds.Mountpoint+"/"+replicaStateFile))
		return err
	}

	zm.l.Debug("creating replica", zap.String("cast", castId), zap.String("replica", id))
	cast.replicas[name] = replica

	return nil
}

// DeleteReplicaDataset orchestrates the deletion of a replica dataset from the underlying
// ZFS filesystem
func (zm *ZFSManager) DeleteReplicaDataset(castId, id string) error {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	castName := zm.castFullName(castId)
	name := zm.replicaFullName(castId, id)

	if _, ok := zm.casts[castName]; !ok {
		zm.l.Error("cannot delete replica, cast not found", zap.String("cast", castId), zap.String("replica", id))
		return CastNotFoundError{castId}
	}

	cast := zm.casts[castName]

	if _, ok := cast.replicas[name]; !ok {
		zm.l.Error("cannot delete replica, not found", zap.String("cast", castId), zap.String("replica", id))
		return ReplicaNotFoundError{castId, id}
	}

	replica := cast.replicas[name]

	zm.l.Debug("getting parent snapshot", zap.String("cast", castId), zap.String("replica", id))
	origin, err := zfs.GetDataset(replica.ds.Origin)
	if err != nil {
		zm.l.Fatal("failed to get parent snapshot", zap.String("cast", castId), zap.String("replica", id), zap.Error(err))
		return err
	}

	zm.l.Debug("deleting replica dataset", zap.String("cast", castId), zap.String("replica", id))
	err = replica.ds.Destroy(0)
	if err != nil {
		zm.l.Fatal("failed to delete replica dataset", zap.String("cast", castId), zap.String("replica", id), zap.Error(err))
		return err
	}

	zm.l.Debug("deleting parent snapshot", zap.String("cast", castId), zap.String("replica", id))
	err = origin.Destroy(0)
	if err != nil {
		zm.l.Fatal("failed to delete parent snapshot", zap.String("cast", castId), zap.String("replica", id), zap.Error(err))
		return err
	}

	zm.l.Debug("deleting replica", zap.String("cast", castId), zap.String("replica", id))
	delete(cast.replicas, name)

	mountPoint := replica.ds.Mountpoint
	err = os.Remove(mountPoint)
	if err != nil {
		return err
	}

	return nil
}

// GetReplicaIds returns a slice of the existing replica ids from a cast
func (zm *ZFSManager) GetReplicaIds(castId string) ([]string, error) {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	castName := zm.castFullName(castId)

	if _, ok := zm.casts[castName]; !ok {
		zm.l.Error("cannot get replica ids, cast not found", zap.String("cast", castId))
		return nil, CastNotFoundError{castId}
	}

	cast := zm.casts[castName]
	replicaIds := make([]string, 0)
	for _, replica := range cast.replicas {
		replicaIds = append(replicaIds, replica.id)
	}

	return replicaIds, nil
}

// GetReplicaPort retrieves the port value from a replica state
func (zm *ZFSManager) GetReplicaPort(castId, id string) (int32, error) {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	castName := zm.castFullName(castId)
	name := zm.replicaFullName(castId, id)

	if _, ok := zm.casts[castName]; !ok {
		zm.l.Error("cannot get replica port, cast not found", zap.String("cast", castId), zap.String("replica", id))
		return 0, CastNotFoundError{castId}
	}

	cast := zm.casts[castName]
	if _, ok := cast.replicas[name]; !ok {
		zm.l.Error("cannot get replica port, not found", zap.String("cast", castId), zap.String("replica", id))
		return 0, ReplicaNotFoundError{castId, id}
	}

	replica := cast.replicas[name]

	return replica.port, nil
}
