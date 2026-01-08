package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
)

var (
	keyToURL = make(map[string]string)
	urlToKey = make(map[string]string)
	mutex    = &sync.Mutex{}
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if key, exists := urlToKey[req.URL]; exists {
		response := ShortenResponse{
			URL: req.URL,
			Key: key,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	key := generateKey()
	for _, exists := keyToURL[key]; exists; {
		key = generateKey()
	}

	keyToURL[key] = req.URL
	urlToKey[req.URL] = key

	response := ShortenResponse{
		URL: req.URL,
		Key: key,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	key := make([]byte, 6)
	for i := range key {
		key[i] = charset[rand.Intn(len(charset))]
	}
	return string(key)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	key := strings.TrimPrefix(path, "/go/")

	mutex.Lock()
	url, exists := keyToURL[key]
	mutex.Unlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", shortenHandler)
	mux.HandleFunc("/go/", redirectHandler)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Server listening on port %d...\n", *port)
	_ = http.ListenAndServe(addr, mux)
}
