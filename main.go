package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	ctx := context.Background()

	// Set the global MeterProvider with a reader that exports via OTLP HTTP every 5 seconds
	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	provider := sdk.NewMeterProvider(
		sdk.WithReader(sdk.NewPeriodicReader(exporter, sdk.WithInterval(5*time.Second))),
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

	// Increment it every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		counter.Add(ctx, 1)
		log.Println("Counter incremented")
	}
}
