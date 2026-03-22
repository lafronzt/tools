// Package stackdriver formats log entries as newline-delimited JSON for
// Google Cloud Logging (formerly Stackdriver). Each entry is written to
// stdout and includes a severity field matching GCL's expected values.
package stackdriver

import (
	"encoding/json"
	"fmt"
	sysLog "log"
	"os"
)

type entry struct {
	Severity string            `json:"severity"`
	Message  string            `json:"message"`
	Labels   map[string]string `json:"labels,omitempty"`
}

func init() {
	sysLog.SetFlags(0)
	sysLog.SetOutput(os.Stdout)
}

func print(e entry) {
	b, err := json.Marshal(e)
	if err != nil {
		sysLog.Printf(`{"severity":"ERROR","message":"stackdriver: failed to marshal log entry: %v"}`, err)
		return
	}
	sysLog.Print(string(b))
}

// Info logs a message at INFO severity.
func Info(body string, t ...interface{}) {
	print(entry{Severity: "INFO", Message: fmt.Sprintf(body, t...)})
}

// InfoL logs a message at INFO severity with labels.
func InfoL(labels map[string]string, body string, t ...interface{}) {
	print(entry{Severity: "INFO", Labels: labels, Message: fmt.Sprintf(body, t...)})
}

// Error logs a message at ERROR severity.
func Error(body string, t ...interface{}) {
	print(entry{Severity: "ERROR", Message: fmt.Sprintf(body, t...)})
}

// ErrorL logs a message at ERROR severity with labels.
func ErrorL(labels map[string]string, body string, t ...interface{}) {
	print(entry{Severity: "ERROR", Labels: labels, Message: fmt.Sprintf(body, t...)})
}

// Critical logs a message at CRITICAL severity.
func Critical(body string, t ...interface{}) {
	print(entry{Severity: "CRITICAL", Message: fmt.Sprintf(body, t...)})
}

// CriticalL logs a message at CRITICAL severity with labels.
func CriticalL(labels map[string]string, body string, t ...interface{}) {
	print(entry{Severity: "CRITICAL", Labels: labels, Message: fmt.Sprintf(body, t...)})
}

// Debug logs a message at DEBUG severity.
func Debug(body string, t ...interface{}) {
	print(entry{Severity: "DEBUG", Message: fmt.Sprintf(body, t...)})
}

// DebugL logs a message at DEBUG severity with labels.
func DebugL(labels map[string]string, body string, t ...interface{}) {
	print(entry{Severity: "DEBUG", Labels: labels, Message: fmt.Sprintf(body, t...)})
}

// Warning logs a message at WARNING severity.
func Warning(body string, t ...interface{}) {
	print(entry{Severity: "WARNING", Message: fmt.Sprintf(body, t...)})
}

// WarningL logs a message at WARNING severity with labels.
func WarningL(labels map[string]string, body string, t ...interface{}) {
	print(entry{Severity: "WARNING", Labels: labels, Message: fmt.Sprintf(body, t...)})
}
