package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
	// Initialize OpenTelemetry Tracer
	tracerProvider, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() { _ = tracerProvider.Shutdown(context.Background()) }()

	// create a new echo instance
	e := echo.New()

	// Use the OpenTelemetry Echo Middleware
	e.Use(otelecho.Middleware("phrase-picker"))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health check endpoint
	e.GET("/health", healthCheckHandler)

	// define a route '/phrase'
	e.GET("/phrase", phraseHandler)

	// start the server on the specified port
	e.Logger.Fatal(e.Start(":10118"))
}

func phraseHandler(c echo.Context) error {
	// select a random phrase
	randomIndex := rand.Intn(len(phrasesList))
	selectedPhrase := phrasesList[randomIndex]

	// create a Phrase struct with the selected phrase
	response := Phrase{Phrase: selectedPhrase}

	// return the response
	return c.JSON(http.StatusOK, response)
}

func healthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

func initTracer() (*sdktrace.TracerProvider, error) {

	ctx := context.Background()
	// Configure a new OTLP exporter using environment variables for sending data to Honeycomb over gRPC
	client := otlptracehttp.NewClient()
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %e", err)
	}

	// Create a new trace provider with the exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)

	// Register the trace provider as the global provider
	otel.SetTracerProvider(tp)

	// Register the W3C trace context and baggage propagators so data is propagated across services/processes
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp, nil
}
