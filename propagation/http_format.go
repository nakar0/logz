package propagation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	httpHeader = "X-Cloud-Trace-Context"
)

// HTTPFormat propagator serializes SpanContext to/from HTTP Headers.
type HTTPFormat struct{}

var _ propagation.TextMapPropagator = &HTTPFormat{}

// Inject injects a context into the carrier as HTTP headers.
func (hf HTTPFormat) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()

	if !sc.TraceID.IsValid() || !sc.SpanID.IsValid() {
		return
	}

	header := fmt.Sprintf("%s/%s;o=%d", sc.TraceID.String(), sc.SpanID.String(), sc.TraceFlags)
	carrier.Set(httpHeader, header)
}

// Extract extracts a context from the carrier if it contains HTTP headers.
func (hf HTTPFormat) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	if h := carrier.Get(httpHeader); h != "" {
		sc, err := extract(h)
		if err == nil && sc.IsValid() {
			return trace.ContextWithRemoteSpanContext(ctx, sc)
		}
	}

	return ctx
}

func extract(h string) (trace.SpanContext, error) {
	sc := trace.SpanContext{}

	// Parse the trace id field.
	slash := strings.Index(h, `/`)
	if slash == -1 {
		return sc, errors.New("failed to parse value")
	}
	tid, h := h[:slash], h[slash+1:]

	traceID, err := trace.TraceIDFromHex(tid)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}

	sc.TraceID = traceID

	// Parse the span id field.
	semicolon := strings.Index(h, `;`)
	if semicolon == -1 {
		return sc, errors.New("failed to parse value")
	}
	sid, h := h[:semicolon], h[semicolon+1:]
	spanID, err := trace.SpanIDFromHex(sid)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}

	sc.SpanID = spanID

	// Parse the options field, options field is optional.
	if !strings.HasPrefix(h, "o=") {
		return sc, errors.New("failed to parse value")
	}

	if h[2:] == "1" {
		sc.TraceFlags = trace.FlagsSampled
	} else if h[2:] == "2" {
		sc.TraceFlags = trace.FlagsDeferred
	} else if h[2:] == "4" {
		sc.TraceFlags = trace.FlagsDebug
	}

	return sc, nil
}

func (hf HTTPFormat) Fields() []string {
	return []string{httpHeader}
}