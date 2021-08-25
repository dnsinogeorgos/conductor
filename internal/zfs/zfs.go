package zfs

import (
	"log"
	"sync"

	"github.com/dnsinogeorgos/conductor/internal/portmanager"

	"github.com/mistifyio/go-zfs"
)

// ZFS stores the state of the filesystem and the underlying hierarchy.
type ZFS struct {
	sync.RWMutex
	PoolName    string
	PoolPath    string
	PoolDev     string
	FsName      string
	FsPath      string
	CastPath    string
	ReplicaPath string
	PortManager *portmanager.PortManager
	Filesystem  *zfs.Dataset
	Casts       map[string]*Cast
}

// NewZFS creates a new filesystem or panics.
func NewZFS(pn string, pd string, pp string, fn string, fp string, cp string, rp string, pstart uint16, pend uint16) *ZFS {
	pm := portmanager.NewPortManager(pstart, pend)

	fs := &ZFS{
		PoolName:    pn,
		PoolDev:     pd,
		PoolPath:    pp,
		FsName:      fn,
		FsPath:      fp,
		CastPath:    cp,
		ReplicaPath: rp,
		PortManager: pm,
	}

	pool := CreatePool(pn, pd, pp)
	fs.Filesystem = CreateFilesystem(pool, fn, fp)

	fs.Casts = make(map[string]*Cast)

	return fs
}

// CreatePool creates a pool if it does not exist. Always returns a pool or panics.
func CreatePool(name string, dev string, mp string) *zfs.Zpool {
	pool, err := zfs.GetZpool(name)
	if err != nil {
		log.Printf("pool does not exist, attempting to create: %v\n", err)

		pool, err = zfs.CreateZpool(name, nil, dev, "-m", mp)
		if err != nil {
			log.Fatalf("failed to create pool: %v\nSeppuku!", err)
		}
	} else {
		log.Printf("pool found: %s\n", pool.Name)
	}

	return pool
}

// CreateFilesystem creates a filesystem in the pool if it does not exist. Always returns a dataset or panics.
func CreateFilesystem(pool *zfs.Zpool, name string, mp string) *zfs.Dataset {
	fullName := pool.Name + "/" + name

	ds, err := zfs.GetDataset(fullName)
	if err != nil {
		log.Printf("filesystem does not exist, attempting to create: %v\n", err)

		properties := map[string]string{
			"mountpoint": mp,
		}
		ds, err = zfs.CreateFilesystem(fullName, properties)
		if err != nil {
			log.Fatalf("failed to create zfs dataset: %v\nSeppuku!", err)
		}
	} else {
		log.Printf("filesystem found: %s\n", ds.Name)
	}

	return ds
}
