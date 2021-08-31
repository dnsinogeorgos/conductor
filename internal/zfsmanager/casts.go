package zfsmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/mistifyio/go-zfs"
	"go.uber.org/zap"
)

const castStateFile = ".cast"

// CastState describes the cast state stored on the dataset
type CastState struct {
	Id        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

// cast contains the state of a cast and it's child relationships
type cast struct {
	id        string
	ds        *zfs.Dataset
	replicas  map[string]*replica
	timestamp time.Time
}

// castFullName returns the full dataset name of the cast
func (zm *ZFSManager) castFullName(id string) string {
	return zm.poolName + "/" + zm.fsName + "/" + id
}

// CreateCastDataset orchestrates the creation of a cast dataset onto the underlying
// ZFS filesystem
func (zm *ZFSManager) CreateCastDataset(id string) (time.Time, error) {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	name := zm.castFullName(id)

	if _, ok := zm.casts[name]; ok {
		zm.l.Sugar().Errorf("cannot create cast '%s', already exists", id)
		return time.Time{}, CastAlreadyExistsError{id}
	}

	zm.l.Sugar().Debugf("snapshotting cast '%s'", id)
	timestamp := time.Now().UTC()
	snapshot, err := zm.fs.Snapshot(id, false)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to snapshot cast '%s'", id), zap.Error(err))
		return time.Time{}, err
	}

	zm.l.Sugar().Debugf("cloning snapshot for cast '%s'", id)
	mountPoint := zm.castPath + "/" + id
	p := map[string]string{
		"mountpoint": mountPoint,
	}
	dsName := zm.fs.Name + "/" + id
	dataset, err := snapshot.Clone(dsName, p)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to clone snapshot of cast '%s'", id), zap.Error(err))
		return time.Time{}, err
	}

	zm.l.Sugar().Debugf("preparing cast '%s'", id)
	replicas := make(map[string]*replica)
	cast := &cast{
		ds:        dataset,
		id:        id,
		replicas:  replicas,
		timestamp: timestamp,
	}

	ss := strings.Split(cast.ds.Name, "/")
	s := ss[len(ss)-1] + "/" + castStateFile

	zm.l.Debug("marshaling cast state to json")
	b, err := json.MarshalIndent(&CastState{
		Id:        cast.id,
		Timestamp: cast.timestamp,
	}, "", "  ")
	if err != nil {
		zm.l.Sugar().Errorf("failed to marshal state json to '%s'", zm.castPath+"/"+s)
		return time.Time{}, err
	}

	zm.l.Sugar().Debugf("writing cast state file to %s", cast.ds.Mountpoint+"/"+castStateFile)
	err = ioutil.WriteFile(cast.ds.Mountpoint+"/"+castStateFile, b, 0644)
	if err != nil {
		zm.l.Sugar().Errorf("failed to write cast state file %s", cast.ds.Mountpoint+"/"+castStateFile)
		return time.Time{}, err
	}

	zm.l.Sugar().Debugf("creating cast '%s'", id)
	zm.casts[name] = cast

	return timestamp, nil
}

// DeleteCastDataset orchestrates the deletion of a cast dataset from the underlying
// ZFS filesystem
func (zm *ZFSManager) DeleteCastDataset(id string) error {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	name := zm.castFullName(id)
	if _, ok := zm.casts[name]; !ok {
		zm.l.Sugar().Errorf("cannot delete cast '%s', not found", id)
		return CastNotFoundError{id}
	}

	cast := zm.casts[name]
	if len(cast.replicas) != 0 {
		zm.l.Sugar().Errorf("cannot delete cast '%s', not empty", id)
		return CastNotEmpty{id}
	}

	zm.l.Sugar().Debugf("getting parent snapshot of cast '%s'", id)
	origin, err := zfs.GetDataset(cast.ds.Origin)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to get parent snapshot of cast '%s'", id), zap.Error(err))
		return err
	}

	zm.l.Sugar().Debugf("deleting cast '%s' dataset", id)
	err = cast.ds.Destroy(0)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to delete cast '%s' dataset", id), zap.Error(err))
		return err
	}

	zm.l.Sugar().Debugf("deleting parent snapshot of cast '%s'", id)
	err = origin.Destroy(0)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to delete parent snapshot of cast '%s'", id), zap.Error(err))
		return err
	}

	zm.l.Sugar().Debugf("deleting cast '%s'", id)
	delete(zm.casts, name)

	zm.l.Sugar().Debugf("cleaning up after deletion of cast '%s'", id)
	mountPoint := cast.ds.Mountpoint
	err = os.Remove(mountPoint)
	if err != nil {
		zm.l.Fatal(fmt.Sprintf("failed to delete mountpoint of cast '%s'", id), zap.Error(err))
		return err
	}
	err = os.Remove(zm.replicaPath + "/" + id)
	if err != nil {
		_, ok := err.(*os.PathError)
		if !ok {
			zm.l.Fatal(fmt.Sprintf("failed to delete replica path of cast '%s'", id), zap.Error(err))
			return err
		}
		zm.l.Sugar().Debugf("did not find replica path of cast '%s'. skipping...", id)
	}

	return nil
}

// GetCastIds returns a slice of the existing cast ids
func (zm *ZFSManager) GetCastIds() []string {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	castIds := make([]string, 0)
	for _, cast := range zm.casts {
		castIds = append(castIds, cast.id)
	}

	return castIds
}

// GetCastTimestamp retrieves the timestamp value from a cast state
func (zm *ZFSManager) GetCastTimestamp(id string) (time.Time, error) {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	name := zm.castFullName(id)

	if _, ok := zm.casts[name]; !ok {
		zm.l.Sugar().Errorf("cannot get cast '%s' timestamp, not found", id)
		return time.Time{}, CastNotFoundError{id}
	}
	cast := zm.casts[name]

	return cast.timestamp, nil
}
