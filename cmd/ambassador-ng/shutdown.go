package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func HandleSignals(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	logger := zap.S()
	defer wg.Done()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	logger.Debug("Signal handler installed.")

	select {
	case <-ctx.Done():
		logger.Debug("Signal handler stopped.")
	case s := <-ch:
		logger.Infow("Shutdown signal received.",
			"signal", s.String(),
		)
		cancel()
		signal.Stop(ch)
	}
}
