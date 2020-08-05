package main

import "github.com/oklog/ulid/v2"

type State struct {
	RunnerInitialized bool
	SigTermReceived   bool
	TimeoutExpired    bool

	// WorkerCount = cap(InProgress)
	InProgress map[ulid.ULID]string
}
