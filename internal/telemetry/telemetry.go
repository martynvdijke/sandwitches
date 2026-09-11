// Package telemetry wires OpenTelemetry traces, metrics and logs for the
// server. Configuration is driven entirely by the standard OTEL_* environment
// variables and the whole subsystem is a no-op unless an OTLP endpoint is
// configured, so deployments without a collector behave exactly as before.
package telemetry

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// DefaultServiceName is used when OTEL_SERVICE_NAME is not set.
const DefaultServiceName = "sandwitches"

// Telemetry owns the lifetime of the OpenTelemetry providers.
type Telemetry struct {
	// Enabled reports whether any OTLP signal is configured.
	Enabled bool
	// LogWriter forwards standard-library log lines to OTLP logs. It is nil
	// when the logs signal is disabled.
	LogWriter io.Writer

	shutdown func(context.Context) error
}

// Shutdown flushes and stops every configured provider. It is safe to call on
// a nil or disabled Telemetry.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t == nil || t.shutdown == nil {
		return nil
	}
	return t.shutdown(ctx)
}

// GinMiddleware returns the request tracing/metrics middleware, or nil when
// telemetry is disabled.
func (t *Telemetry) GinMiddleware() gin.HandlerFunc {
	if t == nil || !t.Enabled {
		return nil
	}
	return otelgin.Middleware(DefaultServiceName,
		otelgin.WithPropagators(otel.GetTextMapPropagator()))
}

// Setup configures the OTLP exporters and SDK providers from the standard
// OTEL_* environment variables. It returns a Telemetry that is disabled
// (no-op) unless a generic or per-signal OTLP endpoint is set, or when
// OTEL_SDK_DISABLED=true. Per-signal endpoints enable only that signal.
func Setup(ctx context.Context, version string) *Telemetry {
	t := &Telemetry{shutdown: func(context.Context) error { return nil }}

	if os.Getenv("OTEL_SDK_DISABLED") == "true" {
		return t
	}

	general := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != ""
	tracesOn := general || os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""
	metricsOn := general || os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT") != ""
	logsOn := general || os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT") != ""
	if !tracesOn && !metricsOn && !logsOn {
		return t
	}

	res, err := newResource(ctx, version)
	if err != nil {
		log.Printf("telemetry: resource detection failed, using default: %v", err)
		res = resource.Default()
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var shutdowns []func(context.Context) error

	if tracesOn {
		exp, err := otlptracehttp.New(ctx)
		if err != nil {
			log.Printf("telemetry: traces disabled: %v", err)
		} else {
			tp := sdktrace.NewTracerProvider(
				sdktrace.WithResource(res),
				sdktrace.WithBatcher(exp),
			)
			otel.SetTracerProvider(tp)
			shutdowns = append(shutdowns, tp.Shutdown)
		}
	}

	if metricsOn {
		exp, err := otlpmetrichttp.New(ctx)
		if err != nil {
			log.Printf("telemetry: metrics disabled: %v", err)
		} else {
			mp := sdkmetric.NewMeterProvider(
				sdkmetric.WithResource(res),
				sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
			)
			otel.SetMeterProvider(mp)
			shutdowns = append(shutdowns, mp.Shutdown)
		}
	}

	if logsOn {
		exp, err := otlploghttp.New(ctx)
		if err != nil {
			log.Printf("telemetry: logs disabled: %v", err)
		} else {
			lp := sdklog.NewLoggerProvider(
				sdklog.WithResource(res),
				sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
			)
			otellog.SetLoggerProvider(lp)

			handler := otelslog.NewHandler(DefaultServiceName,
				otelslog.WithLoggerProvider(lp),
				otelslog.WithVersion(version),
			)
			slog.SetDefault(slog.New(handler))
			t.LogWriter = slog.NewLogLogger(handler, slog.LevelInfo).Writer()
			shutdowns = append(shutdowns, lp.Shutdown)
		}
	}

	t.Enabled = len(shutdowns) > 0
	t.shutdown = func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdowns {
			errs = append(errs, fn(ctx))
		}
		return errors.Join(errs...)
	}
	return t
}

func newResource(ctx context.Context, version string) (*resource.Resource, error) {
	name := os.Getenv("OTEL_SERVICE_NAME")
	if name == "" {
		name = DefaultServiceName
	}
	instance, err := os.Hostname()
	if err != nil || instance == "" {
		instance = uuid.NewString()
	}
	return resource.New(ctx,
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			attribute.String("service.name", name),
			attribute.String("service.version", version),
			attribute.String("service.instance.id", instance),
		),
	)
}
