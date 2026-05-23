package output

import (
	"encoding/json"
	"io"
	"time"

	"github.com/yermakovsa/rpclat/internal/rpcbench"
)

type RunInfo struct {
	Method      string
	Duration    time.Duration
	Concurrency int
	Timeout     time.Duration
}

type jsonOutput struct {
	Method      string       `json:"method"`
	Duration    string       `json:"duration"`
	Concurrency int          `json:"concurrency"`
	Timeout     string       `json:"timeout"`
	Results     []jsonResult `json:"results"`
}

type jsonResult struct {
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
}

func WriteJSON(w io.Writer, run RunInfo, results []rpcbench.EndpointResult, showURLs bool) error {
	out := jsonOutput{
		Method:      run.Method,
		Duration:    run.Duration.String(),
		Concurrency: run.Concurrency,
		Timeout:     run.Timeout.String(),
		Results:     make([]jsonResult, 0, len(results)),
	}

	for _, result := range results {
		item := jsonResult{
			URL:         displayURL(result.URL, showURLs),
			Requests:    result.Requests,
			Successes:   result.Successes,
			Errors:      result.Errors,
			Timeouts:    result.Timeouts,
			ErrorRate:   result.ErrorRate,
			TimeoutRate: result.TimeoutRate,
		}

		if result.HasLatency {
			item.P50MS = durationMillis(result.P50)
			item.P95MS = durationMillis(result.P95)
			item.P99MS = durationMillis(result.P99)
		}

		out.Results = append(out.Results, item)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func durationMillis(d time.Duration) *float64 {
	ms := float64(d) / float64(time.Millisecond)
	return &ms
}
