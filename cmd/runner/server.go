package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sjansen/terraform-fargate-ambassador/internal/messages"
)

type Server struct {
	config  *Config
	queue   chan messages.Job
	server  http.Server
	workers *WorkerTracker
}

func NewServer(cfg *Config) *Server {
	mux := &http.ServeMux{}
	s := &Server{
		config: cfg,
		queue:  make(chan messages.Job, cfg.QueueCapacity),
		server: http.Server{
			Addr:         ":80",
			Handler:      requestLogger(mux),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		workers: NewWorkerTracker(),
	}
	mux.HandleFunc("/availability", s.Availability)
	return s
}

func (s *Server) Availability(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	msg, err := json.Marshal(&messages.AvailabilityStatus{
		AvailableWorkerCount: s.workers.Available(),
	})
	if err != nil {
		log.Err(err).Send()
	} else {
		w.Write(msg)
	}
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) error {
	log.Info().Msg("Startup initiated.")
	for i := 0; i < s.config.WorkerCount; i += 1 {
		worker := &Worker{
			Config:  s.config,
			Queue:   s.queue,
			Tracker: s.workers,
		}
		go worker.Run(ctx, wg)
		wg.Add(1)
		log.Debug().Msg("Worker started.")
	}

	go s.WaitForShutdown(ctx, wg)
	wg.Add(1)

	go s.ReportReadyStatus(ctx, wg)
	wg.Add(1)

	log.Info().Msg("Startup complete.")
	return s.server.ListenAndServe()
}

func (s *Server) ReportReadyStatus(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	msg, err := json.Marshal(&messages.InitializationStatus{
		Success: true,
	})
	if err != nil {
		log.Err(err).
			Msg("Unable to report ready status.")
		return
	}
	for i := 0; i < s.config.ReadyRetries; i++ {
		<-time.After(time.Millisecond * time.Duration(s.config.ReadyDelayMs))
		log.Debug().
			Int("retries", i).
			Str("msg", string(msg)).
			Msg("Reporting ready status.")
		_, err = http.Post(
			s.config.ReadyURL,
			"application/json",
			bytes.NewBuffer(msg),
		)
		if err == nil {
			return
		}
		log.Err(err).
			Int("retries", i).
			Str("url", s.config.ReadyURL).
			Msg("Reporting ready status failed.")
	}
}

func (s *Server) WaitForShutdown(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		s.server.Shutdown(ctx)
	}
}
