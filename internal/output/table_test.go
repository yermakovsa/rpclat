package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yermakovsa/rpclat/internal/rpcbench"
)

func TestWriteTable(t *testing.T) {
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

	err := WriteTable(&buf, results, false)
	if err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	out := buf.String()

	assertContains(t, out, "URL")
	assertContains(t, out, "REQ")
	assertContains(t, out, "OK")
	assertContains(t, out, "ERR")
	assertContains(t, out, "TIMEOUT")
	assertContains(t, out, "ERR%")
	assertContains(t, out, "TIMEOUT%")
	assertContains(t, out, "P50")
	assertContains(t, out, "P95")
	assertContains(t, out, "P99")

	assertContains(t, out, "https://mainnet.infura.io/...")
	assertNotContains(t, out, "API_KEY")

	assertContains(t, out, "https://rpc.example.com/...")
	assertNotContains(t, out, "/secret")

	assertContains(t, out, "100")
	assertContains(t, out, "98")
	assertContains(t, out, "1.0%")
	assertContains(t, out, "80.0%")
	assertContains(t, out, "20.0%")

	assertContains(t, out, "40ms")
	assertContains(t, out, "80ms")
	assertContains(t, out, "120ms")

	assertFieldsContain(t, out, []string{
		"https://rpc.example.com/...",
		"10",
		"0",
		"8",
		"2",
		"80.0%",
		"20.0%",
		"-",
		"-",
		"-",
	})
}

func TestWriteTableShowURLs(t *testing.T) {
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

	err := WriteTable(&buf, results, true)
	if err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	out := buf.String()

	assertContains(t, out, rawURL)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Fatalf("output does not contain %q:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()

	if strings.Contains(s, substr) {
		t.Fatalf("output contains %q, but should not:\n%s", substr, s)
	}
}

func assertFieldsContain(t *testing.T, s string, fields []string) {
	t.Helper()

	lines := strings.Split(s, "\n")
	for _, line := range lines {
		got := strings.Fields(line)
		if len(got) != len(fields) {
			continue
		}

		matches := true
		for i := range fields {
			if got[i] != fields[i] {
				matches = false
				break
			}
		}
		if matches {
			return
		}
	}

	t.Fatalf("output does not contain fields %q:\n%s", fields, s)
}
