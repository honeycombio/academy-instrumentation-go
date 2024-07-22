package main

import (
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// create a new echo instance
	e := echo.New()

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
