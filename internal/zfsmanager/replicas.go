package zfsmanager

import (
	"encoding/json"
	"fmt"
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
		zm.l.Sugar().Errorf("cannot create replica '%s', cast '%s' not found", id, castId)
		return CastNotFoundError{castId}
	}

	cast := zm.casts[castName]

	if _, ok := cast.replicas[name]; ok {
		zm.l.Sugar().Errorf("cannot create replica '%s' in cast '%s', already exists", id, castId)
		return ReplicaAlreadyExistsError{castId, id}
	}

	zm.l.Sugar().Debugf("snapshotting replica %s in cast %s", id, castId)
	snapshot, err := cast.ds.Snapshot(id, false)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to snapshot replica '%s' in cast '%s'", id, castId), zap.Error(err))
		return err
	}

	mountPoint := zm.replicaPath + "/" + castId + "/" + id
	p := map[string]string{
		"mountpoint": mountPoint,
	}

	zm.l.Sugar().Debugf("cloning snapshot for replica '%s' in cast '%s'", id, castId)
	dsName := zm.fs.Name + "/" + castId + "/" + id
	ds, err := snapshot.Clone(dsName, p)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to clone snapshot for replica '%s' in cast '%s'", id, castId), zap.Error(err))
		return err
	}

	zm.l.Sugar().Debugf("preparing replica '%s' in cast '%s'", id, castId)
	replica := &replica{
		ds:   ds,
		id:   id,
		port: port,
	}

	ss := strings.Split(replica.ds.Name, "/")
	s := ss[len(ss)-1] + "/" + replicaStateFile

	zm.l.Debug("marshaling replica state to json")
	b, err := json.MarshalIndent(&ReplicaState{
		Id:   replica.id,
		Port: replica.port,
	}, "", "  ")
	if err != nil {
		zm.l.Sugar().Errorf("failed to marshal state json to '%s'", zm.replicaPath+"/"+s)
		return err
	}

	zm.l.Sugar().Debugf("writing cast state file to %s", replica.ds.Mountpoint+"/"+replicaStateFile)
	err = ioutil.WriteFile(replica.ds.Mountpoint+"/"+replicaStateFile, b, 0644)
	if err != nil {
		zm.l.Sugar().Errorf("failed to write replica state file %s", replica.ds.Mountpoint+"/"+replicaStateFile)
		return err
	}

	zm.l.Sugar().Debugf("creating replica '%s' in cast '%s'", id, castId)
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
		zm.l.Sugar().Errorf("cannot delete replica '%s', cast '%s' not found", id, castId)
		return CastNotFoundError{castId}
	}

	cast := zm.casts[castName]

	if _, ok := cast.replicas[name]; !ok {
		zm.l.Sugar().Errorf("cannot delete replica '%s' in cast '%s', not found", id, castId)
		return ReplicaNotFoundError{castId, id}
	}

	replica := cast.replicas[name]

	zm.l.Sugar().Debugf("getting parent snapshot of replica '%s' in cast '%s'", id, castId)
	origin, err := zfs.GetDataset(replica.ds.Origin)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to get parent snapshot of replica '%s' in cast '%s'", id, castId), zap.Error(err))
		return err
	}

	zm.l.Sugar().Debugf("deleting replica '%s' dataset in cast '%s'", id, castId)
	err = replica.ds.Destroy(0)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to delete replica '%s' dataset in cast '%s'", id, castId), zap.Error(err))
		return err
	}

	zm.l.Sugar().Debugf("deleting parent snapshot of replica '%s' in cast '%s'", id, castId)
	err = origin.Destroy(0)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to delete parent snapshot of replica '%s' in cast '%s'", id, castId), zap.Error(err))
		return err
	}

	zm.l.Sugar().Debugf("deleting replica '%s' in cast '%s'", id, castId)
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
		zm.l.Sugar().Errorf("cannot get replica ids of cast '%s', not found", castId)
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
		zm.l.Sugar().Errorf("cannot get replica '%s' port, cast '%s' not found", id, castId)
		return 0, CastNotFoundError{castId}
	}

	cast := zm.casts[castName]
	if _, ok := cast.replicas[name]; !ok {
		zm.l.Sugar().Errorf("cannot get replica '%s' port, not found", id)
		return 0, ReplicaNotFoundError{castId, id}
	}

	replica := cast.replicas[name]

	return replica.port, nil
}
