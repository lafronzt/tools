# stackdriver/v3

A [`slog.Handler`](https://pkg.go.dev/log/slog#Handler) that formats log entries as newline-delimited JSON for [Google Cloud Logging](https://cloud.google.com/logging/docs/structured-logging) (formerly Stackdriver).

This is the recommended logger for new code. For projects already using v1 or v2, v3 is a drop-in replacement that integrates with Go's standard `log/slog` API.

## Installation

```bash
go get go.lafronz.com/tools/logger/stackdriver/v3
```

## Usage

```go
import (
    "log/slog"
    stackdriver "go.lafronz.com/tools/logger/stackdriver/v3"
)

// Basic setup — reads GOOGLE_CLOUD_PROJECT from the environment.
handler := stackdriver.New()
logger := slog.New(handler)

logger.Info("server started", "port", 8080)
logger.Error("request failed", "error", err, "status", 500)
```

### With options

```go
handler := stackdriver.New(
    stackdriver.WithProjectID("my-gcp-project"),
    stackdriver.WithMinLevel(slog.LevelDebug),
    stackdriver.WithLabels(map[string]string{"service": "api", "env": "prod"}),
)
```

### Per-request trace context

```go
// Attach Cloud Trace IDs (hex strings) to a request-scoped logger.
rlog := slog.New(handler.WithTrace(traceID, spanID))
rlog.Info("handling request", "method", r.Method, "path", r.URL.Path)
```

### Adding labels at runtime

```go
logger := slog.New(handler.WithLabel("version", "v1.2.3"))
```

### Standard slog features

All standard `slog` patterns work: `With`, `WithGroup`, structured attributes, etc.

```go
// Pre-attach fields to every subsequent log entry.
logger = logger.With("request_id", requestID)

// Group attributes under a nested JSON key.
logger.Info("db query", slog.Group("db", "table", "users", "rows", 42))
```

## Output format

Each log entry is a single JSON object:

```json
{
  "severity": "INFO",
  "message": "handling request",
  "time": "2024-01-15T10:30:00.000000000Z",
  "logging.googleapis.com/trace": "projects/my-project/traces/abc123",
  "logging.googleapis.com/spanId": "def456",
  "logging.googleapis.com/labels": {"env": "prod"},
  "method": "GET",
  "path": "/api/v1/users"
}
```

## Severity mapping

| `slog` level | GCL severity |
|---|---|
| `Debug` | `DEBUG` |
| `Info` | `INFO` |
| `Warn` | `WARNING` |
| `Error` | `ERROR` |
| `Error+4` and above | `CRITICAL` |
