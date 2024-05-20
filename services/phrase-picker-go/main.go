package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
)

// phrasesList holds the collection of phrases to choose from
var phrasesList = []string{
    "Hello, world!",
    "Today is a wonderful day!",
    "Programming in Go is fun!",
    "Keep learning new things.",
    "Make it work, make it right, make it fast.",
}

// Phrase is a struct to map the JSON output
type Phrase struct {
    Phrase string `json:"phrase"`
}

func main() {

    // Setup HTTP route and handler
    http.HandleFunc("/phrase", phraseHandler)

    // Start the HTTP server
    // Start the HTTP server
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
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
