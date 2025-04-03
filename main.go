package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

var cache = make(map[string]string) // Simple cache (URL -> Response)
var cacheMutex = sync.Mutex{}       // Ensures safe access to the cache

func main() {
	// Command-line arguments
	port := flag.Int("port", 3000, "Port to run the server on")
	origin := flag.String("origin", "", "Origin server to forward requests to")
	clearCache := flag.Bool("clear-cache", false, "Clear the cache")
	flag.Parse()

	if *clearCache {
		clearCacheData()
		return
	}

	if *origin == "" {
		fmt.Println("Error: Origin URL is required!")
		return
	}

	// Start the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, *origin)
	})

	fmt.Printf("Starting caching proxy on port %d...\n", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request, origin string) {
	url := origin + r.URL.Path

	cacheMutex.Lock()
	data, found := cache[url]
	cacheMutex.Unlock()

	if found {
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(data))
		fmt.Println("Cache HIT:", url)
		return
	}

	// Fetch from the origin server
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Error reaching origin server", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Store in cache
	cacheMutex.Lock()
	cache[url] = string(body)
	cacheMutex.Unlock()

	w.Header().Set("X-Cache", "MISS")
	w.Write(body)
	fmt.Println("Cache MISS:", url)
}

func clearCacheData() {
	cacheMutex.Lock()
	cache = make(map[string]string)
	cacheMutex.Unlock()
	fmt.Println("Cache cleared successfully!")
}
