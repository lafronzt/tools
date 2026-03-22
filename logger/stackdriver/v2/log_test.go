package stackdriver

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
)

func captureOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })
	return &buf
}

func parseEntry(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON output %q: %v", buf.String(), err)
	}
	return m
}

func TestSeverities(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(string, string, string, ...interface{})
		wantSev string
	}{
		{"Info", Info, "INFO"},
		{"Error", Error, "ERROR"},
		{"Critical", Critical, "CRITICAL"},
		{"Debug", Debug, "DEBUG"},
		{"Warning", Warning, "WARNING"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureOutput(t)
			tc.fn("", "", "hello")
			m := parseEntry(t, buf)
			if got := m["severity"]; got != tc.wantSev {
				t.Errorf("severity = %q, want %q", got, tc.wantSev)
			}
			if got := m["message"]; got != "hello" {
				t.Errorf("message = %q, want %q", got, "hello")
			}
		})
	}
}

func TestTraceFormatting(t *testing.T) {
	GCPProjectID = "my-project"
	buf := captureOutput(t)
	Info("trace-abc", "67890", "traced request")
	m := parseEntry(t, buf)

	wantTrace := "projects/my-project/traces/trace-abc"
	if got := m["logging.googleapis.com/trace"]; got != wantTrace {
		t.Errorf("trace = %q, want %q", got, wantTrace)
	}
	// SpanID 67890 decimal → 0x10932 → "0000000000010932"
	wantSpan := "0000000000010932"
	if got := m["logging.googleapis.com/spanId"]; got != wantSpan {
		t.Errorf("spanId = %q, want %q", got, wantSpan)
	}
}

func TestTraceWithNoProject(t *testing.T) {
	GCPProjectID = ""
	buf := captureOutput(t)
	Info("raw-trace-id", "0", "msg")
	m := parseEntry(t, buf)
	if got := m["logging.googleapis.com/trace"]; got != "raw-trace-id" {
		t.Errorf("trace without project = %q, want %q", got, "raw-trace-id")
	}
}

func TestEmptyTraceOmitted(t *testing.T) {
	buf := captureOutput(t)
	Info("", "", "no trace")
	m := parseEntry(t, buf)
	if _, ok := m["logging.googleapis.com/trace"]; ok {
		t.Error("trace field should be absent when empty")
	}
	if _, ok := m["logging.googleapis.com/spanId"]; ok {
		t.Error("spanId field should be absent when empty")
	}
}

func TestLabels(t *testing.T) {
	buf := captureOutput(t)
	labels := map[string]string{"service": "api"}
	InfoL(labels, "", "", "with labels")
	m := parseEntry(t, buf)
	lbls, ok := m["labels"].(map[string]any)
	if !ok {
		t.Fatalf("labels missing or wrong type")
	}
	if lbls["service"] != "api" {
		t.Errorf("unexpected labels: %v", lbls)
	}
}

func TestMessageWithSpecialChars(t *testing.T) {
	buf := captureOutput(t)
	Info("", "", `say "hello"`)
	raw := buf.String()
	if !strings.Contains(raw, `\"hello\"`) {
		t.Errorf("expected escaped quotes in JSON output, got: %s", raw)
	}
	parseEntry(t, buf)
}

func TestSpanIDAlreadyHex(t *testing.T) {
	buf := captureOutput(t)
	// If the caller passes a non-decimal span ID (e.g. already hex), it is
	// returned as-is because ParseUint in base 10 will fail.
	Info("", "deadbeef", "hex span")
	m := parseEntry(t, buf)
	if got := m["logging.googleapis.com/spanId"]; got != "deadbeef" {
		t.Errorf("spanId = %q, want %q", got, "deadbeef")
	}
}
