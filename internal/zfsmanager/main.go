package zfsmanager

import (
	"sync"

	"github.com/mistifyio/go-zfs"
	"go.uber.org/zap"
)

// ZFSManager contains the state of the ZFS pool and datasets. It manages creation and
// deletion of ZFS datasets and their hierarchy. Also stores some state inside the
// datasets and communicates it upstream.
type ZFSManager struct {
	mu          sync.Mutex
	l           *zap.Logger
	poolName    string
	fsName      string
	castPath    string
	replicaPath string
	fs          *zfs.Dataset
	casts       map[string]*cast
}

// New creates a ZFSManager object and loads the current state structure from the
// underlying ZFS pool
func New(pn string, pp string, pd string, fn string, fp string, cp string, rp string, logger *zap.Logger) *ZFSManager {
	pool := getCreatePool(pn, pd, pp, logger)
	fs := getCreateFilesystem(pool, fn, fp, logger)

	zm := &ZFSManager{
		l:           logger,
		poolName:    pn,
		fsName:      fn,
		castPath:    cp,
		replicaPath: rp,
		fs:          fs,
		casts:       make(map[string]*cast),
	}

	zm.mustLoad()

	logger.Sugar().Debugf("initialized zfsmanager on %s with pool name %s and filesystem name %s", pd, pn, fn)

	return zm
}

// getCreatePool discovers the underlying ZFS pool and creates it if it does not exist
func getCreatePool(name string, dev string, mp string, logger *zap.Logger) *zfs.Zpool {
	pool, err := zfs.GetZpool(name)
	if err != nil {
		logger.Info("pool does not exist, creating", zap.Error(err))
		pool, err = zfs.CreateZpool(name, nil, dev, "-m", mp)
		if err != nil {
			logger.Fatal("failed to create pool", zap.Error(err))
		}
	}

	logger.Sugar().Debugf("found pool %s", pool.Name)

	return pool
}

// getCreateFilesystem discovers the underlying filesystem of the ZFS pool and creates
// it if it does not exist
func getCreateFilesystem(pool *zfs.Zpool, name string, mp string, logger *zap.Logger) *zfs.Dataset {
	fullName := pool.Name + "/" + name

	fs, err := zfs.GetDataset(fullName)
	if err != nil {
		logger.Debug("filesystem does not exist, creating", zap.Error(err))

		properties := map[string]string{
			"mountpoint": mp,
		}
		fs, err = zfs.CreateFilesystem(fullName, properties)
		if err != nil {
			logger.Fatal("failed to create filesystem", zap.Error(err))
		}
	}

	logger.Sugar().Debugf("found filesystem %s", fs.Name)

	return fs
}
