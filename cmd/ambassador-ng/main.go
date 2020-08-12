package main

import (
	"context"
	"sync"
)

func Run(ch chan<- error) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go HandleSignals(ctx, cancel, wg)
	wg.Add(1)

	cfg, err := NewConfig(ctx)
	if err != nil {
		cancel()
		ch <- err
	}

	go cfg.Refresh(ctx, wg)
	wg.Add(1)

	wg.Wait()
	close(ch)
}
