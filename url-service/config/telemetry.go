package config

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const defaultOTLPEndpoint = "http://localhost:4318/v1/traces"

func NewTraceProvider() (*sdktrace.TracerProvider, func(), error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpointURL(getEnv("OTEL_OTLP_ENDPOINT", defaultOTLPEndpoint)),
	)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attribute.String("service.name", getEnv("OTEL_SERVICE_NAME", "url-service")),
		),
	)
	if err != nil {
		res = resource.Default()
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = provider.Shutdown(shutdownCtx)
	}

	return provider, shutdown, nil
}