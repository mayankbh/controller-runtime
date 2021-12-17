package tracing

import (
	"context"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

// SetupOTLP sets up a global trace provider sending to OpenTelemetry with some defaults
func SetupOTLP(serviceName string) (io.Closer, error) {
	exp, err := otlp.NewExporter(
		otlp.WithInsecure(),
		otlp.WithAddress("otlp-collector.default:55680"),
	)
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	// TODO ignore errors? idk.
	resource, err := resource.New(context.Background(), resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)))
	if err != nil {
		panic(err)
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)

	return otlpCloser{exp: exp, bsp: bsp}, nil
}

type otlpCloser struct {
	exp *otlp.Exporter
	bsp *sdktrace.BatchSpanProcessor
}

func (s otlpCloser) Close() error {
	s.bsp.Shutdown(context.Background()) // shutdown the processor
	return s.exp.Shutdown(context.Background())
}
