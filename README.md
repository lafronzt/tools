# go.lafronz.com/tools

A collection of Go utility packages for GCP-native applications.

## Packages

### Logger — Google Cloud Logging

Structured JSON logging formatted for [Google Cloud Logging](https://cloud.google.com/logging/docs/structured-logging).

| Module | Description |
|---|---|
| [`logger/stackdriver/v3`](logger/stackdriver/v3/) | **Recommended.** `slog.Handler` implementation. |
| [`logger/stackdriver/v2`](logger/stackdriver/v2/) | Package-level functions with Cloud Trace support. |
| [`logger/stackdriver`](logger/stackdriver/) | Package-level functions, no trace support. |

**Quick start (v3):**

```go
import (
    "log/slog"
    stackdriver "go.lafronz.com/tools/logger/stackdriver/v3"
)

logger := slog.New(stackdriver.New())
logger.Info("server started", "port", 8080)

// Per-request with trace context
rlog := slog.New(stackdriver.New().WithTrace(traceID, spanID))
rlog.Info("handling request")
```

See [`logger/stackdriver/v3/README.md`](logger/stackdriver/v3/README.md) for full documentation.

### Network Helpers

```go
import "go.lafronz.com/tools/netHelpers/addresses"

ip := addresses.GetRealIP(r) // checks X-Real-IP, X-Forwarded-For, RemoteAddr
```

## Requirements

Go 1.24+. All modules are zero-dependency (standard library only).
