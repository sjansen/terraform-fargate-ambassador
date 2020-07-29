package main

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
)

func NewServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/echo", echoHandler)
	return &http.Server{
		Addr:        "0.0.0.0:8000",
		Handler:     requestLogger(mux),
		IdleTimeout: 1 * time.Minute,
	}
}

func WaitForShutdown(ctx context.Context, srv *http.Server, wg *sync.WaitGroup) {
	select {
	case <-ctx.Done():
		logger.Info("Shutdown initiated.")
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
		logger.Infow("echo",
			"msg", msg,
		)
	}
}

func requestLogger(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-type")
		m := httpsnoop.CaptureMetrics(h, w, r)
		remote, _, _ := net.SplitHostPort(r.RemoteAddr)
		logger.Infow(r.Method,
			"uri", r.URL.String(),
			"code", m.Code,
			"time", m.Duration/time.Millisecond,
			"ct", contentType,
			"referer", r.Header.Get("Referer"),
			"remote", remote,
			"ua", r.Header.Get("User-Agent"),
		)
	}
	return http.HandlerFunc(fn)
}
