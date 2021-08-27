package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/dnsinogeorgos/conductor/internal/api"
	"github.com/dnsinogeorgos/conductor/internal/config"
	"github.com/dnsinogeorgos/conductor/internal/zfs"
)

func main() {
	configfile := flag.String("c", "conductor.json", "path to configuration file")
	flag.Parse()

	c, e := config.NewConfig(*configfile)
	if e != nil {
		log.Fatal(e)
	}

	fs := zfs.NewZFS(
		c.PoolName,
		c.PoolDev,
		c.PoolPath,
		c.FsName,
		c.FsPath,
		c.CastPath,
		c.ReplicaPath,
		c.SystemdUnitName,
		c.PortLowerBound,
		c.PortUpperBound,
	)
	fs.MustLoadAll()

	log.Printf("Server started")

	router := api.NewRouter(c.ServiceName, fs)

	log.Fatal(http.ListenAndServe(":8080", router))
}
