package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/felixge/httpsnoop"
	isatty "github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func main() {
	logger = NewLogger(3)

	mux := &http.ServeMux{}
	srv := http.Server{
		Addr:     "127.0.0.1:8080",
		ErrorLog: zap.NewStdLog(logger.Desugar()),
		Handler:  logRequestHandler(mux),
	}

	idleConnsClosed := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		logger.Infow("Shutdown signal received.",
			"signal", s.String(),
		)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf("Shutdown: %v", err)
		}
		close(idleConnsClosed)
		cancel()
	}()

	logger.Info("Startup complete.")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Errorf("ListenAndServe: %v", err)
	}
	<-idleConnsClosed
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(h, w, r)
		remote, _, _ := net.SplitHostPort(r.RemoteAddr)
		logger.Infow("",
			"method", r.Method,
			"uri", r.URL.String(),
			"code", strconv.Itoa(m.Code),
			"referer", r.Header.Get("Referer"),
			"remote", remote,
			"size", strconv.FormatInt(m.Written, 10),
			"time", strconv.FormatInt(int64(m.Duration/time.Millisecond), 10),
			"ua", r.Header.Get("User-Agent"),
		)
	}
	return http.HandlerFunc(fn)
}

// NewLogger returns a logger
//
// Valid levels are:
//   0 = errors only,
//   1 = include warnings,
//   2 = include informational messages,
//   3 = include debug messages.
func NewLogger(verbosity int) *zap.SugaredLogger {
	var level zapcore.Level
	switch {
	case verbosity >= 3:
		level = zapcore.DebugLevel
	case verbosity == 2:
		level = zapcore.InfoLevel
	case verbosity == 1:
		level = zapcore.WarnLevel
	default:
		level = zapcore.ErrorLevel
	}

	var stdout io.Writer = os.Stdout
	encoder := zapcore.CapitalLevelEncoder
	if x, ok := stdout.(interface{ Fd() uintptr }); ok {
		if isatty.IsTerminal(x.Fd()) {
			encoder = zapcore.CapitalColorLevelEncoder
		}
	}
	cfg := zapcore.EncoderConfig{
		LevelKey:       "level",
		MessageKey:     "msg",
		NameKey:        "logger",
		TimeKey:        "timestamp",
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeLevel:    encoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(stdout),
		level,
	)

	return zap.New(core).Sugar()
}
