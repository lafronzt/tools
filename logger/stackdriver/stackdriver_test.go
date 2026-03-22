package stackdriver

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
)

// captureOutput redirects the syslog output to a buffer for the duration of
// the test and restores it afterwards.
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
		name     string
		fn       func(string, ...interface{})
		wantSev  string
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
			tc.fn("hello")
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

func TestLabels(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(map[string]string, string, ...interface{})
		wantSev string
	}{
		{"InfoL", InfoL, "INFO"},
		{"ErrorL", ErrorL, "ERROR"},
		{"CriticalL", CriticalL, "CRITICAL"},
		{"DebugL", DebugL, "DEBUG"},
		{"WarningL", WarningL, "WARNING"},
	}
	labels := map[string]string{"service": "api", "env": "test"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureOutput(t)
			tc.fn(labels, "msg with %s", "labels")
			m := parseEntry(t, buf)
			if got := m["severity"]; got != tc.wantSev {
				t.Errorf("severity = %q, want %q", got, tc.wantSev)
			}
			if got := m["message"]; got != "msg with labels" {
				t.Errorf("message = %q, want %q", got, "msg with labels")
			}
			lbls, ok := m["labels"].(map[string]any)
			if !ok {
				t.Fatalf("labels missing or wrong type: %v", m["labels"])
			}
			if lbls["service"] != "api" || lbls["env"] != "test" {
				t.Errorf("unexpected labels: %v", lbls)
			}
		})
	}
}

func TestMessageFormatting(t *testing.T) {
	buf := captureOutput(t)
	Info("value is %d and %s", 42, "hello")
	m := parseEntry(t, buf)
	if got := m["message"]; got != "value is 42 and hello" {
		t.Errorf("message = %q, want %q", got, "value is 42 and hello")
	}
}

func TestMessageWithSpecialChars(t *testing.T) {
	buf := captureOutput(t)
	// Double quotes in the message must be properly escaped, not replaced.
	Info(`say "hello"`)
	raw := buf.String()
	if !strings.Contains(raw, `\"hello\"`) {
		t.Errorf("expected escaped quotes in JSON output, got: %s", raw)
	}
	// Must still be valid JSON.
	parseEntry(t, buf)
}

func TestNoLabelsField(t *testing.T) {
	buf := captureOutput(t)
	Info("no labels")
	m := parseEntry(t, buf)
	if _, ok := m["labels"]; ok {
		t.Error("labels field should be absent when no labels are provided")
	}
}
