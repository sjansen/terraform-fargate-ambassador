package main

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type WorkerTracker struct {
	sync.RWMutex

	available map[*Worker]struct{}
}

func NewWorkerTracker() *WorkerTracker {
	return &WorkerTracker{
		available: make(map[*Worker]struct{}),
	}
}

func (t *WorkerTracker) Available() int {
	t.RLock()
	defer t.RUnlock()

	return len(t.available)
}

func (t *WorkerTracker) WorkerBusy(w *Worker) {
	t.Lock()
	delete(t.available, w)
	t.Unlock()
	log.Debug().Msg("Worker busy.")
}

func (t *WorkerTracker) WorkerDone(w *Worker) {
	t.Lock()
	delete(t.available, w)
	t.Unlock()
	log.Debug().Msg("Worker done.")
}

func (t *WorkerTracker) WorkerReady(w *Worker) {
	t.Lock()
	t.available[w] = struct{}{}
	t.Unlock()
	log.Debug().Msg("Worker ready.")
}
