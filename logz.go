package logz

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/glassonion1/logz/internal/config"
	"github.com/glassonion1/logz/internal/logger"
	"github.com/glassonion1/logz/internal/severity"
	"github.com/glassonion1/logz/internal/types"
	logzpropagation "github.com/glassonion1/logz/propagation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func init() {
	// In case of App Engine, the value can be obtained.
	// Otherwise, it is an empty string.
	config.ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

	config.WriteAccessLog = logger.WriteAccessLog
}

// Config is configurations for logz
type Config struct {
	// GCP Project ID
	ProjectID string
	// Whether or not to write the access log
	WritesAccessLog bool
}

// SetProjectID sets gcp project id to the logger
func SetProjectID(projectID string) {
	config.ProjectID = projectID
}

// SetConfig sets config to the logger
func SetConfig(conf Config) {
	if conf.ProjectID != "" {
		config.ProjectID = conf.ProjectID
	}
	if !conf.WritesAccessLog {
		config.WriteAccessLog = types.WriteEmptyAccessLog
	}
}

// Debugf writes debug log to the stdout
func Debugf(ctx context.Context, format string, a ...interface{}) {
	logger.WriteApplicationLog(ctx, severity.Debug, format, a...)
}

// Infof writes info log to the stdout
func Infof(ctx context.Context, format string, a ...interface{}) {
	logger.WriteApplicationLog(ctx, severity.Info, format, a...)
}

// Warningf writes warning log to the stdout
func Warningf(ctx context.Context, format string, a ...interface{}) {
	logger.WriteApplicationLog(ctx, severity.Warning, format, a...)
}

// Errorf writes error log to the stdout
func Errorf(ctx context.Context, format string, a ...interface{}) {
	logger.WriteApplicationLog(ctx, severity.Error, format, a...)
}

// Criticalf writes critical log to the stdout
func Criticalf(ctx context.Context, format string, a ...interface{}) {
	logger.WriteApplicationLog(ctx, severity.Critical, format, a...)
}

// Access writes access log to the stderr
func Access(ctx context.Context, r http.Request, statusCode, responseSize int, elapsed time.Duration) {
	req := types.MakeHTTPRequest(r, statusCode, responseSize, elapsed)
	config.WriteAccessLog(ctx, req)
}

// AccessLog writes access log to the stderr without http.Request
func AccessLog(ctx context.Context, method, url, userAgent, remoteIP, protocol string, statusCode, requestSize, responseSize int, elapsed time.Duration) {
	req := types.HTTPRequest{
		RequestMethod: method,
		RequestURL:    url,
		RequestSize:   fmt.Sprintf("%d", requestSize),
		Status:        statusCode,
		ResponseSize:  fmt.Sprintf("%d", responseSize),
		UserAgent:     userAgent,
		RemoteIP:      remoteIP,
		Latency:       types.MakeDuration(elapsed),
		Protocol:      protocol,
	}
	config.WriteAccessLog(ctx, req)
}

// InitTracer initializes OpenTelemetry tracer
func InitTracer() {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)

	props := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		logzpropagation.HTTPFormat{})
	otel.SetTextMapPropagator(props)
}

// StartCollectingSeverity starts collectiong severity
func StartCollectingSeverity(ctx context.Context) context.Context {
	cs := &severity.ContextSeverity{}
	return severity.SetContextSeverity(ctx, cs)
}
