# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go utility library monorepo (`go.lafronz.com/tools`) with reusable packages for GCP-native Go applications. The primary package is a structured logger for Google Cloud Logging (GCL, formerly Stackdriver), with a network helper for extracting real client IPs behind proxies and load balancers.

## Commands

All modules are independent — run commands from within each module's directory.

```bash
# Test a specific module
cd logger/stackdriver     && go test ./...
cd logger/stackdriver/v2  && go test ./...
cd logger/stackdriver/v3  && go test ./...
cd netHelpers/addresses   && go test ./...

# Run a single test by name
go test -run TestFunctionName ./...

# Lint (golangci-lint)
golangci-lint run ./...

# Vet
go vet ./...

# Tidy dependencies after changes
go mod tidy
```

## Architecture

This is a **multi-module monorepo** — each subdirectory containing a `go.mod` is a fully independent module with its own versioning. Modules are not coupled to the root module.

### Module map

| Import path | Directory | Purpose |
|---|---|---|
| `go.lafronz.com/tools/logger/stackdriver` | `logger/stackdriver/` | GCL structured logging v1 |
| `go.lafronz.com/tools/logger/stackdriver/v2` | `logger/stackdriver/v2/` | v2: adds trace/span context |
| `go.lafronz.com/tools/logger/stackdriver/v3` | `logger/stackdriver/v3/` | v3: `slog.Handler` — **preferred for new code** |
| `go.lafronz.com/tools/netHelpers/addresses` | `netHelpers/addresses/` | Real IP extraction from HTTP headers |

### Logger design

All three versions emit newline-delimited JSON to stdout matching the [GCL structured log schema](https://cloud.google.com/logging/docs/structured-logging).

**v1** (`main.go`): Package-level functions with optional labels via `*L` variants.
**v2** (`log.go`): Same API as v1 but each function accepts `trace, spanID string` parameters. Reads `GOOGLE_CLOUD_PROJECT` from env at `init()`. Decimal span IDs are converted to 16-char zero-padded hex.
**v3** (`handler.go`): Implements `slog.Handler`. Create with `stackdriver.New(opts...)`, use with `slog.New(handler)`. Supports `WithTrace`, `WithLabel`, `WithAttrs`, `WithGroup`. Thread-safe.

### v3 is the recommended logger

```go
h := stackdriver.New(
    stackdriver.WithProjectID("my-project"),
    stackdriver.WithLabels(map[string]string{"env": "prod"}),
)
logger := slog.New(h)

// Per-request logger with trace context
rlog := slog.New(h.WithTrace(traceID, spanID))
rlog.Info("handling request", "method", r.Method)
```

### GCL JSON field names

| slog level | GCL `severity` |
|---|---|
| Debug | `DEBUG` |
| Info | `INFO` |
| Warn | `WARNING` |
| Error | `ERROR` |
| Error+4 and above | `CRITICAL` |

Trace fields use the GCL-standard dotted key names: `logging.googleapis.com/trace`, `logging.googleapis.com/spanId`, `logging.googleapis.com/labels`.

### netHelpers/addresses

`GetRealIP(r *http.Request) *string` — checks headers in order: `X-Real-IP` → `X-Forwarded-For` (takes the first IP in a comma-separated list) → `RemoteAddr`. Returns a non-nil pointer always.
