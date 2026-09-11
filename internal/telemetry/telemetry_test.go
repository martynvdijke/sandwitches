package telemetry

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

func clearOTELEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"OTEL_SDK_DISABLED",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		t.Setenv(k, "")
	}
}

func shutdownWithTimeout(t *testing.T, tel *Telemetry) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return tel.Shutdown(ctx)
}

func TestSetupDisabledByDefault(t *testing.T) {
	clearOTELEnv(t)

	tel := Setup(context.Background(), "test")
	if tel.Enabled {
		t.Fatal("telemetry should be disabled without an endpoint")
	}
	if tel.LogWriter != nil {
		t.Fatal("log writer should be nil when logs are disabled")
	}
	if tel.GinMiddleware() != nil {
		t.Fatal("gin middleware should be nil when disabled")
	}
	if err := shutdownWithTimeout(t, tel); err != nil {
		t.Fatalf("no-op shutdown returned error: %v", err)
	}
}

func TestSetupEnabledByEndpoint(t *testing.T) {
	clearOTELEnv(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4318")
	t.Cleanup(func() { otel.SetTracerProvider(noop.NewTracerProvider()) })

	tel := Setup(context.Background(), "test")
	if !tel.Enabled {
		t.Fatal("telemetry should be enabled with an endpoint")
	}
	if tel.LogWriter == nil {
		t.Fatal("log writer should be set when logs are enabled")
	}
	if tel.GinMiddleware() == nil {
		t.Fatal("gin middleware should be set when enabled")
	}
	// The collector is unreachable, so export errors on shutdown are expected.
	_ = shutdownWithTimeout(t, tel)
}

func TestSetupForceDisabled(t *testing.T) {
	clearOTELEnv(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4318")
	t.Setenv("OTEL_SDK_DISABLED", "true")

	if tel := Setup(context.Background(), "test"); tel.Enabled {
		t.Fatal("OTEL_SDK_DISABLED=true must disable telemetry")
	}
}

func TestNewResourceServiceName(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "custom-name")

	res, err := newResource(context.Background(), "1.2.3")
	if err != nil {
		t.Fatalf("newResource: %v", err)
	}

	var name, version string
	for _, kv := range res.Attributes() {
		switch string(kv.Key) {
		case "service.name":
			name = kv.Value.AsString()
		case "service.version":
			version = kv.Value.AsString()
		}
	}
	if name != "custom-name" {
		t.Errorf("service.name = %q, want custom-name", name)
	}
	if version != "1.2.3" {
		t.Errorf("service.version = %q, want 1.2.3", version)
	}
}
