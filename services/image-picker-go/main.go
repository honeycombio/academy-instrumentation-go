package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// filename holds the collection of image files to choose from
var filenames = []string{
	"Angrybird.JPG",
	"Arco&Tub.png",
	"IMG_9343.jpg",
	"angry-lemon-ufo.JPG",
	"austintiara4.png",
	"baby-geese.jpg",
	"bbq.jpg",
	"beach.JPG",
	"bunny-mask.jpg",
	"busted-light.jpg",
	"cat-glowing-eyes.JPG",
	"cat-on-leash.JPG",
	"cat.jpg",
	"clementine.png",
	"cow-peeking.jpg",
	"different-animals-01.png",
	"dratini.png",
	"everything-is-an-experiment.png",
	"experiment.png",
	"fine-food.jpg",
	"flower.jpg",
	"frenwho.png",
	"genshin-spa.jpg",
	"grass-and-desert-guy.png",
	"honeycomb-dogfood-logo.png",
	"horse-maybe.png",
	"is-this-emeri.png",
	"jean-and-statue.png",
	"jessitron.png",
	"keys-drying.jpg",
	"lime-on-soap-dispenser.jpg",
	"loki-closeup.jpg",
	"lynia.png",
	"ninguang-at-work.png",
	"paul-r-allen.png",
	"please.png",
	"roswell-nose.jpg",
	"roswell.JPG",
	"salt-packets-in-jar.jpg",
	"scarred-character.png",
	"square-leaf-with-nuts.jpg",
	"stu.jpeg",
	"sweating-it.png",
	"tanuki.png",
	"tennessee-sunset.JPG",
	"this-is-fine-trash.jpg",
	"three-pillars-2.png",
	"trash-flat.jpg",
	"walrus-painting.jpg",
	"windigo.png",
	"yellow-lines.JPG",
}

// ImageUrl is a struct to map the JSON output
type ImageUrl struct {
	ImageUrl string `json:"imageUrl"`
}

var bucketName string
var imageUrls []string

// init is special function that gets called before main
func init() {
	// Get the bucket name from the environment variable, with a default value
	bucketName = os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		bucketName = "random-pictures"
	}

	for _, filename := range filenames {
		url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, filename)
		imageUrls = append(imageUrls, url)
	}
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
	e.Use(otelecho.Middleware("image-picker"))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health check endpoint
	e.GET("/health", healthCheckHandler)

	// define a route '/imageUrl'
	e.GET("/imageUrl", imageUrlHandler)

	// start the server on the specified port
	e.Logger.Fatal(e.Start(":10116"))
}

func imageUrlHandler(c echo.Context) error {

	// select a random image url
	randomIndex := rand.Intn(len(imageUrls))
	selectedUrl := imageUrls[randomIndex]

	// create a image url struct with the selected image url
	response := ImageUrl{ImageUrl: selectedUrl}

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
