package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

// URL represents the structure of a shortened URL.
type URL struct {
	ID           string    `json:"id"`
	OriginalURL  string    `json:"original_url"`
	ShortURL     string    `json:"short_url"`
	CreationDate time.Time `json:"creation_date"`
	Author       string    `json:"author"`
}

var urlDB = make(map[string]URL) // In-memory database

// generateShortURL returns a hashed short string from original URL
func generateShortURL(originalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalURL))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash[:8] // Return first 8 characters
}

// createURL stores and returns the shortened URL string
func createURL(originalURL string) string {
	shortURL := generateShortURL(originalURL)
	urlDB[shortURL] = URL{
		ID:           shortURL,
		OriginalURL:  originalURL,
		ShortURL:     shortURL,
		CreationDate: time.Now(),
		Author:       "Modit Grover",
	}
	return shortURL
}

// getURL returns the full URL for a given short ID
func getURL(id string) (URL, error) {
	url, ok := urlDB[id]
	if !ok {
		return URL{}, errors.New("URL not found")
	}
	return url, nil
}

// RootPageURL handles the root path
func RootPageURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to TinyGoURL - URL Shortener Service"))
}

// ShortURLHandler handles URL shortening requests
func ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || !isValidURL(requestData.URL) {
		http.Error(w, "Invalid or malformed URL", http.StatusBadRequest)
		return
	}

	shortURL := createURL(requestData.URL)

	response := struct {
		ShortURL string `json:"short_url"`
	}{
		ShortURL: "http://localhost:3000/redirect/" + shortURL,
	}

	json.NewEncoder(w).Encode(response)
}

// redirectURLHandler redirects short URL to the original one
func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/redirect/")
	url, err := getURL(id)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

// isValidURL is a basic URL format validator
func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// main starts the HTTP server
func main() {
	http.HandleFunc("/", RootPageURL)
	http.HandleFunc("/shorten", ShortURLHandler)
	http.HandleFunc("/redirect/", redirectURLHandler)

	log.Println("🚀 Server started at http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
