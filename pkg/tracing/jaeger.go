package tracing

import (
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupJaeger sets up Jaeger with some defaults
func SetupJaeger(serviceName string) (io.Closer, error) {
	// Create and install Jaeger export pipeline
	flush, err := jaeger.InstallNewPipeline(
		// TODO possible to send to multiple sinks?
		//jaeger.WithCollectorEndpoint("http://10.187.96.171:14268/api/traces"), // Jaeger VM. Could maybe create a headless service hitting the VM.
		//jaeger.WithCollectorEndpoint("http://10.193.44.58:14268/api/traces"), // WF proxyHacked in IP. Could maybe create a headless service hitting the VM.
		jaeger.WithAgentEndpoint("localhost:6831"), // Local jaeger agent daemonset.
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		return nil, err
	}

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return funcCloser{f: flush}, nil
}

type funcCloser struct {
	f func()
}

func (c funcCloser) Close() error {
	c.f()
	return nil
}
