package monitoring

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const (
	// See expected otel trace key names from https://opentelemetry.io/docs/specs/otlp/#json-protobuf-encoding
	KeyTraceID = "traceId"
	KeySpanID  = "spanId"
)

// NewJSON returns a new JSON logger with standard configuration.
func NewJSONLogger(to io.Writer, level slog.Level) *slog.Logger {
	h := slog.NewJSONHandler(to, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(TraceHandler{next: h})
}

var _ slog.Handler = TraceHandler{}

// TraceHandler is a [slog.Handler] that adds traceId and spanId attributes to logs when present within ctx.
type TraceHandler struct {
	next slog.Handler
}

// Enabled implements [slog.Handler.Enabled].
func (t TraceHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return t.next.Enabled(ctx, l)
}

// Handle implements [slog.Handler.Handle].
func (t TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	trace := trace.SpanContextFromContext(ctx)
	if trace.HasSpanID() {
		r.AddAttrs(slog.String(KeySpanID, trace.SpanID().String()))
	}
	if trace.HasTraceID() {
		r.AddAttrs(
			slog.String(KeyTraceID, trace.TraceID().String()),
		)
	}
	return t.next.Handle(ctx, r)
}

// WithAttr implements [slog.Handler.WithAttrs].
func (t TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return TraceHandler{
		next: t.next.WithAttrs(attrs),
	}
}

// WithGroup implements [slog.Handler.WithGroup].
func (t TraceHandler) WithGroup(name string) slog.Handler {
	return TraceHandler{
		next: t.next.WithGroup(name),
	}
}
