package telemetry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// ServiceName is reported unless OTEL_SERVICE_NAME overrides it.
const ServiceName = "trevstack-server"

// Setup installs global OpenTelemetry providers for each signal that is
// configured through the standard OTEL_* environment variables.
//
// A signal is enabled when OTEL_EXPORTER_OTLP_ENDPOINT, its signal-specific
// OTLP endpoint, or its OTEL_{TRACES,METRICS,LOGS}_EXPORTER is set, so the
// server exports nothing by default. OTEL_SDK_DISABLED=true disables everything.
//
// version is reported as service.version unless OTEL_RESOURCE_ATTRIBUTES
// overrides it.
//
// The returned function flushes and stops every provider that was installed.
func Setup(ctx context.Context, version string) (func(context.Context) error, error) {
	var shutdowns []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdowns {
			errs = append(errs, fn(ctx))
		}
		return errors.Join(errs...)
	}

	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") {
		return shutdown, nil
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	res, err := newResource(ctx, version)
	if err != nil {
		return shutdown, fmt.Errorf("could not create resource: %w", err)
	}

	if enabled("TRACES") {
		exporter, err := autoexport.NewSpanExporter(ctx)
		if err != nil {
			return shutdown, fmt.Errorf("could not create span exporter: %w", err)
		}
		if !autoexport.IsNoneSpanExporter(exporter) {
			provider := sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(exporter),
				sdktrace.WithResource(res),
			)
			shutdowns = append(shutdowns, provider.Shutdown)
			otel.SetTracerProvider(provider)
		}
	}

	if enabled("METRICS") {
		reader, err := autoexport.NewMetricReader(ctx)
		if err != nil {
			return shutdown, fmt.Errorf("could not create metric reader: %w", err)
		}
		if !autoexport.IsNoneMetricReader(reader) {
			provider := sdkmetric.NewMeterProvider(
				sdkmetric.WithReader(reader),
				sdkmetric.WithResource(res),
			)
			shutdowns = append(shutdowns, provider.Shutdown)
			otel.SetMeterProvider(provider)

			err = runtime.Start()
			if err != nil {
				return shutdown, fmt.Errorf("could not start runtime metrics: %w", err)
			}
		}
	}

	if enabled("LOGS") {
		exporter, err := autoexport.NewLogExporter(ctx)
		if err != nil {
			return shutdown, fmt.Errorf("could not create log exporter: %w", err)
		}
		if !autoexport.IsNoneLogExporter(exporter) {
			provider := sdklog.NewLoggerProvider(
				sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
				sdklog.WithResource(res),
			)
			shutdowns = append(shutdowns, provider.Shutdown)
			global.SetLoggerProvider(provider)
		}
	}

	return shutdown, nil
}

// newResource describes the server. Detectors later in the list take
// precedence, so OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES override the
// defaults.
func newResource(ctx context.Context, version string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(ServiceName),
			semconv.ServiceVersion(version),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
	)
}

// enabled reports whether the given signal (TRACES, METRICS, or LOGS) has been
// configured. autoexport would otherwise default every signal to OTLP and try
// to reach a collector on localhost.
func enabled(signal string) bool {
	return os.Getenv("OTEL_"+signal+"_EXPORTER") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_"+signal+"_ENDPOINT") != ""
}
