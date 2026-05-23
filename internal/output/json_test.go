package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/yermakovsa/rpclat/internal/rpcbench"
)

func TestWriteJSON(t *testing.T) {
	results := []rpcbench.EndpointResult{
		{
			URL:         "https://mainnet.infura.io/v3/API_KEY",
			Requests:    100,
			Successes:   98,
			Errors:      1,
			Timeouts:    1,
			ErrorRate:   0.01,
			TimeoutRate: 0.01,
			P50:         40 * time.Millisecond,
			P95:         80 * time.Millisecond,
			P99:         120 * time.Millisecond,
			HasLatency:  true,
		},
		{
			URL:         "https://rpc.example.com/secret",
			Requests:    10,
			Successes:   0,
			Errors:      8,
			Timeouts:    2,
			ErrorRate:   0.8,
			TimeoutRate: 0.2,
			HasLatency:  false,
		},
	}

	var buf bytes.Buffer

	err := WriteJSON(&buf, RunInfo{
		Method:      "eth_blockNumber",
		Duration:    30 * time.Second,
		Concurrency: 2,
		Timeout:     5 * time.Second,
	}, results, false)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	var got struct {
		Method      string `json:"method"`
		Duration    string `json:"duration"`
		Concurrency int    `json:"concurrency"`
		Timeout     string `json:"timeout"`
		Results     []struct {
			URL         string   `json:"url"`
			Requests    int      `json:"requests"`
			Successes   int      `json:"successes"`
			Errors      int      `json:"errors"`
			Timeouts    int      `json:"timeouts"`
			ErrorRate   float64  `json:"error_rate"`
			TimeoutRate float64  `json:"timeout_rate"`
			P50MS       *float64 `json:"p50_ms"`
			P95MS       *float64 `json:"p95_ms"`
			P99MS       *float64 `json:"p99_ms"`
		} `json:"results"`
	}

	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}

	if got.Method != "eth_blockNumber" {
		t.Fatalf("method = %q, want %q", got.Method, "eth_blockNumber")
	}
	if got.Duration != "30s" {
		t.Fatalf("duration = %q, want %q", got.Duration, "30s")
	}
	if got.Concurrency != 2 {
		t.Fatalf("concurrency = %d, want 2", got.Concurrency)
	}
	if got.Timeout != "5s" {
		t.Fatalf("timeout = %q, want %q", got.Timeout, "5s")
	}
	if len(got.Results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(got.Results))
	}

	first := got.Results[0]
	if first.URL != "https://mainnet.infura.io/..." {
		t.Fatalf("first url = %q, want redacted URL", first.URL)
	}
	if first.Requests != 100 {
		t.Fatalf("first requests = %d, want 100", first.Requests)
	}
	if first.Successes != 98 {
		t.Fatalf("first successes = %d, want 98", first.Successes)
	}
	if first.Errors != 1 {
		t.Fatalf("first errors = %d, want 1", first.Errors)
	}
	if first.Timeouts != 1 {
		t.Fatalf("first timeouts = %d, want 1", first.Timeouts)
	}
	assertFloat(t, "first error_rate", first.ErrorRate, 0.01)
	assertFloat(t, "first timeout_rate", first.TimeoutRate, 0.01)
	assertFloatPtr(t, "first p50_ms", first.P50MS, 40)
	assertFloatPtr(t, "first p95_ms", first.P95MS, 80)
	assertFloatPtr(t, "first p99_ms", first.P99MS, 120)

	second := got.Results[1]
	if second.URL != "https://rpc.example.com/..." {
		t.Fatalf("second url = %q, want redacted URL", second.URL)
	}
	if second.P50MS != nil {
		t.Fatalf("second p50_ms = %v, want nil", *second.P50MS)
	}
	if second.P95MS != nil {
		t.Fatalf("second p95_ms = %v, want nil", *second.P95MS)
	}
	if second.P99MS != nil {
		t.Fatalf("second p99_ms = %v, want nil", *second.P99MS)
	}
}

func TestWriteJSONShowURLs(t *testing.T) {
	rawURL := "https://mainnet.infura.io/v3/API_KEY?token=secret#frag"

	results := []rpcbench.EndpointResult{
		{
			URL:        rawURL,
			Requests:   1,
			Successes:  1,
			P50:        time.Millisecond,
			P95:        time.Millisecond,
			P99:        time.Millisecond,
			HasLatency: true,
		},
	}

	var buf bytes.Buffer

	err := WriteJSON(&buf, RunInfo{
		Method:      "eth_blockNumber",
		Duration:    time.Second,
		Concurrency: 1,
		Timeout:     time.Second,
	}, results, true)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	var got struct {
		Results []struct {
			URL string `json:"url"`
		} `json:"results"`
	}

	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}

	if len(got.Results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(got.Results))
	}
	if got.Results[0].URL != rawURL {
		t.Fatalf("url = %q, want %q", got.Results[0].URL, rawURL)
	}
}

func assertFloat(t *testing.T, name string, got, want float64) {
	t.Helper()

	const tolerance = 0.000001
	if got < want-tolerance || got > want+tolerance {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func assertFloatPtr(t *testing.T, name string, got *float64, want float64) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s = nil, want %v", name, want)
	}

	assertFloat(t, name, *got, want)
}
