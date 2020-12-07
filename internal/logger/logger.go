package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/glassonion1/logz/internal/severity"
	"github.com/glassonion1/logz/internal/tracer"
)

var NowFunc = time.Now

type LogEntry struct {
	Severity    severity.Severity `json:"severity,string"`
	Message     string            `json:"message"`
	Time        time.Time         `json:"time"`
	Trace       string            `json:"logging.googleapis.com/trace"`
	SpanID      string            `json:"logging.googleapis.com/spanId"`
	JSONPayload interface{}       `json:"jsonPayload"`
	HTTPRequest *HttpRequest      `json:"httpRequest,omitempty"`
}

type HttpRequest struct {
	RequestMethod string `json:"requestMethod"`
	RequestURL    string `json:"requestUrl"`
	Latency       string `json:"latency"`
	UserAgent     string `json:"userAgent"`
	RemoteIP      string `json:"remoteIp"`
	Status        int32  `json:"status"`
	Protocol      string `json:"protocol"`
	RequestSize   string `json:"requestSize"`
	ResponseSize  string `json:"responseSize"`
}

// Looger is for GCP
type Logger struct {
	ProjectID string
}

// New creates an Looger instance
func New() *Logger {
	// In case of App Engine, the value can be obtained.
	// Otherwise, it is an empty string.
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	return &Logger{
		ProjectID: projectID,
	}
}

// WriteLog writes a log to stdout
func (l *Logger) WriteLog(ctx context.Context, severity severity.Severity, format string, a ...interface{}) {
	// Gets the traceID and spanID
	traceID, spanID := tracer.TraceIDAndSpanID(ctx)

	trace := fmt.Sprintf("projects/%s/traces/%s", l.ProjectID, traceID)
	msg := fmt.Sprintf(format, a...)
	ety := &LogEntry{
		Severity: severity,
		Message:  msg,
		Time:     NowFunc(),
		Trace:    trace,
		SpanID:   spanID,
	}

	if err := json.NewEncoder(os.Stdout).Encode(ety); err != nil {
		panic(err)
	}
}
