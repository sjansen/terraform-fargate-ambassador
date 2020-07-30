package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog/log"
)

type Server struct{}

func (s *Server) Run(cfg *Config) error {
	log.Info().Msg("Startup initiated.")
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go HandleSignals(ctx, cancel, wg)
	wg.Add(1)

	if cfg.Debug {
		for _, kv := range os.Environ() {
			fmt.Println(kv)
		}

		go MonitorDiskUsage(ctx, wg)
		wg.Add(1)
	}

	srv := NewServer()
	go WaitForShutdown(ctx, srv, wg)
	wg.Add(1)
	log.Info().Msg("Startup complete.")

	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Error().Msgf("ListenAndServe: %v", err)
	}

	wg.Wait()
	log.Info().Msg("Shutdown complete.")
	return err
}

func NewServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/status", statusHandler)
	return &http.Server{
		Addr:        "0.0.0.0:8080",
		Handler:     requestLogger(mux),
		IdleTimeout: 1 * time.Minute,
	}
}

func WaitForShutdown(ctx context.Context, srv *http.Server, wg *sync.WaitGroup) {
	select {
	case <-ctx.Done():
		log.Info().Msg("Shutdown initiated.")
		srv.Shutdown(ctx)
	}
	wg.Done()
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method != "POST":
		w.WriteHeader(http.StatusMethodNotAllowed)
	case r.ParseForm() != nil:
		w.WriteHeader(http.StatusBadRequest)
	case r.PostFormValue("msg") == "":
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusOK)
		msg := r.PostFormValue("msg")
		fmt.Println(msg)
		http.PostForm(cfg.AmbassadorURL+"/echo",
			url.Values{"msg": {msg}},
		)
	}
}

func requestLogger(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-type")
		m := httpsnoop.CaptureMetrics(h, w, r)
		remote, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Debug().
			Str("method", r.Method).
			Str("uri", r.URL.String()).
			Int("code", m.Code).
			Dur("time", m.Duration/time.Millisecond).
			Str("ct", contentType).
			Str("referer", r.Header.Get("Referer")).
			Str("remote", remote).
			Str("ua", r.Header.Get("User-Agent")).
			Send()
	}
	return http.HandlerFunc(fn)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status := GetStatus()
	b, err := json.Marshal(status)
	if err != nil {
		log.Error().Msgf("Encoding server status failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
