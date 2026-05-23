package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/yermakovsa/rpclat/internal/cli"
	"github.com/yermakovsa/rpclat/internal/output"
	"github.com/yermakovsa/rpclat/internal/rpcbench"
)

func main() {
	cfg, err := cli.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			cli.WriteUsage(os.Stdout)
			os.Exit(0)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	runner := rpcbench.NewRunner(cfg.Method)

	fmt.Fprintf(
		os.Stderr,
		"Running benchmark for %s with %d workers per endpoint across %d endpoint(s)...\n",
		cfg.Duration,
		cfg.Concurrency,
		len(cfg.URLs),
	)

	results, err := runner.Run(context.Background(), rpcbench.Options{
		URLs:        cfg.URLs,
		Duration:    cfg.Duration,
		Concurrency: cfg.Concurrency,
		Timeout:     cfg.Timeout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch cfg.Output {
	case "json":
		err = output.WriteJSON(os.Stdout, output.RunInfo{
			Method:      cfg.Method,
			Duration:    cfg.Duration,
			Concurrency: cfg.Concurrency,
			Timeout:     cfg.Timeout,
		}, results, cfg.ShowURLs)
	default:
		err = output.WriteTable(os.Stdout, results, cfg.ShowURLs)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
