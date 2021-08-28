package zfsmanager

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"
)

type CastState struct {
	Id   string `json:"id"`
	Date string `json:"date"`
}

// LoadState returns the cast state from disk.
func (cast *Cast) LoadState(path string) error {
	cast.RLock()
	defer cast.RUnlock()

	ss := strings.Split(cast.Name, "/")
	s := ss[len(ss)-1] + "/" + castStateFile
	f, e := ioutil.ReadFile(path + "/" + s)
	if e != nil {
		return e
	}

	fcast := &CastState{}
	e = json.Unmarshal(f, fcast)
	if e != nil {
		return e
	}

	cast.Id = fcast.Id
	cast.Date = fcast.Date

	return nil
}

// SaveState stores the cast state to disk.
func (cast *Cast) SaveState(timestamp time.Time) error {
	cast.Lock()
	defer cast.Unlock()

	b, err := json.MarshalIndent(&CastState{
		Id:   cast.Id,
		Date: timestamp.Format(time.RFC3339),
	}, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cast.Mountpoint+"/"+castStateFile, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

type ReplicaState struct {
	Id   string `json:"id"`
	Port int32  `json:"port"`
}

// LoadState returns the replica state from disk.
func (replica *Replica) LoadState(path string, castId string) error {
	ss := strings.Split(replica.Name, "/")
	s := ss[len(ss)-1] + "/" + replicaStateFile
	f, e := ioutil.ReadFile(path + "/" + castId + "/" + s)
	if e != nil {
		return e
	}

	freplica := &ReplicaState{}
	e = json.Unmarshal(f, freplica)
	if e != nil {
		return e
	}

	replica.Id = freplica.Id
	replica.Port = freplica.Port

	return nil
}

// SaveState stores the replica state to disk.
func (replica *Replica) SaveState() error {
	b, err := json.MarshalIndent(&ReplicaState{
		Id:   replica.Id,
		Port: replica.Port,
	}, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(replica.Mountpoint+"/"+replicaStateFile, b, 0644)
	if err != nil {
		return err
	}

	return nil
}
