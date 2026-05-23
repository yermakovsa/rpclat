package rpcbench

import (
	"context"
	"sync"
	"time"
)

type Options struct {
	URLs        []string
	Duration    time.Duration
	Concurrency int
	Timeout     time.Duration
}

type Runner struct {
	client *rpcClient
}

func NewRunner(method string) *Runner {
	return &Runner{
		client: newRPCClient(method),
	}
}

func (r *Runner) Run(ctx context.Context, opts Options) ([]EndpointResult, error) {
	runCtx, cancel := context.WithTimeout(ctx, opts.Duration)
	defer cancel()

	results := make([]EndpointResult, len(opts.URLs))

	var wg sync.WaitGroup
	for i, url := range opts.URLs {

		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = r.runEndpoint(runCtx, url, opts.Concurrency, opts.Timeout)
		}()
	}

	wg.Wait()
	return results, nil
}

func (r *Runner) runEndpoint(ctx context.Context, url string, concurrency int, timeout time.Duration) EndpointResult {
	var (
		mu           sync.Mutex
		observations []Observation
		wg           sync.WaitGroup
	)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				if ctx.Err() != nil {
					return
				}

				reqCtx, cancel := context.WithTimeout(ctx, timeout)
				start := time.Now()
				outcome := r.client.call(reqCtx, url)
				latency := time.Since(start)
				runDone := ctx.Err() != nil
				cancel()

				if runDone {
					return
				}

				obs := Observation{
					Outcome: outcome,
				}
				if outcome == OutcomeSuccess {
					obs.Latency = latency
				}

				mu.Lock()
				observations = append(observations, obs)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return aggregateEndpoint(url, observations)
}
