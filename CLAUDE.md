# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go utility library monorepo (`go.lafronz.com/tools`) containing reusable packages for GCP-native Go applications. The primary package is a structured logger for Google Cloud Logging (formerly Stackdriver), with a network helper for extracting real client IPs behind proxies/load balancers.

## Commands

All modules are independent—run commands from within each module's directory.

```bash
# Test a module
cd logger/stackdriver && go test ./...
cd logger/stackdriver/v2 && go test ./...
cd netHelpers/addresses && go test ./...

# Run a single test
go test -run TestFunctionName ./...

# Lint (install golangci-lint first)
golangci-lint run ./...

# Vet
go vet ./...

# Update go.mod to latest Go version
go mod tidy
```

## Architecture

This is a **multi-module monorepo**—each subdirectory with a `go.mod` is a fully independent module with its own versioning and dependency graph. Modules are not coupled to the root `go.lafronz.com/tools` module.

### Module Layout

| Module path | Directory | Purpose |
|---|---|---|
| `go.lafronz.com/tools` | `/` | Root (currently empty, no Go files) |
| `go.lafronz.com/tools/logger/stackdriver` | `logger/stackdriver/` | GCL structured logging v1 |
| `go.lafronz.com/tools/logger/stackdriver/v2` | `logger/stackdriver/v2/` | GCL structured logging v2 with trace/span |
| `go.lafronz.com/tools/netHelpers/addresses` | `netHelpers/addresses/` | Real IP extraction from HTTP headers |

### Logger Design

Both logger versions write JSON directly to stdout using Go's `log` package (flags set to 0 to suppress timestamps). The JSON is formatted to match [Google Cloud Logging structured log fields](https://cloud.google.com/logging/docs/structured-logging).

**v1** (`logger/stackdriver/main.go`): Package-level functions with optional labels via `*L` variants (e.g., `Info()` / `InfoL()`).

**v2** (`logger/stackdriver/v2/log.go`): Adds `trace` and `spanID` string parameters to every function. Reads `GOOGLE_CLOUD_PROJECT` env var at `init()` time to format the trace as `projects/{id}/traces/{trace_id}`. SpanID is reformatted from decimal string to hex.

### Modernization Goals

The logger package needs updates to match current Go industry standards. Key areas:

- **Go version**: All `go.mod` files use `go 1.15`; update to a current supported version (1.21+)
- **`slog` integration**: Go 1.21 introduced `log/slog` as the standard structured logging interface; the logger should implement or wrap `slog.Handler`
- **API ergonomics**: The `*L` suffix pattern for labels and positional `trace/spanID` parameters are not idiomatic; prefer functional options or a logger struct with `With()` methods
- **JSON serialization**: The `String()` method builds JSON via `fmt.Sprintf` (fragile); use `encoding/json` marshaling of the full struct consistently
- **Severity casing**: `Warning` severity is emitted as `"Warning"` but all others are uppercase (e.g., `"ERROR"`); GCL expects `"WARNING"`
- **SpanID conversion**: v2 converts spanID string via `fmt.Sprintf("%x", spanID)` which hex-encodes ASCII bytes, not a decimal-to-hex integer conversion
- **No tests**: All packages lack test coverage; tests are needed before refactoring
