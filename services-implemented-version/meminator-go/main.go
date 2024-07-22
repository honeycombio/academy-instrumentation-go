package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	imageMaxWidthPx  = 1000
	imageMaxHeightPx = 1000
)

type Response struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

type Request struct {
	Phrase   string `json:"phrase"`
	ImageURL string `json:"imageUrl"`
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
	e.Use(otelecho.Middleware("meminator"))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health check endpoint
	e.GET("/health", healthCheckHandler)

	// define a route '/applyPhraseToPicture'
	e.POST("/applyPhraseToPicture", meminateHandler)

	// start the server on the specified port
	e.Logger.Fatal(e.Start(":10117"))
}

func meminateHandler(c echo.Context) error {
	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	phrase := ""
	if req.Phrase != "" {
		phrase = req.Phrase
	}
	imageURL := ""
	if req.ImageURL != "" {
		imageURL = req.ImageURL
	}

	inputImagePath, err := downloadImage(imageURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to download image"})
	}
	defer os.Remove(inputImagePath)

	outputImagePath := generateRandomFilename(inputImagePath)

	cmd := exec.Command("convert",
		inputImagePath,
		"-resize", fmt.Sprintf("%dx%d>", imageMaxWidthPx, imageMaxHeightPx),
		"-gravity", "North",
		"-pointsize", "48",
		"-fill", "white",
		"-undercolor", "#00000080",
		"-font", "Angkor-Regular",
		"-annotate", "0", phrase,
		outputImagePath)

	if err := cmd.Run(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Subprocess failed with return code: %v", err)})
	}

	defer os.Remove(outputImagePath)
	return c.File(outputImagePath)
}

func healthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

func downloadImage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: %s", resp.Status)
	}

	extension := getFileExtension(url)
	tempFile, err := os.CreateTemp("", fmt.Sprintf("*%s", extension))

	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// GenerateRandomFilename generates a random filename with the same extension as the input filename.
func generateRandomFilename(inputFilename string) string {
	// Extract the extension from the input filename
	extension := getFileExtension(inputFilename)

	// Generate a UUID and convert it to a string without dashes
	randomUUID := uuid.New().String()
	randomFilename := strings.ReplaceAll(randomUUID, "-", "")

	// Append the extension to the random filename
	randomFilenameWithExtension := randomFilename + extension
	randomFilepath := filepath.Join("/tmp", randomFilenameWithExtension)

	return randomFilepath
}

// GetFileExtension extracts the file extension from a URL or filename.
func getFileExtension(url string) string {
	parts := strings.Split(url, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
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
