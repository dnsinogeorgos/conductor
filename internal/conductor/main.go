package conductor

import (
	"sync"

	"go.uber.org/zap"

	"github.com/dnsinogeorgos/conductor/internal/config"
	"github.com/dnsinogeorgos/conductor/internal/portmanager"
	"github.com/dnsinogeorgos/conductor/internal/unitmanager"
	"github.com/mistifyio/go-zfs"
)

// Conductor stores the state of the filesystem and the underlying hierarchy.
type Conductor struct {
	sync.RWMutex
	poolName    string
	poolPath    string
	poolDev     string
	fsName      string
	fsPath      string
	castPath    string
	replicaPath string
	l           *zap.Logger
	um          *unitmanager.UnitManager
	pm          *portmanager.PortManager
	// zm          *zfsmanager.ZFSManager
	fs    *zfs.Dataset
	casts map[string]*Cast
}

// New creates a new filesystem or panics.
func New(cfg *config.Config, logger *zap.Logger) *Conductor {
	um := unitmanager.New(cfg.MainUnitName, logger)
	pm := portmanager.New(cfg.PortLowerBound, cfg.PortUpperBound, logger)
	// zm := zfsmanager.New(cfg.PoolName, cfg.PoolDev, cfg.PoolPath, cfg.FilesystemName, cfg.FilesystemPath, cfg.CastPath, cfg.ReplicaPath, logger)
	// zm.MustLoadAll(um, pm, logger)

	pool := CreatePool(cfg.PoolName, cfg.PoolDev, cfg.PoolPath, logger)
	fs := CreateFilesystem(pool, cfg.FilesystemName, cfg.FilesystemPath, logger)

	conductor := &Conductor{
		poolName:    cfg.PoolName,
		poolDev:     cfg.PoolDev,
		poolPath:    cfg.PoolPath,
		fsName:      cfg.FilesystemName,
		fsPath:      cfg.FilesystemPath,
		castPath:    cfg.CastPath,
		replicaPath: cfg.ReplicaPath,
		l:           logger,
		um:          um,
		pm:          pm,
		fs:          fs,
		casts:       make(map[string]*Cast),
	}

	conductor.MustLoadAll()

	return conductor
}

// CreatePool creates a pool if it does not exist. Always returns a pool or panics.
func CreatePool(name string, dev string, mp string, logger *zap.Logger) *zfs.Zpool {
	pool, err := zfs.GetZpool(name)
	if err != nil {
		logger.Sugar().Debugf("pool does not exist, attempting to create: %v", err)

		pool, err = zfs.CreateZpool(name, nil, dev, "-m", mp)
		if err != nil {
			logger.Fatal("failed to create pool", zap.Error(err))
		}
	} else {
		logger.Sugar().Debugf("pool found: %s", pool.Name)
	}

	return pool
}

// CreateFilesystem creates a filesystem in the pool if it does not exist. Always returns a dataset or panics.
func CreateFilesystem(pool *zfs.Zpool, name string, mp string, logger *zap.Logger) *zfs.Dataset {
	fullName := pool.Name + "/" + name

	ds, err := zfs.GetDataset(fullName)
	if err != nil {
		logger.Sugar().Debugf("filesystem does not exist, attempting to create: %v", err)

		properties := map[string]string{
			"mountpoint": mp,
		}
		ds, err = zfs.CreateFilesystem(fullName, properties)
		if err != nil {
			logger.Fatal("failed to create zfs dataset", zap.Error(err))
		}
	} else {
		logger.Sugar().Debugf("filesystem found: %s", ds.Name)
	}

	return ds
}
