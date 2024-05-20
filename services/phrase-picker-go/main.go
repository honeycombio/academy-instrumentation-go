package main

import (
    "encoding/json"
    "math/rand"
    "net/http"
    "time"
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
    // Seed the random number generator
    rand.Seed(time.Now().UnixNano())

    // Setup HTTP route and handler
    http.HandleFunc("/phrase", phraseHandler)

    // Start the HTTP server
    http.ListenAndServe(":8080", nil)
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
