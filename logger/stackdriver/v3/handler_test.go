package stackdriver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	stackdriver "go.lafronz.com/tools/logger/stackdriver/v3"
)

// newTestHandler returns a Handler that writes to buf.
func newTestHandler(buf *bytes.Buffer, opts ...stackdriver.Option) *stackdriver.Handler {
	return stackdriver.New(append([]stackdriver.Option{stackdriver.WithOutput(buf)}, opts...)...)
}

func parseEntry(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	line := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("invalid JSON %q: %v", line, err)
	}
	return m
}

func TestSeverityMapping(t *testing.T) {
	tests := []struct {
		level   slog.Level
		wantSev string
	}{
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARNING"},
		{slog.LevelError, "ERROR"},
		{slog.LevelError + 4, "CRITICAL"},
	}
	for _, tc := range tests {
		t.Run(tc.wantSev, func(t *testing.T) {
			var buf bytes.Buffer
			h := newTestHandler(&buf, stackdriver.WithMinLevel(slog.LevelDebug))
			logger := slog.New(h)
			logger.Log(context.Background(), tc.level, "msg")
			m := parseEntry(t, &buf)
			if got := m["severity"]; got != tc.wantSev {
				t.Errorf("severity = %q, want %q", got, tc.wantSev)
			}
		})
	}
}

func TestEnabledRespectLevel(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf, stackdriver.WithMinLevel(slog.LevelWarn))
	logger := slog.New(h)
	logger.Info("should be suppressed")
	if buf.Len() > 0 {
		t.Errorf("expected no output for suppressed level, got: %s", buf.String())
	}
	logger.Warn("should appear")
	if buf.Len() == 0 {
		t.Error("expected output for warn level")
	}
}

func TestMessage(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	slog.New(h).Info("hello world")
	m := parseEntry(t, &buf)
	if got := m["message"]; got != "hello world" {
		t.Errorf("message = %q, want %q", got, "hello world")
	}
}

func TestTimeField(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	slog.New(h).Info("time test")
	m := parseEntry(t, &buf)
	ts, ok := m["time"].(string)
	if !ok {
		t.Fatal("time field missing or not a string")
	}
	if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
		t.Errorf("time field %q is not RFC3339Nano: %v", ts, err)
	}
}

func TestExtraAttributes(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	slog.New(h).Info("msg", "key", "value", "count", 42)
	m := parseEntry(t, &buf)
	if m["key"] != "value" {
		t.Errorf("key = %v, want %q", m["key"], "value")
	}
	if m["count"] != float64(42) {
		t.Errorf("count = %v, want 42", m["count"])
	}
}

func TestWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	logger := slog.New(h.WithAttrs([]slog.Attr{slog.String("service", "api")}))
	logger.Info("msg")
	m := parseEntry(t, &buf)
	if m["service"] != "api" {
		t.Errorf("service = %v, want %q", m["service"], "api")
	}
}

// TestWithAttrsEmpty verifies the slog spec: WithAttrs with an empty or nil
// slice must return the receiver unchanged (no allocation, no mutation).
func TestWithAttrsEmpty(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)

	h2 := h.WithAttrs(nil)
	if h2 != slog.Handler(h) {
		t.Error("WithAttrs(nil) should return the receiver unchanged")
	}
	h3 := h.WithAttrs([]slog.Attr{})
	if h3 != slog.Handler(h) {
		t.Error("WithAttrs([]slog.Attr{}) should return the receiver unchanged")
	}
}

func TestWithGroup(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	logger := slog.New(h).WithGroup("request")
	logger.Info("msg", "method", "GET")
	m := parseEntry(t, &buf)
	req, ok := m["request"].(map[string]any)
	if !ok {
		t.Fatalf("request group missing or wrong type: %v", m["request"])
	}
	if req["method"] != "GET" {
		t.Errorf("request.method = %v, want GET", req["method"])
	}
}

// TestWithGroupEmpty verifies the slog spec: WithGroup("") must return the
// receiver unchanged.
func TestWithGroupEmpty(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	h2 := h.WithGroup("")
	if h2 != slog.Handler(h) {
		t.Error("WithGroup(\"\") should return the receiver unchanged")
	}
}

// TestNestedWithGroup verifies that stacking multiple WithGroup calls produces
// a correctly nested JSON object.
func TestNestedWithGroup(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	logger := slog.New(h).WithGroup("a").WithGroup("b")
	logger.Info("msg", "key", "val")
	m := parseEntry(t, &buf)

	a, ok := m["a"].(map[string]any)
	if !ok {
		t.Fatalf("expected top-level 'a' group, got: %v", m)
	}
	b, ok := a["b"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'b' nested under 'a', got: %v", a)
	}
	if b["key"] != "val" {
		t.Errorf("key should be under a.b, got: %v", b)
	}
}

// TestWithAttrsBeforeGroup is the critical slog contract test: attributes
// bound via WithAttrs must appear at the nesting level active when WithAttrs
// was called, NOT nested under any groups added afterwards.
func TestWithAttrsBeforeGroup(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)

	// Bind "service" at root level, then open a "request" group.
	logger := slog.New(h.WithAttrs([]slog.Attr{slog.String("service", "api")})).WithGroup("request")
	logger.Info("msg", "method", "GET")
	m := parseEntry(t, &buf)

	// "service" must be at root, not nested under "request".
	if m["service"] != "api" {
		t.Errorf("service should be at root level; full entry: %v", m)
	}
	req, ok := m["request"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'request' group, got: %v", m)
	}
	if req["method"] != "GET" {
		t.Errorf("method should be under request group, got: %v", req)
	}
	if req["service"] != nil {
		t.Errorf("service must NOT appear under request group, got: %v", req)
	}
}

func TestWithTrace(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf, stackdriver.WithProjectID("my-project"))
	logger := slog.New(h.WithTrace("abc123", "span456"))
	logger.Info("traced")
	m := parseEntry(t, &buf)

	wantTrace := "projects/my-project/traces/abc123"
	if got := m["logging.googleapis.com/trace"]; got != wantTrace {
		t.Errorf("trace = %q, want %q", got, wantTrace)
	}
	if got := m["logging.googleapis.com/spanId"]; got != "span456" {
		t.Errorf("spanId = %q, want %q", got, "span456")
	}
}

func TestWithTraceNoProject(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf, stackdriver.WithProjectID(""))
	logger := slog.New(h.WithTrace("abc123", ""))
	logger.Info("traced")
	m := parseEntry(t, &buf)
	if got := m["logging.googleapis.com/trace"]; got != "abc123" {
		t.Errorf("trace without project = %q, want %q", got, "abc123")
	}
}

func TestWithLabel(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	logger := slog.New(h.WithLabel("env", "prod").WithLabel("region", "us-east1"))
	logger.Info("msg")
	m := parseEntry(t, &buf)
	lbls, ok := m["logging.googleapis.com/labels"].(map[string]any)
	if !ok {
		t.Fatalf("labels missing or wrong type")
	}
	if lbls["env"] != "prod" || lbls["region"] != "us-east1" {
		t.Errorf("unexpected labels: %v", lbls)
	}
}

func TestWithLabelsOption(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf, stackdriver.WithLabels(map[string]string{"service": "worker"}))
	slog.New(h).Info("msg")
	m := parseEntry(t, &buf)
	lbls, ok := m["logging.googleapis.com/labels"].(map[string]any)
	if !ok {
		t.Fatalf("labels missing or wrong type")
	}
	if lbls["service"] != "worker" {
		t.Errorf("service label = %v, want worker", lbls["service"])
	}
}

func TestNoTraceFieldsWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	slog.New(h).Info("no trace")
	m := parseEntry(t, &buf)
	if _, ok := m["logging.googleapis.com/trace"]; ok {
		t.Error("trace field should be absent when not set")
	}
	if _, ok := m["logging.googleapis.com/spanId"]; ok {
		t.Error("spanId field should be absent when not set")
	}
}

// TestConcurrentSafety verifies no data race occurs when multiple goroutines
// log through the same handler. Run with -race to catch violations.
func TestConcurrentSafety(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	logger := slog.New(h)
	done := make(chan struct{})
	for range 10 {
		go func() {
			logger.Info("concurrent log")
			done <- struct{}{}
		}()
	}
	for range 10 {
		<-done
	}
}

func TestMessageWithSpecialChars(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	slog.New(h).Info(`say "hello" & <world>`)
	// Output must be valid JSON.
	parseEntry(t, &buf)
}
