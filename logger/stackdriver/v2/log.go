// Package stackdriver formats log entries as newline-delimited JSON for
// Google Cloud Logging (formerly Stackdriver). This v2 package extends v1
// with Cloud Trace support: each log entry can carry a trace ID and span ID
// that correlate it with a distributed trace in GCL.
//
// The GCP project ID is read from the GOOGLE_CLOUD_PROJECT environment
// variable at init time, and is used to build the full trace resource name
// (projects/{id}/traces/{trace_id}) required by GCL.
package stackdriver

import (
	"encoding/json"
	"fmt"
	sysLog "log"
	"os"
	"strconv"
)

// GCPProjectID is the GCP project used to format trace resource names.
// It is populated from the GOOGLE_CLOUD_PROJECT environment variable at
// init time but may be overridden by the caller.
var GCPProjectID string

func init() {
	sysLog.SetFlags(0)
	sysLog.SetOutput(os.Stdout)
	GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
}

type entry struct {
	Severity string            `json:"severity"`
	Message  string            `json:"message"`
	Labels   map[string]string `json:"labels,omitempty"`
	Trace    string            `json:"logging.googleapis.com/trace,omitempty"`
	SpanID   string            `json:"logging.googleapis.com/spanId,omitempty"`
}

// formatTrace returns a fully-qualified GCL trace resource name.
func formatTrace(traceID string) string {
	if GCPProjectID != "" {
		return fmt.Sprintf("projects/%s/traces/%s", GCPProjectID, traceID)
	}
	return traceID
}

// formatSpanID converts a decimal span ID string to a zero-padded 16-char
// hex string as required by GCL. If the value is already non-decimal (e.g.
// already hex), it is returned unchanged.
func formatSpanID(spanID string) string {
	if n, err := strconv.ParseUint(spanID, 10, 64); err == nil {
		return fmt.Sprintf("%016x", n)
	}
	return spanID
}

func print(e entry) {
	if e.Trace != "" {
		e.Trace = formatTrace(e.Trace)
	}
	if e.SpanID != "" {
		e.SpanID = formatSpanID(e.SpanID)
	}
	b, err := json.Marshal(e)
	if err != nil {
		sysLog.Printf(`{"severity":"ERROR","message":"stackdriver: failed to marshal log entry: %v"}`, err)
		return
	}
	sysLog.Print(string(b))
}

// Info logs a message at INFO severity with optional trace context.
func Info(trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "INFO", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// InfoL logs a message at INFO severity with labels and optional trace context.
func InfoL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "INFO", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// Error logs a message at ERROR severity with optional trace context.
func Error(trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "ERROR", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// ErrorL logs a message at ERROR severity with labels and optional trace context.
func ErrorL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "ERROR", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// Critical logs a message at CRITICAL severity with optional trace context.
func Critical(trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "CRITICAL", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// CriticalL logs a message at CRITICAL severity with labels and optional trace context.
func CriticalL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "CRITICAL", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// Debug logs a message at DEBUG severity with optional trace context.
func Debug(trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "DEBUG", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// DebugL logs a message at DEBUG severity with labels and optional trace context.
func DebugL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "DEBUG", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// Warning logs a message at WARNING severity with optional trace context.
func Warning(trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "WARNING", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}

// WarningL logs a message at WARNING severity with labels and optional trace context.
func WarningL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	print(entry{Severity: "WARNING", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID})
}
