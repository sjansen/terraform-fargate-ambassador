package main

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog/log"
)

func NewServer() *http.Server {
	mux := &http.ServeMux{}
	return &http.Server{
		Addr:        "0.0.0.0:8080",
		Handler:     logRequestHandler(mux),
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

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(h, w, r)
		remote, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Info().
			Str("method", r.Method).
			Str("uri", r.URL.String()).
			Int("code", m.Code).
			Int64("size", m.Written).
			Dur("time", m.Duration/time.Millisecond).
			Str("referer", r.Header.Get("Referer")).
			Str("remote", remote).
			Str("ua", r.Header.Get("User-Agent")).
			Send()
	}
	return http.HandlerFunc(fn)
}
