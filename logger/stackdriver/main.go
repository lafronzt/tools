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
}

func init() {
	sysLog.SetFlags(0)
	sysLog.SetOutput(os.Stdout)
}

func (l log) String() string {
	l.Message = strings.ReplaceAll(l.Message, "\"", "'")

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

// Info formats the logs as an info message parsing for StackDriver
func Info(body string, t ...interface{}) {
	sysLog.Print(log{Severity: "INFO", Message: fmt.Sprintf(body, t...)}.String())
}

func InfoL(labels map[string]string, body string, t ...interface{}) {
	sysLog.Print(log{Severity: "INFO", Labels: labels, Message: fmt.Sprintf(body, t...)}.String())
}

// Error formats the logs for error message parsing for StackDriver
func Error(body string, t ...interface{}) {
	sysLog.Print(log{Severity: "ERROR", Message: fmt.Sprintf(body, t...)}.String())
}

func ErrorL(labels map[string]string, body string, t ...interface{}) {
	sysLog.Print(log{Severity: "ERROR", Labels: labels, Message: fmt.Sprintf(body, t...)}.String())
}

// Critical formats the logs for error message parsing for StackDriver
func Critical(body string, t ...interface{}) {
	sysLog.Print(log{Severity: "CRITICAL", Message: fmt.Sprintf(body, t...)}.String())
}

func CriticalL(labels map[string]string, body string, t ...interface{}) {
	sysLog.Print(log{Severity: "CRITICAL", Labels: labels, Message: fmt.Sprintf(body, t...)}.String())
}

// Debug formats the logs for error message parsing for StackDriver
func Debug(body string, t ...interface{}) {
	sysLog.Print(log{Severity: "DEBUG", Message: fmt.Sprintf(body, t...)}.String())
}

func DebugL(labels map[string]string, body string, t ...interface{}) {
	sysLog.Print(log{Severity: "DEBUG", Labels: labels, Message: fmt.Sprintf(body, t...)}.String())
}

// Warning formats the logs for error message parsing for StackDriver
func Warning(body string, t ...interface{}) {
	sysLog.Print(log{Severity: "Warning", Message: fmt.Sprintf(body, t...)}.String())
}

func WarningL(labels map[string]string, body string, t ...interface{}) {
	sysLog.Print(log{Severity: "Warning", Labels: labels, Message: fmt.Sprintf(body, t...)}.String())
}
