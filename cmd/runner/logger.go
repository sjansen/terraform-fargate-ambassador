package main

import (
	stdlog "log"
	"net"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func configureLogger(cfg *Config) {
	log.Warn().
		Int("log-level", cfg.LogLevel).
		Msg("Configuring Logging")
	switch cfg.LogLevel {
	case 0:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 4:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	stdlog.SetOutput(log.Logger)
}

func requestLogger(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-type")
		m := httpsnoop.CaptureMetrics(h, w, r)
		remote, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Info().
			Str("method", r.Method).
			Int("code", m.Code).
			Str("uri", r.URL.String()).
			Str("ct", contentType).
			Int64("len", r.ContentLength).
			Dur("time", m.Duration/time.Millisecond).
			Str("referer", r.Header.Get("Referer")).
			Str("remote", remote).
			Str("ua", r.Header.Get("User-Agent")).
			Send()
	}
	return http.HandlerFunc(fn)
}
