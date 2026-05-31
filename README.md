# rpclat

`rpclat` compares Ethereum JSON-RPC endpoints from the machine or container where it runs.

Use it when you have a few RPC URLs and want a quick view of latency, errors, and timeouts from your laptop, VPS, CI runner, Docker container, or server.

For region comparisons, run `rpclat` from each deployment region you are considering and compare the results.

It calls `eth_blockNumber` for a fixed amount of time and reports request counts, errors, timeouts, and latency percentiles.

It is a quick check from one environment, not a load-testing tool.

## Usage

Run without building a binary:

```bash
go run ./cmd/rpclat --url https://rpc.example
```

Compare multiple endpoints:

```bash
go run ./cmd/rpclat \
  --url https://rpc1.example \
  --url https://rpc2.example \
  --duration 30s \
  --concurrency 5 \
  --timeout 5s
```

Print JSON instead of a table:

```bash
go run ./cmd/rpclat --url https://rpc.example --output json
```

Options:

```text
--url URL            RPC endpoint URL. Repeat to compare endpoints.
--duration DURATION  Run duration. Default: 1m
--concurrency N      Workers per endpoint. Default: 5
--timeout DURATION   Timeout per request. Default: 5s
--output FORMAT      table or json. Default: table
--show-urls          Print full URLs instead of redacted URLs.
```

Rows are not sorted; they follow the order of the `--url` flags.

## Output

Default output is a table:

```text
URL                      REQ  OK   ERR  TIMEOUT  ERR%  TIMEOUT%  P50   P95   P99
https://rpc1.example     100  100  0    0        0.0%  0.0%      30ms  45ms  60ms
https://rpc2.example/... 100  96   2    2        2.0%  2.0%      40ms  80ms  120ms
```

`REQ` is the total number of requests recorded. `OK` is successful JSON-RPC responses. `ERR` is non-timeout errors. `TIMEOUT` is requests that exceeded the per-request timeout.

`ERR%` and `TIMEOUT%` are calculated from total requests. Timeouts are counted separately from errors.

`P50`, `P95`, and `P99` are latency percentiles calculated from successful requests only. If an endpoint has no successful requests, latency values are shown as `-`.

Errors include transport errors, non-2xx HTTP responses, invalid JSON responses, invalid JSON-RPC responses, and JSON-RPC error responses.

`rpclat` checks the JSON-RPC response envelope and that `result` is present. It does not validate the block number value itself.

### JSON

Use `--output json` for JSON output. Rates are fractions, and latency values are in milliseconds:

```json
{
  "method": "eth_blockNumber",
  "duration": "30s",
  "concurrency": 5,
  "timeout": "5s",
  "results": [
    {
      "url": "https://rpc.example",
      "requests": 100,
      "successes": 98,
      "errors": 1,
      "timeouts": 1,
      "error_rate": 0.01,
      "timeout_rate": 0.01,
      "p50_ms": 40,
      "p95_ms": 80,
      "p99_ms": 120
    }
  ]
}
```

If an endpoint has no successful requests, JSON percentile fields are `null`.

## URL redaction

RPC URLs often contain API keys or tokens, so URLs are redacted by default in table and JSON output.

For example:

```text
https://rpc.example/key/secret?token=abc#frag
```

is printed as:

```text
https://rpc.example/...
```

Redaction removes user info, query strings, fragments, and non-root path contents.

Use `--show-urls` only when it is safe to expose full endpoint URLs in your terminal, logs, or CI output.

Redaction is not a secret scanner. It only avoids printing common token locations.

## Build

Build the binary:

```bash
go build ./cmd/rpclat
```

Run it on Unix-like systems:

```bash
./rpclat --url https://rpc.example
```

Run it on Windows PowerShell:

```powershell
.\rpclat.exe --url https://rpc.example
```

## Docker

Build a local image and run it:

```bash
docker build -t rpclat .
docker run --rm rpclat --url https://rpc.example
```

Or run the published image:

```bash
docker run --rm ghcr.io/yermakovsa/rpclat:v0.1.0 --url https://rpc.example
```

Inside Docker, `localhost` means the container itself.

## Non-goals

`rpclat` is intentionally narrow. It does not support custom methods, custom payloads, config files, retries, provider-specific logic, or distributed benchmarking.

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build ./...
```

## License

MIT