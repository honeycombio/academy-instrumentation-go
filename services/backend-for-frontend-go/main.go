package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	port         = 10115
	imagePicker  = "http://image-picker:10116/imageUrl"
	meminator    = "http://meminator:10117/applyPhraseToPicture"
	phrasePicker = "http://phrase-picker:10118/phrase"
)

type FetchOptions struct {
	Method string
	Body   interface{}
}

func fetchFromService(url string, options *FetchOptions) (*http.Response, error) {
	var (
		req *http.Request
		err error
	)

	if options != nil && options.Body != nil {
		body, err := json.Marshal(options.Body)
		if err != nil {
			return nil, err
		}
		req, _ = http.NewRequest(options.Method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	return client.Do(req)
}

func createPicture(w http.ResponseWriter, r *http.Request) {
	phraseResponse, err := fetchFromService(phrasePicker, nil)
	if err != nil || phraseResponse.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch phrase", http.StatusInternalServerError)
		return
	}
	defer phraseResponse.Body.Close()
	var phraseResult map[string]interface{}
	if err := json.NewDecoder(phraseResponse.Body).Decode(&phraseResult); err != nil {
		http.Error(w, "Failed to decode phrase response", http.StatusInternalServerError)
		return
	}

	imageResponse, err := fetchFromService(imagePicker, nil)
	if err != nil || imageResponse.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}
	defer imageResponse.Body.Close()
	var imageResult map[string]interface{}
	if err := json.NewDecoder(imageResponse.Body).Decode(&imageResult); err != nil {
		http.Error(w, "Failed to decode image response", http.StatusInternalServerError)
		return
	}

	meminatorResponse, err := fetchFromService(meminator, &FetchOptions{
		Method: "POST",
		Body:   mergeMaps(phraseResult, imageResult),
	})
	if err != nil || meminatorResponse.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch picture from meminator", http.StatusInternalServerError)
		return
	}
	defer meminatorResponse.Body.Close()

	w.Header().Set("Content-Type", "image/png")
	if _, err := io.Copy(w, meminatorResponse.Body); err != nil {
		http.Error(w, "Failed to stream picture data", http.StatusInternalServerError)
	}
}

func mergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {

	http.HandleFunc("/createPicture", createPicture)
	http.HandleFunc("/health", healthCheck)

	fmt.Printf("Server is running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
