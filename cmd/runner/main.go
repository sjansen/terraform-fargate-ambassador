package main

import (
	"context"
	"fmt"
	"os"
	"sync"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	configureLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go HandleSignals(ctx, cancel, wg)
	wg.Add(1)

	srv := NewServer(cfg)
	srv.Run(ctx, wg)
	wg.Wait()
}
