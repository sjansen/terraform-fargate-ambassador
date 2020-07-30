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
)

type Server struct{}

func (s *Server) Run(cfg *Config) error {
	logger.Info("Startup initiated.")
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

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("ListenAndServe: %v", err)
			cancel()
		}
		wg.Done()
	}()
	wg.Add(1)

	a, err := NewAmbassador(ctx, cfg)
	if err != nil {
		logger.Errorw("Failed to create Ambassador.",
			"error", err,
		)
		return err
	}
	logger.Info("Startup complete.")

OuterLoop:
	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutdown initiated.")
			break OuterLoop
		default:
			msgs, err := a.ReceiveMessages()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else if len(msgs) > 0 {
				logger.Infow("Message(s) received.",
					"count", len(msgs),
				)
				for _, msg := range msgs {
					http.PostForm(cfg.ApplicationURL+"/echo",
						url.Values{"msg": {msg.Body}},
					)
					a.DeleteMessage(msg.Handle)
					time.Sleep(1 * time.Second)
				}
			}
		}
	}

	wg.Wait()
	logger.Info("Shutdown complete.")

	return nil
}

func NewServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/status", statusHandler)
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
		logger.Debugw(r.Method,
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

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status := GetStatus()
	b, err := json.Marshal(status)
	if err != nil {
		logger.Errorw("Encoding server status failed.",
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
