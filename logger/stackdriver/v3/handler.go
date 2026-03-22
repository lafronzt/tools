// Package stackdriver provides a [slog.Handler] that emits newline-delimited
// JSON to stdout formatted for Google Cloud Logging (formerly Stackdriver).
//
// # Quick start
//
//	handler := stackdriver.New()
//	logger := slog.New(handler)
//	logger.Info("server started", "port", 8080)
//
// # Trace context
//
// Stamp every entry on a request path with its Cloud Trace identifiers:
//
//	rlog := handler.WithTrace(traceID, spanID)
//	logger := slog.New(rlog)
//	logger.Info("handling request", "method", r.Method)
//
// # Labels
//
// GCL labels are indexed and filterable in the console:
//
//	h := stackdriver.New(stackdriver.WithLabels(map[string]string{"env": "prod"}))
package stackdriver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
	"sync"
	"time"
)

// Handler implements [slog.Handler] for Google Cloud Logging.
// The zero value is not usable; create one with [New].
type Handler struct {
	projectID string
	out       io.Writer
	level     slog.Leveler
	mu        *sync.Mutex

	// Copied on every WithAttrs / WithGroup / clone call.
	preAttrs []slog.Attr
	groups   []string
	trace    string
	spanID   string
	labels   map[string]string
}

// Option configures a [Handler].
type Option func(*Handler)

// WithProjectID sets the GCP project ID used to build the full trace resource
// name (projects/{id}/traces/{trace_id}). Defaults to the
// GOOGLE_CLOUD_PROJECT environment variable.
func WithProjectID(id string) Option {
	return func(h *Handler) { h.projectID = id }
}

// WithOutput redirects log output. Defaults to [os.Stdout].
func WithOutput(w io.Writer) Option {
	return func(h *Handler) { h.out = w }
}

// WithMinLevel sets the minimum level that is handled. Defaults to
// [slog.LevelInfo].
func WithMinLevel(l slog.Leveler) Option {
	return func(h *Handler) { h.level = l }
}

// WithLabels pre-populates GCL labels that appear on every log entry.
func WithLabels(labels map[string]string) Option {
	return func(h *Handler) { maps.Copy(h.labels, labels) }
}

// New creates a [Handler] ready for use with [slog.New].
func New(opts ...Option) *Handler {
	h := &Handler{
		projectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		out:       os.Stdout,
		level:     slog.LevelInfo,
		mu:        &sync.Mutex{},
		labels:    make(map[string]string),
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// WithTrace returns a new Handler that stamps every log entry with the given
// Cloud Trace trace ID and span ID. Both are expected to be hex strings.
func (h *Handler) WithTrace(traceID, spanID string) *Handler {
	h2 := h.clone()
	h2.trace = traceID
	h2.spanID = spanID
	return h2
}

// WithLabel returns a new Handler with an additional GCL label attached to
// every log entry. Labels are indexed and filterable in the GCL console.
func (h *Handler) WithLabel(key, value string) *Handler {
	h2 := h.clone()
	h2.labels[key] = value
	return h2
}

// Enabled reports whether the handler handles records at the given level.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle formats r as a GCL JSON log entry and writes it to the output.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	entry := map[string]any{
		"severity": gcpSeverity(r.Level),
		"message":  r.Message,
		"time":     r.Time.UTC().Format(time.RFC3339Nano),
	}

	if h.trace != "" {
		if h.projectID != "" {
			entry["logging.googleapis.com/trace"] = fmt.Sprintf("projects/%s/traces/%s", h.projectID, h.trace)
		} else {
			entry["logging.googleapis.com/trace"] = h.trace
		}
	}
	if h.spanID != "" {
		entry["logging.googleapis.com/spanId"] = h.spanID
	}
	if len(h.labels) > 0 {
		entry["logging.googleapis.com/labels"] = h.labels
	}

	// Merge pre-added attrs (from WithAttrs) and record attrs, respecting
	// any active group nesting set by WithGroup.
	allAttrs := make([]slog.Attr, 0, len(h.preAttrs)+r.NumAttrs())
	allAttrs = append(allAttrs, h.preAttrs...)
	r.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})
	addAttrs(entry, h.groups, allAttrs)

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = h.out.Write(data)
	return err
}

// WithAttrs returns a new Handler with the given attrs added to every record.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := h.clone()
	h2.preAttrs = append(h2.preAttrs, attrs...)
	return h2
}

// WithGroup returns a new Handler that nests subsequent attributes under a
// JSON object with the given key.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groups = append(h2.groups, name)
	return h2
}

// clone returns a copy of h with independent slice and map fields.
// The mutex is shared so all clones serialize writes to the same writer.
func (h *Handler) clone() *Handler {
	h2 := *h
	h2.labels = maps.Clone(h.labels)
	h2.preAttrs = slices.Clone(h.preAttrs)
	h2.groups = slices.Clone(h.groups)
	return &h2
}

// addAttrs inserts attrs into dst, nesting them under any active groups.
func addAttrs(dst map[string]any, groups []string, attrs []slog.Attr) {
	if len(groups) == 0 {
		for _, a := range attrs {
			resolveAttr(dst, a)
		}
		return
	}
	child, _ := dst[groups[0]].(map[string]any)
	if child == nil {
		child = make(map[string]any)
	}
	addAttrs(child, groups[1:], attrs)
	dst[groups[0]] = child
}

// resolveAttr recursively adds a single slog.Attr into dst.
func resolveAttr(dst map[string]any, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		sub := make(map[string]any)
		for _, ga := range a.Value.Group() {
			resolveAttr(sub, ga)
		}
		if a.Key == "" {
			// Inline group: merge directly into the parent map.
			for k, v := range sub {
				dst[k] = v
			}
		} else {
			dst[a.Key] = sub
		}
		return
	}
	dst[a.Key] = a.Value.Any()
}

// gcpSeverity maps slog levels to GCL severity strings.
// Levels above ERROR are mapped to CRITICAL.
func gcpSeverity(level slog.Level) string {
	switch {
	case level < slog.LevelInfo:
		return "DEBUG"
	case level < slog.LevelWarn:
		return "INFO"
	case level < slog.LevelError:
		return "WARNING"
	case level < slog.LevelError+4:
		return "ERROR"
	default:
		return "CRITICAL"
	}
}
