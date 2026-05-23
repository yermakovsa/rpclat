package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"
)

const DefaultMethod = "eth_blockNumber"

type Config struct {
	URLs        []string
	Duration    time.Duration
	Concurrency int
	Timeout     time.Duration
	Method      string
	Output      string
	ShowURLs    bool
}

func Parse(args []string) (Config, error) {
	cfg := defaultConfig()

	fs := newFlagSet(&cfg, io.Discard)
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}

	if len(cfg.URLs) == 0 {
		return cfg, errors.New("at least one --url is required")
	}
	if cfg.Duration <= 0 {
		return cfg, errors.New("--duration must be greater than zero")
	}
	if cfg.Concurrency <= 0 {
		return cfg, errors.New("--concurrency must be greater than zero")
	}
	if cfg.Timeout <= 0 {
		return cfg, errors.New("--timeout must be greater than zero")
	}
	if cfg.Output != "table" && cfg.Output != "json" {
		return cfg, fmt.Errorf("--output must be either table or json, got %q", cfg.Output)
	}

	return cfg, nil
}

func WriteUsage(w io.Writer) {
	cfg := defaultConfig()
	fs := newFlagSet(&cfg, w)
	fs.Usage()
}

func defaultConfig() Config {
	return Config{
		Duration:    time.Minute,
		Concurrency: 5,
		Timeout:     5 * time.Second,
		Method:      DefaultMethod,
		Output:      "table",
	}
}

func newFlagSet(cfg *Config, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("rpclat", flag.ContinueOnError)
	fs.SetOutput(output)

	fs.Var((*urlList)(&cfg.URLs), "url", "Ethereum JSON-RPC endpoint URL. Repeat to compare multiple endpoints.")
	fs.DurationVar(&cfg.Duration, "duration", cfg.Duration, "How long to benchmark each endpoint.")
	fs.IntVar(&cfg.Concurrency, "concurrency", cfg.Concurrency, "Number of workers to run per endpoint.")
	fs.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "Timeout for each JSON-RPC request.")
	fs.StringVar(&cfg.Output, "output", cfg.Output, "Output format: table or json.")
	fs.BoolVar(&cfg.ShowURLs, "show-urls", cfg.ShowURLs, "Print full RPC URLs instead of redacting secrets.")

	fs.Usage = func() {
		fmt.Fprintln(output, "Usage:")
		fmt.Fprintln(output, "  rpclat --url URL [--url URL] [flags]")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Checks Ethereum JSON-RPC latency from the machine or container where it runs.")
		fmt.Fprintln(output, "The benchmark method is fixed to eth_blockNumber.")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Flags:")
		fs.PrintDefaults()
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Examples:")
		fmt.Fprintln(output, "  rpclat --url http://localhost:8545")
		fmt.Fprintln(output, "  rpclat --url https://rpc1.example --url https://rpc2.example --duration 1m --concurrency 5 --timeout 5s")
		fmt.Fprintln(output, "  rpclat --url https://rpc.example --output json")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "URLs are redacted in output by default. Use --show-urls only when it is safe to print full endpoint URLs.")
	}

	return fs
}

type urlList []string

func (u *urlList) String() string {
	return fmt.Sprint([]string(*u))
}

func (u *urlList) Set(value string) error {
	if value == "" {
		return errors.New("url must not be empty")
	}

	*u = append(*u, value)
	return nil
}
