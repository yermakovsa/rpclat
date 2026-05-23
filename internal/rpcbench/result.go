package rpcbench

import "time"

type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeError
	OutcomeTimeout
)

type Observation struct {
	Outcome Outcome
	Latency time.Duration
}

type EndpointResult struct {
	URL         string
	Requests    int
	Successes   int
	Errors      int
	Timeouts    int
	ErrorRate   float64
	TimeoutRate float64
	P50         time.Duration
	P95         time.Duration
	P99         time.Duration
	HasLatency  bool
}
