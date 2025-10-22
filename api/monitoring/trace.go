package monitoring

import (
	"io"
	"runtime/debug"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// WriterTracerProvider returns a trace provider that writes pretty printed traces to w.
func WriterTracerProvider(w io.Writer) (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(w), stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	res, err := resources()
	if err != nil {
		return nil, err
	}
	return trace.NewTracerProvider(trace.WithBatcher(exporter), trace.WithResource(res)), nil
}

// resources return standard trace resources for our service.
func resources() (*resource.Resource, error) {
	version := "dev"
	info, _ := debug.ReadBuildInfo()
	for _, v := range info.Settings {
		if v.Key == "vcs.revision" {
			version = v.Value
		}
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("kart-api"),
			semconv.ServiceVersion(version),
		),
	)
}
