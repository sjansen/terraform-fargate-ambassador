package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/sjansen/terraform-fargate-ambassador/internal/messages"
)

type Worker struct {
	Config  *Config
	Queue   <-chan messages.Job
	Tracker *WorkerTracker
}

func (w *Worker) Handle(job *messages.Job) {
	msg, err := json.Marshal(&messages.Result{
		Message:  job.MediaURL,
		Metadata: job.Metadata,
	})
	if err == nil {
		_, err = http.Post(job.CallbackURL, "application/json", bytes.NewBuffer(msg))
	}
	if err != nil {
		log.Err(err).Send()
	}
}

func (w *Worker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	t := w.Tracker
	for {
		t.WorkerReady(w)
		select {
		case <-ctx.Done():
			t.WorkerDone(w)
			return
		case job := <-w.Queue:
			t.WorkerBusy(w)
			w.Handle(&job)
		}
	}
}

func rot13(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'm':
			b.WriteRune(r + 13)
		case r >= 'A' && r <= 'M':
			b.WriteRune(r + 13)
		case r >= 'n' && r <= 'z':
			b.WriteRune(r - 13)
		case r >= 'N' && r <= 'Z':
			b.WriteRune(r - 13)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
