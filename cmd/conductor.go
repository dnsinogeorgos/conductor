package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/dnsinogeorgos/conductor/internal/api"
	"github.com/dnsinogeorgos/conductor/internal/conductor"
	"github.com/dnsinogeorgos/conductor/internal/config"
)

const appName = "conductor"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.NewConfig(appName)
	if err != nil {
		return err
	}

	logger, _ := zap.NewProduction()
	if cfg.Debug == true {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	cnd := conductor.New(cfg, logger)
	logger.Info("server started")

	router := api.NewRouter(cnd)
	addr := cfg.Address + ":" + strconv.Itoa(int(cfg.Port))
	log.Fatal(http.ListenAndServe(addr, router))

	return nil
}
