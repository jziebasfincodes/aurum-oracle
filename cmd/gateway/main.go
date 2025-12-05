// api_gateway.go - Tiered Access Gateway with Hardcoded Port
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type APIKey struct {
	Key        string
	ClientName string
	RateLimit  int 
	Tier       string // "free" or "paid"
}

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
}

var (
	validAPIKeys = map[string]APIKey{
		// FREE TIER (Delayed Data)
		"YOUR_DEMO_KEY": {Key: "YOUR_DEMO_KEY", ClientName: "Free User", RateLimit: 60, Tier: "free"},
		
		// PAID TIER (Real-Time)
		"YOUR_PAID_KEY": {Key: "YOUR_PAID_KEY", ClientName: "Paid User A", RateLimit: 1000, Tier: "paid"},
	}
	rateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}
	aggregatorURL = "http://localhost:9000"
)

// --- Rate Limiter ---
func (rl *RateLimiter) StartCleanupService() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, times := range rl.requests {
				var active []time.Time
				for _, t := range times {
					if now.Sub(t) < 1*time.Minute {
						active = append(active, t)
					}
				}
				if len(active) == 0 { delete(rl.requests, ip) } else { rl.requests[ip] = active }
			}
			rl.mu.Unlock()
		}
	}()
}

func (rl *RateLimiter) Allow(apiKey string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)
	requests, exists := rl.requests[apiKey]
	if !exists {
		rl.requests[apiKey] = []time.Time{now}
		return true
	}
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) { validRequests = append(validRequests, reqTime) }
	}
	if len(validRequests) >= limit {
		rl.requests[apiKey] = validRequests
		return false
	}
	validRequests = append(validRequests, now)
	rl.requests[apiKey] = validRequests
	return true
}

// --- Handlers ---

func authenticate(r *http.Request) (*APIKey, error) {
	key := r.Header.Get("X-API-Key")
	if key == "" { key = r.URL.Query().Get("api_key") }
	keyInfo, ok := validAPIKeys[key]
	if !ok { return nil, fmt.Errorf("invalid API key") }
	if !rateLimiter.Allow(key, keyInfo.RateLimit) { return nil, fmt.Errorf("rate limit exceeded") }
	return &keyInfo, nil
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" { w.WriteHeader(http.StatusOK); return }

	clientInfo, err := authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	
	// --- TIER ENFORCEMENT LOGIC ---
	targetURL := aggregatorURL + r.URL.Path
	
	// If Free Tier, force delayed parameter
	if clientInfo.Tier == "free" && r.URL.Path == "/price" {
		targetURL += "?delayed=true"
	}
	
	// Proxy
	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Oracle Consensus Unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	w.Header().Set("X-Client", clientInfo.ClientName)
	w.Header().Set("X-Tier", clientInfo.Tier)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	if url := os.Getenv("AGGREGATOR_URL"); url != "" { aggregatorURL = url }
	
	// HARDCODED PORT 3000 (Critical Fix)
	port := "3000"
	
	rateLimiter.StartCleanupService()
	http.HandleFunc("/", proxyHandler)
	log.Printf("🛡️  AURUM API Gateway (Tiered) Active on port :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}