package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestLifecycle(t *testing.T) {
	defer goleak.VerifyNone(t)

	ts := NewTestServer()
	defer ts.Close()

	os.Setenv("AMBASSADOR_CONFIG_URL", ts.URL+"/config")
	go time.AfterFunc(5*time.Second, func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	})

	result := make(chan error)
	timeout := time.After(15 * time.Second)
	go Run(result)
	select {
	case <-timeout:
		t.Fatal("Test timed out without finishing.")
	case err := <-result:
		if err != nil {
			t.Fatalf("Run() failed with error: %e", err)
		}
	}

	t.Logf("/config called %d times", ts.ConfigCalled)
	if ts.ConfigCalled < 1 {
		t.Log("Expected calls to /config missing.")
		t.Fail()
	}
}

type TestServer struct {
	*httptest.Server

	ConfigCalled int
}

func NewTestServer() *TestServer {
	mux := &http.ServeMux{}
	s := &TestServer{
		Server: httptest.NewServer(mux),
	}
	mux.HandleFunc("/config", s.GetConfig)
	return s
}

func (s *TestServer) Close() {
	s.Server.Close()
}

func (s *TestServer) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s.ConfigCalled++
	fmt.Fprintln(w, "hunter2")
}
