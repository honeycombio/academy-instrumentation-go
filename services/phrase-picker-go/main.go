package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// phrasesList holds the collection of phrases to choose from
var phrasesList = []string{
	"you're muted",
	"not dead yet",
	"Let them.",
	"Boiling Loves Company!",
	"Must we?",
	"SRE not-sorry",
	"Honeycomb at home",
	"There is no cloud",
	"This is fine",
	"It's a trap!",
	"Not Today",
	"You had one job",
	"bruh",
	"have you tried restarting?",
	"try again after coffee",
	"deploy != release",
	"oh, just the crimes",
	"not a bug, it's a feature",
	"test in prod",
	"who broke the build?",
}

// Phrase is a struct to map the JSON output
type Phrase struct {
	Phrase string `json:"phrase"`
}

func main() {
	ctx := context.Background()

	// Configure a new OTLP exporter using environment variables for sending data to Honeycomb over gRPC
	client := otlptracehttp.NewClient()
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %e", err)
	}

	// Create a new tracer provider with a batch span processor and the otlp exporter
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
	)

	// Handle shutdown to ensure all sub processes are closed correctly and telemetry is exported
	defer func() {
		_ = exp.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
	}()

	// Register the global Tracer provider
	otel.SetTracerProvider(tp)

	// Register the W3C trace context and baggage propagators so data is propagated across services/processes
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Implement an HTTP handler func to be instrumented
	handler := http.HandlerFunc(phraseHandler)

	// Setup handler instrumentation
	wrappedHandler := otelhttp.NewHandler(handler, "hello")
	http.Handle("/phrase", wrappedHandler)

	// Start the HTTP server
	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":10114", nil))

}

func phraseHandler(w http.ResponseWriter, r *http.Request) {
	// Select a random phrase
	randomIndex := rand.Intn(len(phrasesList))
	selectedPhrase := phrasesList[randomIndex]

	// Create a Phrase struct with the selected phrase
	response := Phrase{Phrase: selectedPhrase}

	// Set content type and CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Encode the response as JSON
	json.NewEncoder(w).Encode(response)
}
