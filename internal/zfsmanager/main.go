package zfsmanager

import (
	"sync"

	"github.com/mistifyio/go-zfs"
	"go.uber.org/zap"
)

type ZFSManager struct {
	mu       sync.RWMutex
	l        *zap.Logger
	poolName string
	// poolPath    string
	// poolDev     string
	fsName string
	// fsPath      string
	castPath    string
	replicaPath string
	fs          *zfs.Dataset
	casts       map[string]*Cast
}

func New(pn string, pp string, pd string, fn string, fp string, cp string, rp string, logger *zap.Logger) *ZFSManager {
	pool := CreatePool(pn, pd, pp, logger)
	fs := CreateFilesystem(pool, fn, fp, logger)

	zm := &ZFSManager{
		mu:       sync.RWMutex{},
		l:        logger,
		poolName: pn,
		// poolPath:    pp,
		// poolDev:     pd,
		fsName: fn,
		// fsPath:      fp,
		castPath:    cp,
		replicaPath: rp,
		fs:          fs,
		casts:       make(map[string]*Cast),
	}

	logger.Sugar().Debugf("initialized zfsmanager on %s with pool name %s and filesystem name %s", pd, pn, fn)

	return zm
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
