package zfsmanager

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"go.uber.org/zap"
)

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

// loadReplicas discovers the underlying replicas of a cast and populates their
// current state
func (zm *ZFSManager) loadReplicas(castId string) error {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	castName := zm.castFullName(castId)

	if _, ok := zm.casts[castName]; !ok {
		zm.l.Fatal("cannot load cast, not found", zap.String("cast", castId))
		return CastNotFoundError{castName}
	}
	cast := zm.casts[castName]

	zm.l.Debug("reading replica datasets", zap.String("cast", castId))
	children, err := cast.ds.Children(1)
	if err != nil {
		zm.l.Fatal("failed to read replica datasets", zap.Error(err))
		return err
	}

	zm.l.Debug("iterating replica datasets", zap.String("cast", castId))
	for _, replicaDataset := range children {
		if replicaDataset.Type == "filesystem" {
			zm.l.Debug("loading replica", zap.String("replica", replicaDataset.Name), zap.String("cast", cast.id))

			replica := &replica{
				ds:     replicaDataset,
				parent: cast,
			}
			err := zm.loadReplicaState(replica)
			if err != nil {
				return err
			}

			cast.replicas[replicaDataset.Name] = replica
		}
	}

	zm.l.Info("loaded replicas", zap.Int("replicas", len(cast.replicas)), zap.String("cast", cast.id))
	return nil
}

// loadCastState loads the state stored inside the replica dataset
func (zm *ZFSManager) loadReplicaState(replica *replica) error {
	zm.l.Debug("reading replica state", zap.String("path", zm.replicaPath))
	ss := strings.Split(replica.ds.Name, "/")
	s := ss[len(ss)-1] + "/" + replicaStateFile
	f, e := ioutil.ReadFile(zm.replicaPath + "/" + replica.parent.id + "/" + s)
	if e != nil {
		zm.l.Error("failed to read replica state file", zap.String("path", zm.replicaPath+"/"+s))
		return e
	}

	zm.l.Debug("unmarshaling replica state json", zap.String("path", zm.replicaPath+"/"+s))
	freplica := &ReplicaState{}
	e = json.Unmarshal(f, freplica)
	if e != nil {
		zm.l.Error("failed to unmarshal replica state json", zap.String("path", zm.replicaPath+"/"+s))
		return e
	}

	zm.l.Debug("loading replica state", zap.String("replica", freplica.Id))
	replica.id = freplica.Id
	replica.port = freplica.Port

	return nil
}

// mustLoad executes the load methods recursively and exits if an error occurs
func (zm *ZFSManager) mustLoad() {
	err := zm.loadCasts()
	if err != nil {
		zm.l.Fatal("failed to load cast datasets", zap.Error(err))
		return
	}

	castIds := zm.GetCastIds()
	for _, id := range castIds {
		err = zm.loadReplicas(id)
		if err != nil {
			zm.l.Fatal("failed to load replica datasets", zap.Error(err))
			return
		}
	}
}
