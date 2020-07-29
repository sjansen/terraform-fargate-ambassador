package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cfg *Config

func main() {
	config, err := NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	cfg = config

	if !cfg.Debug {
		log.Logger = log.Level(zerolog.InfoLevel)
	}
	stdlog.SetOutput(log.Logger)

	log.Info().Msg("Startup initiated.")
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go HandleSignals(ctx, cancel, wg)
	wg.Add(1)

	go MonitorDiskUsage(ctx, wg)
	wg.Add(1)

	srv := NewServer()
	go WaitForShutdown(ctx, srv, wg)
	wg.Add(1)

	log.Info().Msg("Startup complete.")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error().Msgf("ListenAndServe: %v", err)
	}

	wg.Wait()
	log.Info().Msg("Shutdown complete.")
}
