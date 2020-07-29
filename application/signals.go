package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
)

func HandleSignals(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		return
	case s := <-sigs:
		log.Info().
			Str("signal", s.String()).
			Msg("Shutdown signal received.")
		cancel()
	}
}
