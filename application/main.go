package main

import (
	"context"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog/log"
)

func main() {
	stdlog.SetOutput(log.Logger)

	mux := &http.ServeMux{}
	srv := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: logRequestHandler(mux),
	}

	idleConnsClosed := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		log.Info().
			Str("signal", s.String()).
			Msg("Shutdown signal received.")
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Msgf("Shutdown: %v", err)
		}
		close(idleConnsClosed)
		cancel()
	}()

	log.Info().Msg("Startup complete.")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error().Msgf("ListenAndServe: %v", err)
	}
	<-idleConnsClosed
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
