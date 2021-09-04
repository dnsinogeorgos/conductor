package conductor

import (
	"sync"
	"time"

	"github.com/dnsinogeorgos/conductor/internal/config"
	"github.com/dnsinogeorgos/conductor/internal/portmanager"
	"github.com/dnsinogeorgos/conductor/internal/unitmanager"
	"github.com/dnsinogeorgos/conductor/internal/zfsmanager"
	"go.uber.org/zap"
)

// Conductor contains the managers and the current state structure
type Conductor struct {
	mu    sync.RWMutex
	l     *zap.Logger
	um    *unitmanager.UnitManager
	pm    *portmanager.PortManager
	zm    *zfsmanager.ZFSManager
	casts map[string]*Cast
}

// New creates a Conductor object and populates the current state structure
func New(cfg *config.Config, logger *zap.Logger) *Conductor {
	um := unitmanager.New(
		cfg.MainUnit,
		cfg.ConfigTemplatePath,
		cfg.UnitTemplateString,
		cfg.ConfigPathTemplateString,
		logger,
	)
	pm := portmanager.New(
		cfg.PortLowerBound,
		cfg.PortUpperBound,
		logger,
	)
	zm := zfsmanager.New(
		cfg.PoolName,
		cfg.PoolDev,
		cfg.PoolPath,
		cfg.FilesystemName,
		cfg.FilesystemPath,
		cfg.CastPath,
		cfg.ReplicaPath,
		logger,
	)

	conductor := &Conductor{
		l:     logger,
		um:    um,
		pm:    pm,
		zm:    zm,
		casts: nil,
	}

	conductor.mustLoad()
	logger.Debug("initialized conductor")

	return conductor
}

func (cnd *Conductor) Shutdown() {
	cnd.l.Info("received signal, shutting down")

	cnd.mu.Lock()
	cnd.l.Info("goodbye")
}

// loadCasts discovers the underlying casts and populates their current state
func (cnd *Conductor) loadCasts() (map[string]*Cast, error) {
	casts := make(map[string]*Cast)

	castIds := cnd.zm.GetCastIds()
	for _, castId := range castIds {
		timestamp, err := cnd.zm.GetCastTimestamp(castId)
		if err != nil {
			return nil, err
		}

		casts[castId] = &Cast{
			Id:        castId,
			Timestamp: timestamp.Format(time.RFC3339),
			replicas:  nil,
		}
	}

	return casts, nil
}

// loadReplicas discovers the underlying replicas of a cast and populates their
// current state
func (cnd *Conductor) loadReplicas(castId string) (map[string]*Replica, error) {
	replicas := make(map[string]*Replica)
	replicaIds, err := cnd.zm.GetReplicaIds(castId)
	if err != nil {
		return nil, err
	}
	for _, replicaId := range replicaIds {
		port, _ := cnd.zm.GetReplicaPort(castId, replicaId)
		if err != nil {
			return replicas, err
		}
		urn := cnd.getUniqueReplicaName(castId, replicaId)
		cnd.l.Debug("binding port for replica", zap.String("cast", castId), zap.String("replica", replicaId))
		err = cnd.pm.Bind(port, urn)
		if err != nil {
			return replicas, err
		}

		replicas[replicaId] = &Replica{
			Id:   replicaId,
			Port: port,
		}
	}

	return replicas, nil
}

// mustLoad executes the load methods recursively and exits if an error occurs
func (cnd *Conductor) mustLoad() {
	casts, err := cnd.loadCasts()
	if err != nil {
		cnd.l.Fatal("failed to populate conductor with casts")
		return
	}

	for _, cast := range casts {
		replicas, err := cnd.loadReplicas(cast.Id)
		if err != nil {
			cnd.l.Fatal("failed to populate cast with replicas", zap.String("cast", cast.Id))
			return
		}

		cast.replicas = replicas
	}

	cnd.casts = casts
	return
}
