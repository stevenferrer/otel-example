package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricglobal "go.opentelemetry.io/otel/metric/global"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkres "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	otlpAgentAddrEnv     = "OTEL_EXPORTER_OTLP_ENDPOINT"
	defaultOtlpAgentAddr = "0.0.0.0:4317"
)

// Init initializes telemetry providers
func Init(ctx context.Context, svcName string) (func() error, error) {
	// Refer to https://opentelemetry.io/docs/instrumentation/go/getting-started/#creating-a-resource
	res, err := sdkres.New(ctx,
		sdkres.WithFromEnv(),
		sdkres.WithProcess(),
		sdkres.WithTelemetrySDK(),
		sdkres.WithHost(),
		sdkres.WithAttributes(semconv.ServiceNameKey.String(svcName)),
	)
	if err != nil {
		return nil, fmt.Errorf("new resource: %w", err)
	}

	otelAgentAddr, ok := os.LookupEnv(otlpAgentAddrEnv)
	if !ok {
		otelAgentAddr = defaultOtlpAgentAddr
	}

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(otelAgentAddr),
	)
	if err != nil {
		return nil, fmt.Errorf("new metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(2*time.Second),
			),
		),
	)
	metricglobal.SetMeterProvider(meterProvider)

	return func() error {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := meterProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown meter provider: %w", err)
		}

		return nil
	}, nil
}
