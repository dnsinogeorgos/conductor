package zfsmanager

import (
	"encoding/json"
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

// GetCastMountPoint returns the mount point path of the cast
func (zm *ZFSManager) GetCastMountPoint(id string) string {
	return zm.castPath + "/" + id
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

	name := zm.getCastFullName(id)

	if _, ok := zm.casts[name]; !ok {
		zm.l.Error("cannot get cast timestamp, not found", zap.String("cast", id))
		return time.Time{}, CastNotFoundError{id}
	}
	cast := zm.casts[name]

	return cast.timestamp, nil
}

// CreateCastDataset orchestrates the creation of a cast dataset onto the underlying
// ZFS filesystem
func (zm *ZFSManager) CreateCastDataset(id string, preHook func() error, postHook func() error) (time.Time, error) {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	name := zm.getCastFullName(id)

	if _, ok := zm.casts[name]; ok {
		zm.l.Error("cannot create cast, already exists", zap.String("cast", id))
		return time.Time{}, CastAlreadyExistsError{id}
	}

	err := preHook()
	if err != nil {
		return time.Time{}, err
	}

	zm.l.Debug("snapshotting cast", zap.String("cast", id))
	timestamp := time.Now().UTC()
	snapshot, err := zm.fs.Snapshot(id, false)
	if err != nil {
		zm.l.Fatal("failed to snapshot cast", zap.String("cast", id), zap.Error(err))
		return time.Time{}, err
	}

	err = postHook()
	if err != nil {
		return time.Time{}, err
	}

	zm.l.Debug("cloning snapshot for cast", zap.String("cast", id))
	mountPoint := zm.castPath + "/" + id
	p := map[string]string{
		"mountpoint": mountPoint,
	}
	dsName := zm.fs.Name + "/" + id
	dataset, err := snapshot.Clone(dsName, p)
	if err != nil {
		zm.l.Fatal("failed to clone snapshot", zap.String("cast", id), zap.Error(err))
		return time.Time{}, err
	}

	zm.l.Debug("preparing cast", zap.String("cast", id))
	replicas := make(map[string]*replica)
	cast := &cast{
		ds:        dataset,
		id:        id,
		replicas:  replicas,
		timestamp: timestamp,
	}

	err = zm.saveCastState(cast)
	if err != nil {
		return time.Time{}, err
	}

	zm.l.Debug("creating cast", zap.String("cast", id))
	zm.casts[name] = cast

	return timestamp, nil
}

// DeleteCastDataset orchestrates the deletion of a cast dataset from the underlying
// ZFS filesystem
func (zm *ZFSManager) DeleteCastDataset(id string) error {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	name := zm.getCastFullName(id)
	if _, ok := zm.casts[name]; !ok {
		zm.l.Error("cannot delete cast, not found", zap.String("cast", id))
		return CastNotFoundError{id}
	}

	cast := zm.casts[name]
	if len(cast.replicas) != 0 {
		zm.l.Error("cannot delete cast, not empty", zap.String("cast", id))
		return CastNotEmpty{id}
	}

	zm.l.Debug("getting parent snapshot of cast", zap.String("cast", id))
	origin, err := zfs.GetDataset(cast.ds.Origin)
	if err != nil {
		zm.l.Fatal("failed to get parent snapshot", zap.String("cast", id), zap.Error(err))
		return err
	}

	zm.l.Debug("deleting cast dataset", zap.String("cast", id))
	err = cast.ds.Destroy(0)
	if err != nil {
		zm.l.Fatal("failed to delete cast dataset", zap.String("cast", id), zap.Error(err))
		return err
	}

	zm.l.Debug("deleting parent snapshot", zap.String("cast", id))
	err = origin.Destroy(0)
	if err != nil {
		zm.l.Fatal("failed to delete parent snapshot", zap.String("cast", id), zap.Error(err))
		return err
	}

	zm.l.Debug("deleting cast", zap.String("cast", id))
	delete(zm.casts, name)

	zm.l.Debug("cleaning up after deletion", zap.String("cast", id))
	mountPoint := cast.ds.Mountpoint
	err = os.Remove(mountPoint)
	if err != nil {
		zm.l.Fatal("failed to delete mountpoint", zap.String("cast", id), zap.Error(err))
		return err
	}
	err = os.Remove(zm.replicaPath + "/" + id)
	if err != nil {
		_, ok := err.(*os.PathError)
		if !ok {
			zm.l.Fatal("failed to delete replica path of cast", zap.String("cast", id), zap.Error(err))
			return err
		}
		zm.l.Debug("did not find replica path. skipping...", zap.String("cast", id))
	}

	return nil
}

// getCastFullName returns the full dataset name of the cast
func (zm *ZFSManager) getCastFullName(id string) string {
	return zm.poolName + "/" + zm.fsName + "/" + id
}

// saveCastState saves the cast state into the cast dataset
func (zm *ZFSManager) saveCastState(cast *cast) error {
	ss := strings.Split(cast.ds.Name, "/")
	s := ss[len(ss)-1] + "/" + castStateFile

	zm.l.Debug("marshaling cast state to json", zap.String("cast", cast.id))
	b, err := json.MarshalIndent(&CastState{
		Id:        cast.id,
		Timestamp: cast.timestamp,
	}, "", "  ")
	if err != nil {
		zm.l.Error("failed to marshal cast state json", zap.String("path", zm.castPath+"/"+s))
		return err
	}

	zm.l.Debug("writing cast state file", zap.String("path", cast.ds.Mountpoint+"/"+castStateFile))
	err = ioutil.WriteFile(cast.ds.Mountpoint+"/"+castStateFile, b, 0644)
	if err != nil {
		zm.l.Error("failed to write cast state file", zap.String("path", cast.ds.Mountpoint+"/"+castStateFile))
		return err
	}

	return nil
}

// loadCasts discovers the underlying casts and populates their current state
func (zm *ZFSManager) loadCasts() error {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	zm.l.Debug("reading cast datasets")
	children, err := zm.fs.Children(1)
	if err != nil {
		zm.l.Fatal("failed to read cast datasets", zap.Error(err))
		return err
	}

	zm.l.Debug("iterating cast datasets")
	for _, castDataset := range children {
		if castDataset.Type == "filesystem" {
			zm.l.Debug("loading cast from filesystem", zap.String("cast", castDataset.Name))

			cast := &cast{
				ds:       castDataset,
				replicas: make(map[string]*replica),
			}
			err := zm.loadCastState(cast)
			if err != nil {
				return err
			}

			zm.casts[castDataset.Name] = cast
		}
	}

	zm.l.Info("loaded casts from filesystem", zap.Int("casts", len(zm.casts)), zap.String("filesystem", zm.fsName))
	return nil
}

// loadCastState loads the state stored inside the cast dataset
func (zm *ZFSManager) loadCastState(cast *cast) error {
	zm.l.Debug("reading cast state", zap.String("path", zm.castPath))
	ss := strings.Split(cast.ds.Name, "/")
	s := ss[len(ss)-1] + "/" + castStateFile
	f, e := ioutil.ReadFile(zm.castPath + "/" + s)
	if e != nil {
		zm.l.Error("failed to read cast state file", zap.String("path", zm.castPath+"/"+s))
		return e
	}

	zm.l.Debug("unmarshaling cast state json", zap.String("path", zm.castPath+"/"+s))
	fcast := &CastState{}
	e = json.Unmarshal(f, fcast)
	if e != nil {
		zm.l.Error("failed to unmarshal cast state json", zap.String("path", zm.castPath+"/"+s))
		return e
	}

	zm.l.Debug("loading cast state", zap.String("cast", fcast.Id))
	cast.id = fcast.Id
	cast.timestamp = fcast.Timestamp

	return nil
}
