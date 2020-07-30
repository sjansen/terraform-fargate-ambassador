package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog/log"
)

func NewServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/echo", echoHandler)
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
		http.PostForm(cfg.LinkURL+"/echo",
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
