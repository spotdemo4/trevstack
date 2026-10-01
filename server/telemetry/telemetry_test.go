package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

var signalEnv = []string{
	"OTEL_SDK_DISABLED",
	"OTEL_EXPORTER_OTLP_ENDPOINT",
	"OTEL_TRACES_EXPORTER",
	"OTEL_METRICS_EXPORTER",
	"OTEL_LOGS_EXPORTER",
	"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
	"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
	"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
}

// clearEnv unsets every variable that enables a signal so the host
// environment cannot leak into a test.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range signalEnv {
		t.Setenv(key, "")
	}
}

func TestEnabled(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		signal string
		want   bool
	}{
		{name: "unset", signal: "TRACES", want: false},
		{name: "otlp endpoint", env: map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318"}, signal: "LOGS", want: true},
		{name: "signal endpoint", env: map[string]string{"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": "http://localhost:4318/v1/metrics"}, signal: "METRICS", want: true},
		{name: "other signal endpoint", env: map[string]string{"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": "http://localhost:4318/v1/metrics"}, signal: "TRACES", want: false},
		{name: "signal exporter", env: map[string]string{"OTEL_TRACES_EXPORTER": "console"}, signal: "TRACES", want: true},
		{name: "other signal exporter", env: map[string]string{"OTEL_TRACES_EXPORTER": "console"}, signal: "LOGS", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			if got := enabled(tt.signal); got != tt.want {
				t.Errorf("enabled(%q) = %v, want %v", tt.signal, got, tt.want)
			}
		})
	}
}

func TestNewResource(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantName    string
		wantVersion string
	}{
		{name: "defaults", wantName: ServiceName, wantVersion: "1.2.3"},
		{
			name: "environment overrides",
			env: map[string]string{
				"OTEL_SERVICE_NAME":        "custom",
				"OTEL_RESOURCE_ATTRIBUTES": "service.version=9.9.9",
			},
			wantName:    "custom",
			wantVersion: "9.9.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_SERVICE_NAME", "")
			t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			res, err := newResource(context.Background(), "1.2.3")
			if err != nil {
				t.Fatalf("newResource: %v", err)
			}

			name, _ := res.Set().Value(semconv.ServiceNameKey)
			if got := name.AsString(); got != tt.wantName {
				t.Errorf("service.name = %q, want %q", got, tt.wantName)
			}
			version, _ := res.Set().Value(semconv.ServiceVersionKey)
			if got := version.AsString(); got != tt.wantVersion {
				t.Errorf("service.version = %q, want %q", got, tt.wantVersion)
			}
		})
	}
}

func TestSetupDisabledByDefault(t *testing.T) {
	clearEnv(t)

	shutdown, err := Setup(context.Background(), "test")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	if _, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		t.Error("expected no SDK tracer provider")
	}
	if _, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider); ok {
		t.Error("expected no SDK meter provider")
	}
	if _, ok := global.GetLoggerProvider().(*sdklog.LoggerProvider); ok {
		t.Error("expected no SDK logger provider")
	}
}

func TestSetupSDKDisabled(t *testing.T) {
	clearEnv(t)
	t.Setenv("OTEL_SDK_DISABLED", "true")
	t.Setenv("OTEL_TRACES_EXPORTER", "console")

	shutdown, err := Setup(context.Background(), "test")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	if _, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		t.Error("expected no SDK tracer provider")
	}
}

func TestSetupNoneExporter(t *testing.T) {
	clearEnv(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	shutdown, err := Setup(context.Background(), "test")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	if _, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		t.Error("expected no SDK tracer provider")
	}
}

// TestSetupConsole must run last since it installs global providers.
func TestSetupConsole(t *testing.T) {
	clearEnv(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "console")
	t.Setenv("OTEL_METRICS_EXPORTER", "console")
	t.Setenv("OTEL_LOGS_EXPORTER", "console")

	shutdown, err := Setup(context.Background(), "test")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	if _, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); !ok {
		t.Error("expected SDK tracer provider")
	}
	if _, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider); !ok {
		t.Error("expected SDK meter provider")
	}
	if _, ok := global.GetLoggerProvider().(*sdklog.LoggerProvider); !ok {
		t.Error("expected SDK logger provider")
	}

	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}
