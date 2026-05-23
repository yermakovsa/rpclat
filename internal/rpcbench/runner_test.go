package rpcbench

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunnerReturnsResultPerURLInInputOrder(t *testing.T) {
	serverA := newRPCServer(t, http.StatusOK, `{"jsonrpc":"2.0","id":1,"result":"0x1"}`)
	defer serverA.Close()

	serverB := newRPCServer(t, http.StatusOK, `{"jsonrpc":"2.0","id":1,"result":"0x2"}`)
	defer serverB.Close()

	runner := NewRunner("eth_blockNumber")

	results, err := runner.Run(context.Background(), Options{
		URLs:        []string{serverA.URL, serverB.URL},
		Duration:    50 * time.Millisecond,
		Concurrency: 1,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].URL != serverA.URL {
		t.Fatalf("results[0].URL = %q, want %q", results[0].URL, serverA.URL)
	}
	if results[1].URL != serverB.URL {
		t.Fatalf("results[1].URL = %q, want %q", results[1].URL, serverB.URL)
	}
}

func TestRunnerAggregatesSuccessfulResponses(t *testing.T) {
	server := newRPCServer(t, http.StatusOK, `{"jsonrpc":"2.0","id":1,"result":"0x1234"}`)
	defer server.Close()

	runner := NewRunner("eth_blockNumber")

	results, err := runner.Run(context.Background(), Options{
		URLs:        []string{server.URL},
		Duration:    50 * time.Millisecond,
		Concurrency: 1,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	result := results[0]
	if result.Requests == 0 {
		t.Fatal("Requests = 0, want > 0")
	}
	if result.Successes == 0 {
		t.Fatal("Successes = 0, want > 0")
	}
	if result.Errors != 0 {
		t.Fatalf("Errors = %d, want 0", result.Errors)
	}
	if !result.HasLatency {
		t.Fatal("HasLatency = false, want true")
	}
}

func TestRunnerFailingEndpointDoesNotFailWholeRun(t *testing.T) {
	failingServer := newRPCServer(t, http.StatusInternalServerError, `server error`)
	defer failingServer.Close()

	okServer := newRPCServer(t, http.StatusOK, `{"jsonrpc":"2.0","id":1,"result":"0x1234"}`)
	defer okServer.Close()

	runner := NewRunner("eth_blockNumber")

	results, err := runner.Run(context.Background(), Options{
		URLs:        []string{failingServer.URL, okServer.URL},
		Duration:    50 * time.Millisecond,
		Concurrency: 1,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	if results[0].Requests == 0 {
		t.Fatal("failing endpoint Requests = 0, want > 0")
	}
	if results[0].Errors == 0 {
		t.Fatal("failing endpoint Errors = 0, want > 0")
	}
	if results[0].Successes != 0 {
		t.Fatalf("failing endpoint Successes = %d, want 0", results[0].Successes)
	}

	if results[1].Requests == 0 {
		t.Fatal("successful endpoint Requests = 0, want > 0")
	}
	if results[1].Successes == 0 {
		t.Fatal("successful endpoint Successes = 0, want > 0")
	}
}

func TestRunnerCountsTimeoutsSeparately(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1234"}`))
	}))
	defer server.Close()

	runner := NewRunner("eth_blockNumber")

	results, err := runner.Run(context.Background(), Options{
		URLs:        []string{server.URL},
		Duration:    50 * time.Millisecond,
		Concurrency: 1,
		Timeout:     5 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	result := results[0]
	if result.Requests == 0 {
		t.Fatal("Requests = 0, want > 0")
	}
	if result.Timeouts == 0 {
		t.Fatal("Timeouts = 0, want > 0")
	}
	if result.Errors != 0 {
		t.Fatalf("Errors = %d, want 0", result.Errors)
	}
	if result.Successes != 0 {
		t.Fatalf("Successes = %d, want 0", result.Successes)
	}
	if result.HasLatency {
		t.Fatal("HasLatency = true, want false")
	}
}

func TestRunnerUsesConcurrency(t *testing.T) {
	var requests int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requests, 1)

		time.Sleep(time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1234"}`))
	}))
	defer server.Close()

	runner := NewRunner("eth_blockNumber")

	results, err := runner.Run(context.Background(), Options{
		URLs:        []string{server.URL},
		Duration:    75 * time.Millisecond,
		Concurrency: 3,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	result := results[0]
	if result.Requests < 3 {
		t.Fatalf("Requests = %d, want at least 3", result.Requests)
	}
	if atomic.LoadInt64(&requests) < 3 {
		t.Fatalf("server requests = %d, want at least 3", requests)
	}
}

func TestRunnerStopsAfterDuration(t *testing.T) {
	server := newRPCServer(t, http.StatusOK, `{"jsonrpc":"2.0","id":1,"result":"0x1234"}`)
	defer server.Close()

	runner := NewRunner("eth_blockNumber")

	done := make(chan error, 1)

	go func() {
		_, err := runner.Run(context.Background(), Options{
			URLs:        []string{server.URL},
			Duration:    50 * time.Millisecond,
			Concurrency: 1,
			Timeout:     time.Second,
		})
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after duration")
	}
}

func newRPCServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
}
