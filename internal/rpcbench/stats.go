package rpcbench

import (
	"math"
	"slices"
	"time"
)

func aggregateEndpoint(url string, observations []Observation) EndpointResult {
	result := EndpointResult{
		URL:      url,
		Requests: len(observations),
	}

	latencies := make([]time.Duration, 0, len(observations))

	for _, obs := range observations {
		switch obs.Outcome {
		case OutcomeSuccess:
			result.Successes++
			latencies = append(latencies, obs.Latency)
		case OutcomeTimeout:
			result.Timeouts++
		default:
			result.Errors++
		}
	}

	if result.Requests > 0 {
		result.ErrorRate = float64(result.Errors) / float64(result.Requests)
		result.TimeoutRate = float64(result.Timeouts) / float64(result.Requests)
	}

	if len(latencies) > 0 {
		slices.Sort(latencies)

		result.HasLatency = true
		result.P50 = percentile(latencies, 50)
		result.P95 = percentile(latencies, 95)
		result.P99 = percentile(latencies, 99)
	}

	return result
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	index := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
