package output

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/yermakovsa/rpclat/internal/rpcbench"
)

func WriteTable(w io.Writer, results []rpcbench.EndpointResult, showURLs bool) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintln(tw, "URL\tREQ\tOK\tERR\tTIMEOUT\tERR%\tTIMEOUT%\tP50\tP95\tP99"); err != nil {
		return err
	}

	for _, result := range results {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\n",
			displayURL(result.URL, showURLs),
			result.Requests,
			result.Successes,
			result.Errors,
			result.Timeouts,
			formatRate(result.ErrorRate),
			formatRate(result.TimeoutRate),
			formatLatency(result.P50, result.HasLatency),
			formatLatency(result.P95, result.HasLatency),
			formatLatency(result.P99, result.HasLatency),
		); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func formatRate(rate float64) string {
	return fmt.Sprintf("%.1f%%", rate*100)
}

func formatLatency(d time.Duration, ok bool) string {
	if !ok {
		return "-"
	}

	if d < time.Microsecond {
		return d.String()
	}

	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}

	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}

	return d.Round(100 * time.Millisecond).String()
}
