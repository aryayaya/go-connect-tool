package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go-connect-tool/checker"
	"go-connect-tool/store"
)

var (
	dataStore *store.Store
)

func main() {
	var err error
	dataStore, err = store.NewStore("data.json")
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API Endpoints
	http.HandleFunc("/api/sites", handleSites)
	http.HandleFunc("/api/proxy", handleProxy)
	http.HandleFunc("/api/test", handleTest)

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleSites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(dataStore.GetSites())
	case "POST":
		var site store.Site
		if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Generate ID if missing (simple timestamp based)
		if site.ID == "" {
			site.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		dataStore.AddSite(site)
		json.NewEncoder(w).Encode(site)
	case "DELETE":
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id", http.StatusBadRequest)
			return
		}
		dataStore.RemoveSite(id)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(dataStore.GetProxy())
	case "POST":
		var config store.ProxyConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dataStore.UpdateProxy(config)
		json.NewEncoder(w).Encode(config)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional: Test specific site ID
	// For now, let's just test all sites and return results
	// Or we can accept a list of IDs to test.
	// Let's implement "Test All" for simplicity in this handler.

	sites := dataStore.GetSites()
	proxy := dataStore.GetProxy()
	results := make([]checker.CheckResult, len(sites))

	var wg sync.WaitGroup
	resultChan := make(chan checker.CheckResult, len(sites))

	for _, site := range sites {
		wg.Add(1)
		go func(s store.Site) {
			defer wg.Done()
			resultChan <- checker.CheckSite(s, proxy)
		}(site)
	}

	wg.Wait()
	close(resultChan)

	i := 0
	for res := range resultChan {
		results[i] = res
		i++
	}

	json.NewEncoder(w).Encode(results)
}
