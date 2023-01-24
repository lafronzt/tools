// Package stackdriver is used to format the logs to better fit StackDriver JSON filtering
package stackdriver

import (
	"encoding/json"
	"fmt"
	sysLog "log"
	"os"
	"strings"
)

type log struct {
	Severity string            `json:"severity"`
	Message  string            `json:"message"`
	Labels   map[string]string `json:"labels"`
	Trace    string            `json:"trace"`
	SpanID   string            `json:"spanId"`
}

var GCPProjectID string

func init() {
	sysLog.SetFlags(0)
	sysLog.SetOutput(os.Stdout)
	GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
}

func (l log) String() string {
	l.Message = strings.ReplaceAll(l.Message, "\"", "'")

	if l.Trace != "" && l.SpanID != "" {

		// SpanID is a decimal number, but StackDriver expects a hex number
		// so we convert it here
		l.SpanID = fmt.Sprintf("%x", l.SpanID)

		if len(l.Labels) == 0 {
			return fmt.Sprintf(
				"{\"severity\":\"%v\", \"message\":\"%v\", \"trace\":\"%v\", \"spanId\":\"%v\"}",
				l.Severity, l.Message, l.Trace, l.SpanID)

		} else {
			lblStr, err := json.Marshal(l.Labels)
			if err != nil {
				// return as if there were no labels
				return fmt.Sprintf(
					"{\"severity\":\"%v\", \"message\":\"%v\", \"trace\":\"%v\", \"spanId\":\"%v\"}",
					l.Severity, l.Message, l.Trace, l.SpanID)
			}

			return fmt.Sprintf(
				"{\"severity\":\"%v\", \"message\":\"%v\", \"trace\":\"%v\", \"spanId\":\"%v\", \"labels\":%v}",
				l.Severity, l.Message, l.Trace, l.SpanID, string(lblStr))
		}

	} else {

		if len(l.Labels) == 0 {
			return fmt.Sprintf(
				"{\"severity\":\"%v\", \"message\":\"%v\"}", l.Severity, l.Message)

		} else {
			lblStr, err := json.Marshal(l.Labels)
			if err != nil {
				// return as if there were no labels
				return fmt.Sprintf(
					"{\"severity\":\"%v\", \"message\":\"%v\"",
					l.Severity, l.Message)
			}

			return fmt.Sprintf(
				"{\"severity\":\"%v\", \"message\":\"%v\", \"labels\":%v}",
				l.Severity, l.Message, string(lblStr))
		}
	}
}

// Info formats the logs as an info message parsing for StackDriver
func Info(trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "INFO", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// InfoL formats the logs as an info message parsing for StackDriver with a label
func InfoL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "INFO", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// Error formats the logs for error message parsing for StackDriver
func Error(trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "ERROR", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// ErrorL formats the logs for error message parsing for StackDriver with a label
func ErrorL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "ERROR", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// Critical formats the logs for error message parsing for StackDriver
func Critical(trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "CRITICAL", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// CriticalL formats the logs for error message parsing for StackDriver with a label
func CriticalL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "CRITICAL", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// Debug formats the logs for error message parsing for StackDriver
func Debug(trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "DEBUG", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// DebugL formats the logs for error message parsing for StackDriver with a label
func DebugL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "DEBUG", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// Warning formats the logs for error message parsing for StackDriver
func Warning(trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "Warning", Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}

// WarningL formats the logs for error message parsing for StackDriver with a label
func WarningL(labels map[string]string, trace string, spanID string, body string, l ...interface{}) {
	sysLog.Print(log{Severity: "Warning", Labels: labels, Message: fmt.Sprintf(body, l...), Trace: trace, SpanID: spanID}.String())
}
