package rpcbench

import (
	"testing"
	"time"
)

func TestAggregateEndpointEmpty(t *testing.T) {
	result := aggregateEndpoint("http://localhost:8545", nil)

	if result.URL != "http://localhost:8545" {
		t.Fatalf("URL = %q, want %q", result.URL, "http://localhost:8545")
	}
	if result.Requests != 0 {
		t.Fatalf("Requests = %d, want 0", result.Requests)
	}
	if result.Successes != 0 {
		t.Fatalf("Successes = %d, want 0", result.Successes)
	}
	if result.Errors != 0 {
		t.Fatalf("Errors = %d, want 0", result.Errors)
	}
	if result.Timeouts != 0 {
		t.Fatalf("Timeouts = %d, want 0", result.Timeouts)
	}
	if result.ErrorRate != 0 {
		t.Fatalf("ErrorRate = %v, want 0", result.ErrorRate)
	}
	if result.TimeoutRate != 0 {
		t.Fatalf("TimeoutRate = %v, want 0", result.TimeoutRate)
	}
	if result.HasLatency {
		t.Fatal("HasLatency = true, want false")
	}
}

func TestAggregateEndpointAllSuccess(t *testing.T) {
	observations := []Observation{
		{Outcome: OutcomeSuccess, Latency: 10 * time.Millisecond},
		{Outcome: OutcomeSuccess, Latency: 20 * time.Millisecond},
		{Outcome: OutcomeSuccess, Latency: 30 * time.Millisecond},
		{Outcome: OutcomeSuccess, Latency: 40 * time.Millisecond},
	}

	result := aggregateEndpoint("http://localhost:8545", observations)

	if result.Requests != 4 {
		t.Fatalf("Requests = %d, want 4", result.Requests)
	}
	if result.Successes != 4 {
		t.Fatalf("Successes = %d, want 4", result.Successes)
	}
	if result.Errors != 0 {
		t.Fatalf("Errors = %d, want 0", result.Errors)
	}
	if result.Timeouts != 0 {
		t.Fatalf("Timeouts = %d, want 0", result.Timeouts)
	}
	if result.ErrorRate != 0 {
		t.Fatalf("ErrorRate = %v, want 0", result.ErrorRate)
	}
	if result.TimeoutRate != 0 {
		t.Fatalf("TimeoutRate = %v, want 0", result.TimeoutRate)
	}
	if !result.HasLatency {
		t.Fatal("HasLatency = false, want true")
	}

	assertDuration(t, "P50", result.P50, 20*time.Millisecond)
	assertDuration(t, "P95", result.P95, 40*time.Millisecond)
	assertDuration(t, "P99", result.P99, 40*time.Millisecond)
}

func TestAggregateEndpointMixedOutcomes(t *testing.T) {
	observations := []Observation{
		{Outcome: OutcomeSuccess, Latency: 100 * time.Millisecond},
		{Outcome: OutcomeError},
		{Outcome: OutcomeTimeout},
		{Outcome: OutcomeSuccess, Latency: 50 * time.Millisecond},
		{Outcome: OutcomeError},
	}

	result := aggregateEndpoint("http://localhost:8545", observations)

	if result.Requests != 5 {
		t.Fatalf("Requests = %d, want 5", result.Requests)
	}
	if result.Successes != 2 {
		t.Fatalf("Successes = %d, want 2", result.Successes)
	}
	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2", result.Errors)
	}
	if result.Timeouts != 1 {
		t.Fatalf("Timeouts = %d, want 1", result.Timeouts)
	}

	assertFloat(t, "ErrorRate", result.ErrorRate, 0.4)
	assertFloat(t, "TimeoutRate", result.TimeoutRate, 0.2)

	if !result.HasLatency {
		t.Fatal("HasLatency = false, want true")
	}

	assertDuration(t, "P50", result.P50, 50*time.Millisecond)
	assertDuration(t, "P95", result.P95, 100*time.Millisecond)
	assertDuration(t, "P99", result.P99, 100*time.Millisecond)
}

func TestAggregateEndpointUsesSuccessfulLatenciesOnly(t *testing.T) {
	observations := []Observation{
		{Outcome: OutcomeTimeout, Latency: 5 * time.Second},
		{Outcome: OutcomeError, Latency: 3 * time.Second},
		{Outcome: OutcomeSuccess, Latency: 10 * time.Millisecond},
		{Outcome: OutcomeSuccess, Latency: 20 * time.Millisecond},
	}

	result := aggregateEndpoint("http://localhost:8545", observations)

	if result.Successes != 2 {
		t.Fatalf("Successes = %d, want 2", result.Successes)
	}
	if !result.HasLatency {
		t.Fatal("HasLatency = false, want true")
	}

	assertDuration(t, "P50", result.P50, 10*time.Millisecond)
	assertDuration(t, "P95", result.P95, 20*time.Millisecond)
	assertDuration(t, "P99", result.P99, 20*time.Millisecond)
}

func TestPercentileNearestRank(t *testing.T) {
	values := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}

	tests := []struct {
		name string
		p    int
		want time.Duration
	}{
		{name: "p50", p: 50, want: 50 * time.Millisecond},
		{name: "p95", p: 95, want: 100 * time.Millisecond},
		{name: "p99", p: 99, want: 100 * time.Millisecond},
		{name: "p100", p: 100, want: 100 * time.Millisecond},
		{name: "p1", p: 1, want: 10 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentile(values, tt.p)
			assertDuration(t, tt.name, got, tt.want)
		})
	}
}

func TestPercentileEmpty(t *testing.T) {
	got := percentile(nil, 95)
	assertDuration(t, "percentile", got, 0)
}

func assertDuration(t *testing.T, name string, got, want time.Duration) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %s, want %s", name, got, want)
	}
}

func assertFloat(t *testing.T, name string, got, want float64) {
	t.Helper()

	const tolerance = 0.000001
	if got < want-tolerance || got > want+tolerance {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}
