package rpcbench

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRPCClientCallSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1234"}`))
	}))
	defer server.Close()

	client := newRPCClient("eth_blockNumber")

	got := client.call(context.Background(), server.URL)
	if got != OutcomeSuccess {
		t.Fatalf("call outcome = %v, want %v", got, OutcomeSuccess)
	}
}

func TestRPCClientCallHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newRPCClient("eth_blockNumber")

	got := client.call(context.Background(), server.URL)
	if got != OutcomeError {
		t.Fatalf("call outcome = %v, want %v", got, OutcomeError)
	}
}

func TestRPCClientCallInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := newRPCClient("eth_blockNumber")

	got := client.call(context.Background(), server.URL)
	if got != OutcomeError {
		t.Fatalf("call outcome = %v, want %v", got, OutcomeError)
	}
}

func TestRPCClientCallJSONRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": 1,
			"error": {
				"code": -32601,
				"message": "method not found"
			}
		}`))
	}))
	defer server.Close()

	client := newRPCClient("eth_blockNumber")

	got := client.call(context.Background(), server.URL)
	if got != OutcomeError {
		t.Fatalf("call outcome = %v, want %v", got, OutcomeError)
	}
}

func TestRPCClientCallTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1234"}`))
	}))
	defer server.Close()

	client := newRPCClient("eth_blockNumber")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	got := client.call(ctx, server.URL)
	if got != OutcomeTimeout {
		t.Fatalf("call outcome = %v, want %v", got, OutcomeTimeout)
	}
}

func TestRPCClientCallRequestError(t *testing.T) {
	client := newRPCClient("eth_blockNumber")

	got := client.call(context.Background(), "://bad-url")
	if got != OutcomeError {
		t.Fatalf("call outcome = %v, want %v", got, OutcomeError)
	}
}

func TestRPCClientCallRequestShape(t *testing.T) {
	var (
		gotMethod      string
		gotContentType string
		gotRPCMethod   string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")

		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		gotRPCMethod = req.Method

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1234"}`))
	}))
	defer server.Close()

	client := newRPCClient("eth_blockNumber")

	got := client.call(context.Background(), server.URL)
	if got != OutcomeSuccess {
		t.Fatalf("call outcome = %v, want %v", got, OutcomeSuccess)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("HTTP method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", gotContentType, "application/json")
	}
	if gotRPCMethod != "eth_blockNumber" {
		t.Fatalf("RPC method = %q, want %q", gotRPCMethod, "eth_blockNumber")
	}
}

func TestRPCClientCallInvalidJSONRPCResponse(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty object", body: `{}`},
		{name: "wrong jsonrpc version", body: `{"jsonrpc":"1.0","id":1,"result":"0x1234"}`},
		{name: "wrong id", body: `{"jsonrpc":"2.0","id":2,"result":"0x1234"}`},
		{name: "missing result", body: `{"jsonrpc":"2.0","id":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := newRPCClient("eth_blockNumber")

			got := client.call(context.Background(), server.URL)
			if got != OutcomeError {
				t.Fatalf("call outcome = %v, want %v", got, OutcomeError)
			}
		})
	}
}
