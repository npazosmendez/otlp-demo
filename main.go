package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func main() {
	ctx := context.Background()

	// Set the global MeterProvider with a reader that exports via OTLP HTTP every 5 seconds
	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(5*time.Second))),
	)
	otel.SetMeterProvider(provider)

	// Create the counter
	meter := otel.Meter("otlp-example")
	counter, err := meter.Int64Counter("example.counter",
		metric.WithDescription("The example counter that is incremented every 5 seconds"),
	)
	if err != nil {
		log.Fatalf("failed to create counter metric: %v", err)
	}

	// Create OTLP log exporter
	logexporter, err := otlploghttp.New(ctx)
	if err != nil {
		log.Fatalf("failed to create log logexporter: %v", err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logexporter)),
	)
	global.SetLoggerProvider(lp)

	// Create OTLP trace exporter
	traceexporter, err := otlptracehttp.New(ctx)
	if err != nil {
		log.Fatalf("failed to create trace traceexporter: %v", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceexporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("example-service"),
		)),
	)
	otel.SetTracerProvider(tp)

	// Create a tracer
	tracer := otel.Tracer("example-tracer")

	// Get the logger
	logger := otelslog.NewLogger("example-logger")

	// Increment it every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		// counter
		counter.Add(ctx, 1)
		log.Println("Counter incremented")

		// log
		logger.InfoContext(ctx, "Example log", "now", time.Now().Unix())

		_, span := tracer.Start(ctx, "example-span")

		// Simulate some work
		time.Sleep(1 * time.Second)
		span.End()
	}
}
