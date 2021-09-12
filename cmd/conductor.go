package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"go.uber.org/zap"

	"github.com/dnsinogeorgos/conductor/internal/api"
	"github.com/dnsinogeorgos/conductor/internal/conductor"
	"github.com/dnsinogeorgos/conductor/internal/config"
	"github.com/dnsinogeorgos/signal"
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
	cnd.MustLoad()
	logger.Info("server started")

	sigs := make([]*signal.Signal, 0)
	sigConstants := []syscall.Signal{syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM}
	for _, sig := range sigConstants {
		sigs = append(sigs, &signal.Signal{
			Signal:  sig,
			Handler: cnd.Shutdown,
		})
	}
	signal.Handle(sigs)

	router := api.NewRouter(cnd)
	addr := cfg.Address + ":" + strconv.Itoa(int(cfg.Port))
	log.Fatal(http.ListenAndServe(addr, router))

	return nil
}
