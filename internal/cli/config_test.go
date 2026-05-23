package cli

import (
	"errors"
	"flag"
	"testing"
	"time"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse([]string{
		"--url", "http://localhost:8545",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(cfg.URLs) != 1 {
		t.Fatalf("len(URLs) = %d, want 1", len(cfg.URLs))
	}
	if cfg.URLs[0] != "http://localhost:8545" {
		t.Fatalf("URLs[0] = %q, want %q", cfg.URLs[0], "http://localhost:8545")
	}
	if cfg.Duration != time.Minute {
		t.Fatalf("Duration = %s, want %s", cfg.Duration, time.Minute)
	}
	if cfg.Concurrency != 5 {
		t.Fatalf("Concurrency = %d, want 5", cfg.Concurrency)
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %s, want %s", cfg.Timeout, 5*time.Second)
	}
	if cfg.Method != DefaultMethod {
		t.Fatalf("Method = %q, want %q", cfg.Method, DefaultMethod)
	}
	if cfg.Output != "table" {
		t.Fatalf("Output = %q, want table", cfg.Output)
	}
	if cfg.ShowURLs {
		t.Fatal("ShowURLs = true, want false")
	}
}

func TestParseMultipleURLs(t *testing.T) {
	cfg, err := Parse([]string{
		"--url", "http://one",
		"--url", "http://two",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(cfg.URLs) != 2 {
		t.Fatalf("len(URLs) = %d, want 2", len(cfg.URLs))
	}
	if cfg.URLs[0] != "http://one" {
		t.Fatalf("URLs[0] = %q, want %q", cfg.URLs[0], "http://one")
	}
	if cfg.URLs[1] != "http://two" {
		t.Fatalf("URLs[1] = %q, want %q", cfg.URLs[1], "http://two")
	}
}

func TestParseCustomValues(t *testing.T) {
	cfg, err := Parse([]string{
		"--url", "http://localhost:8545",
		"--duration", "10s",
		"--concurrency", "2",
		"--timeout", "500ms",
		"--output", "json",
		"--show-urls",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.Duration != 10*time.Second {
		t.Fatalf("Duration = %s, want 10s", cfg.Duration)
	}
	if cfg.Concurrency != 2 {
		t.Fatalf("Concurrency = %d, want 2", cfg.Concurrency)
	}
	if cfg.Timeout != 500*time.Millisecond {
		t.Fatalf("Timeout = %s, want 500ms", cfg.Timeout)
	}
	if cfg.Output != "json" {
		t.Fatalf("Output = %q, want json", cfg.Output)
	}
	if !cfg.ShowURLs {
		t.Fatal("ShowURLs = false, want true")
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "zero duration",
			args: []string{"--url", "http://localhost:8545", "--duration", "0s"},
			want: "--duration must be greater than zero",
		},
		{
			name: "negative duration",
			args: []string{"--url", "http://localhost:8545", "--duration", "-1s"},
			want: "--duration must be greater than zero",
		},
		{
			name: "zero concurrency",
			args: []string{"--url", "http://localhost:8545", "--concurrency", "0"},
			want: "--concurrency must be greater than zero",
		},
		{
			name: "negative concurrency",
			args: []string{"--url", "http://localhost:8545", "--concurrency", "-1"},
			want: "--concurrency must be greater than zero",
		},
		{
			name: "zero timeout",
			args: []string{"--url", "http://localhost:8545", "--timeout", "0s"},
			want: "--timeout must be greater than zero",
		},
		{
			name: "negative timeout",
			args: []string{"--url", "http://localhost:8545", "--timeout", "-1s"},
			want: "--timeout must be greater than zero",
		},
		{
			name: "unsupported output",
			args: []string{"--url", "http://localhost:8545", "--output", "csv"},
			want: `--output must be either table or json, got "csv"`,
		},
		{
			name: "missing url",
			args: nil,
			want: "at least one --url is required",
		},
		{
			name: "empty url",
			args: []string{"--url", ""},
			want: `invalid value "" for flag -url: url must not be empty`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.args)
			if err == nil {
				t.Fatal("Parse returned nil error, want error")
			}
			if err.Error() != tt.want {
				t.Fatalf("error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestParseHelp(t *testing.T) {
	_, err := Parse([]string{"--help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("error = %v, want flag.ErrHelp", err)
	}
}
