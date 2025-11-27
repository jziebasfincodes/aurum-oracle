// oracle_node.go - The "Ultimate" Multi-Source Oracle
// Sources: GoldAPI.io , GoldAPI.com, Swissquote, Binance, Kraken, Investing
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_PORT = "8080"

type PriceSource struct {
	Name  string
	Type  string // "fiat" or "crypto"
	Fetch func(keys map[string]string) (float64, error)
}

var sources = []PriceSource{
	// 1. GoldAPI.io (The New Paid Source)
	{
		Name: "GoldAPI_IO",
		Type: "fiat",
		Fetch: func(keys map[string]string) (float64, error) {
			apiKey := keys["GOLDAPI_IO_KEY"]
			if apiKey == "" { return 0, fmt.Errorf("missing API key") }
			
			client := &http.Client{Timeout: 10 * time.Second}
			req, _ := http.NewRequest("GET", "https://www.goldapi.io/api/XAU/USD", nil)
			req.Header.Set("x-access-token", apiKey)
			
			resp, err := client.Do(req)
			if err != nil { return 0, err }
			defer resp.Body.Close()
			
			if resp.StatusCode != 200 { return 0, fmt.Errorf("status %d", resp.StatusCode) }
			
			var data struct { Price float64 `json:"price"` }
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return 0, err }
			return data.Price, nil
		},
	},
	// 2. Gold-API.com (Backup)
	{
		Name: "GoldAPI_COM",
		Type: "fiat",
		Fetch: func(keys map[string]string) (float64, error) {
			apiKey := keys["GOLDAPI_COM_KEY"]
			if apiKey == "" { return 0, fmt.Errorf("missing API key") }
			
			client := &http.Client{Timeout: 10 * time.Second}
			req, _ := http.NewRequest("GET", "https://gold-api.com/api/XAU/USD", nil)
			req.Header.Set("Authorization", "Bearer " + apiKey)
			
			resp, err := client.Do(req)
			if err != nil { return 0, err }
			defer resp.Body.Close()
			
			if resp.StatusCode != 200 { return 0, fmt.Errorf("status %d", resp.StatusCode) }
			
			var data struct { Price float64 `json:"price"` }
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return 0, err }
			return data.Price, nil
		},
	},
	// 3. Swissquote (Forex)
	{
		Name: "Swissquote",
		Type: "fiat",
		Fetch: func(keys map[string]string) (float64, error) {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get("https://forex-data-feed.swissquote.com/public-quotes/bboquotes/instrument/XAU/USD")
			if err != nil { return 0, err }
			defer resp.Body.Close()
			var data []struct {
				Topo struct { Ask float64 `json:"ask"` } `json:"topo"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return 0, err }
			if len(data) == 0 { return 0, fmt.Errorf("empty response") }
			return data[0].Topo.Ask, nil
		},
	},
	// 4. Binance PAXG (Crypto)
	{
		Name: "Binance",
		Type: "crypto",
		Fetch: func(keys map[string]string) (float64, error) {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get("https://api.binance.us/api/v3/ticker/price?symbol=PAXGUSDT")
			if err != nil {
				resp, err = client.Get("https://api.binance.com/api/v3/ticker/price?symbol=PAXGUSDT")
			}
			if err != nil { return 0, err }
			defer resp.Body.Close()
			var data struct { Price string `json:"price"` }
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return 0, err }
			return strconv.ParseFloat(data.Price, 64)
		},
	},
	// 5. Kraken PAXG (Crypto)
	{
		Name: "Kraken",
		Type: "crypto",
		Fetch: func(keys map[string]string) (float64, error) {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get("https://api.kraken.com/0/public/Ticker?pair=PAXGUSD")
			if err != nil { return 0, err }
			defer resp.Body.Close()
			var data struct { Result map[string]struct { C []string `json:"c"` } `json:"result"` }
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return 0, err }
			for _, pair := range data.Result { return strconv.ParseFloat(pair.C[0], 64) }
			return 0, fmt.Errorf("parse error")
		},
	},
	// 6. Investing.com (Scraper)
	{
		Name: "Investing.com",
		Type: "fiat",
		Fetch: func(keys map[string]string) (float64, error) {
			client := &http.Client{Timeout: 8 * time.Second}
			req, _ := http.NewRequest("GET", "https://www.investing.com/currencies/xau-usd", nil)
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AurumBot/1.0)")
			resp, err := client.Do(req)
			if err != nil { return 0, err }
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			html := string(body)
			re := regexp.MustCompile(`([0-9]{1},?[0-9]{3}\.[0-9]{2})`)
			matches := re.FindAllString(html, -1)
			for _, match := range matches {
				cleaned := strings.ReplaceAll(match, ",", "")
				price, err := strconv.ParseFloat(cleaned, 64)
				if err == nil && price > 1500 && price < 5000 { return price, nil }
			}
			return 0, fmt.Errorf("price pattern not found")
		},
	},
}

func fetchAllPrices(keys map[string]string) (float64, int, map[string]interface{}, float64, error) {
	type result struct {
		name  string
		stype string
		price float64
		err   error
		dur   time.Duration
	}
	results := make(chan result, len(sources))
	
	for _, source := range sources {
		go func(s PriceSource) {
			start := time.Now()
			price, err := s.Fetch(keys)
			results <- result{name: s.Name, stype: s.Type, price: price, err: err, dur: time.Since(start)}
		}(source)
	}
	
	var prices []float64
	var cryptoPrices []float64
	var fiatPrices []float64
	successCount := 0
	
	for i := 0; i < len(sources); i++ {
		res := <-results
		if res.err != nil {
			log.Printf("⚠️  [%s] Failed: %v", res.name, res.err)
			continue
		}
		if res.price < 1000 { continue }
		
		log.Printf("✅ [%s] $%.2f (%dms)", res.name, res.price, res.dur.Milliseconds())
		prices = append(prices, res.price)
		
		if res.stype == "crypto" { cryptoPrices = append(cryptoPrices, res.price) }
		if res.stype == "fiat" { fiatPrices = append(fiatPrices, res.price) }
		successCount++
	}
	
	if len(prices) == 0 { return 0, 0, nil, 0, fmt.Errorf("all sources failed") }
	
	sort.Float64s(prices)
	median := prices[len(prices)/2]
	if len(prices)%2 == 0 { median = (prices[len(prices)/2-1] + prices[len(prices)/2]) / 2 }
	
	stats := make(map[string]interface{})
	
	// Fiat Stats
	if len(fiatPrices) > 0 {
		sort.Float64s(fiatPrices)
		fiatMedian := fiatPrices[len(fiatPrices)/2]
		stats["fiat_median"] = fiatMedian
		stats["fiat_sources"] = len(fiatPrices)
	} else {
		stats["fiat_median"] = 0.0
		stats["fiat_sources"] = 0
	}

	// Crypto Stats
	if len(cryptoPrices) > 0 {
		sort.Float64s(cryptoPrices)
		cryptoMedian := cryptoPrices[len(cryptoPrices)/2]
		stats["crypto_median"] = cryptoMedian
		stats["crypto_sources"] = len(cryptoPrices)
	} else {
		stats["crypto_median"] = 0.0
		stats["crypto_sources"] = 0
	}
	
	// Spread Calculation
	spread := 0.0
	fiatVal := stats["fiat_median"].(float64)
	cryptoVal := stats["crypto_median"].(float64)
	
	if fiatVal > 0 && cryptoVal > 0 {
		spread = cryptoVal - fiatVal
	}
	
	// Calibration
	offsetStr := os.Getenv("PRICE_OFFSET")
	if offsetStr != "" {
		offset, err := strconv.ParseFloat(offsetStr, 64)
		if err == nil && offset != 0 { median += offset }
	}
	return median, successCount, stats, spread, nil
}

func priceHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	keys := map[string]string{
		"GOLDAPI_IO_KEY":  os.Getenv("GOLDAPI_IO_KEY"),
		"GOLDAPI_COM_KEY": os.Getenv("GOLDAPI_COM_KEY"),
		"POLYGON_KEY":     os.Getenv("POLYGON_KEY"),
	}
	
	price, sources, stats, spread, err := fetchAllPrices(keys)
	latency := time.Since(start).Milliseconds()
	
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	// --- NEW STRUCTURED RESPONSE ---
	// Prioritizing Fiat (Non-Crypto) as requested
	
	response := map[string]interface{}{
		"fiat": map[string]interface{}{
			"price":   stats["fiat_median"],
			"sources": stats["fiat_sources"],
		},
		"crypto": map[string]interface{}{
			"price":   stats["crypto_median"],
			"sources": stats["crypto_sources"],
		},
		"aggregate": map[string]interface{}{
			"price":   price,
			"total_sources": sources,
		},
		"spread":     spread,
		"latency_ms": latency,
		"timestamp":  time.Now().Unix(),
	}
	
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = DEFAULT_PORT }
	http.HandleFunc("/price", priceHandler)
	http.HandleFunc("/health", healthHandler)
	log.Printf("Aurum Node listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}