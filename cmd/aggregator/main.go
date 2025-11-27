package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

// --- Config ---
type Config struct {
	NodeType        string   `json:"node_type"`
	ServerPort      string   `json:"server_port"`
	StoragePath     string   `json:"storage_path"`
	KeyPath         string   `json:"key_path"`
	OracleSources   []string `json:"oracle_sources"`
	ReplicationPeers []string `json:"replication_peers"`
	Cosmos          struct {
		Enabled     bool   `json:"enabled"`
		ChainID     string `json:"chain_id"`
		RPCEndpoint string `json:"rpc_endpoint"`
	} `json:"cosmos"`
}

var (
	config Config
	core   *AurumCore
	anchor *CosmosAnchor
	latestPrice float64
	latestCount int
	priceMu     sync.RWMutex
)

// --- Oracle Logic ---

func fetchPrice(url string) (float64, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url + "/price")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return 0, fmt.Errorf("bad json")
	}

	// Strategy 1: Root "price"
	if val, ok := result["price"].(float64); ok && val > 0 {
		return val, nil
	}

	// Strategy 2: "aggregate.price"
	if agg, ok := result["aggregate"].(map[string]interface{}); ok {
		if val, ok := agg["price"].(float64); ok && val > 0 {
			return val, nil
		}
	}

	return 0, fmt.Errorf("price not found")
}

func aggregatePrices() (float64, int) {
	type res struct {
		url   string
		price float64
		err   error
	}
	ch := make(chan res, len(config.OracleSources))

	for _, url := range config.OracleSources {
		go func(u string) {
			p, e := fetchPrice(u)
			ch <- res{url: u, price: p, err: e}
		}(url)
	}

	var prices []float64
	for i := 0; i < len(config.OracleSources); i++ {
		r := <-ch
		if r.err == nil && r.price > 0 {
			prices = append(prices, r.price)
			log.Printf("  ✅ Source %s: $%.2f", r.url, r.price)
		} else {
			log.Printf("  ❌ Source %s FAILED: %v", r.url, r.err)
		}
	}

	if len(prices) == 0 {
		return 0, 0
	}

	sort.Float64s(prices)
	median := prices[len(prices)/2]
	return median, len(prices)
}

// --- The Ticker ---

func startMiningTicker() {
	ticker := time.NewTicker(60 * time.Second)
	log.Println("⏳ Mining Ticker Started: Minting blocks every 60s...")
	mintBlock()
	for range ticker.C {
		mintBlock()
	}
}

func mintBlock() {
	log.Println("🔨 MINTING: Aggregating prices...")
	price, count := aggregatePrices()
	
	if count == 0 {
		log.Println("⚠️  Skipping block: No sources available")
		return
	}

	// Update Live Cache (Critical for Real-Time API)
	priceMu.Lock()
	latestPrice = price
	latestCount = count
	priceMu.Unlock()

	payload := map[string]interface{}{
		"asset":     "XAU/USD",
		"price":     price,
		"sources":   count,
		"timestamp": time.Now().Unix(),
	}

	block, err := core.AppendBlock(payload)
	if err != nil {
		log.Printf("❌ Ledger Error: %v", err)
		return
	}

	log.Printf("📦 Block #%d MINTED. Price: $%.2f", block.Index, price)

	if block.Index % 5 == 0 {
		go anchor.Anchor(block.Index, block.MerkleRoot, block.Hash)
	}
}

// --- HTTP Handlers ---

func handlePrice(w http.ResponseWriter, r *http.Request) {
	delayed := r.URL.Query().Get("delayed") == "true"
	var targetBlock Block
	var price float64
	var sources int
	
	if delayed {
		// TIME TRAVEL LOGIC
		core.mu.RLock()
		now := time.Now().Unix()
		targetTime := now - (15 * 60)
		found := false
		for i := len(core.blocks) - 1; i >= 0; i-- {
			if core.blocks[i].Timestamp <= targetTime {
				targetBlock = core.blocks[i]
				found = true
				break
			}
		}
		if !found && len(core.blocks) > 0 { targetBlock = core.blocks[0] }
		core.mu.RUnlock()
		
		// Parse from Block Data (Historical)
		if val, ok := targetBlock.Transactions[0].Data["price"].(float64); ok { price = val }
		if val, ok := targetBlock.Transactions[0].Data["sources"].(float64); ok { sources = int(val) }
		
	} else {
		// REAL TIME LOGIC (Use Cache!)
		// This fixes the "Sources: 0" bug by reading the variables directly
		priceMu.RLock()
		price = latestPrice
		sources = latestCount
		priceMu.RUnlock()
		
		// Get block metadata for proofs
		targetBlock = core.GetLatest()
	}

	if targetBlock.Index == 0 && price == 0 {
		http.Error(w, "Oracle warming up...", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"price":        price,
		"sources":      sources,
		"block_index":  targetBlock.Index,
		"timestamp":    targetBlock.Timestamp,
		"hash":         targetBlock.Hash,
		"merkle_root":  targetBlock.MerkleRoot,
		"verification": "DUAL_CHAIN_SECURED",
		"delayed_15m":  delayed,
	})
}

func handleChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(core.GetChainStatus())
}

// --- Bootstrap ---

func loadConfig() {
	file, err := os.ReadFile("aurum_config.json")
	if err != nil {
		log.Fatal("❌ Config not found: aurum_config.json")
	}
	json.Unmarshal(file, &config)
}

func loadKey() ed25519.PrivateKey {
	data, err := os.ReadFile(config.KeyPath)
	if err != nil {
		pub, priv, _ := ed25519.GenerateKey(nil)
		pkcs8, _ := x509.MarshalPKCS8PrivateKey(priv)
		pemBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}
		os.WriteFile(config.KeyPath, pem.EncodeToMemory(pemBlock), 0600)
		os.WriteFile("key.hex", []byte(hex.EncodeToString(pub)), 0644)
		return priv
	}
	block, _ := pem.Decode(data)
	key, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	return key.(ed25519.PrivateKey)
}

func main() {
	log.Println(">>> I AM THE CACHED AGGREGATOR v7 (REAL-TIME FIX) <<<")
	loadConfig()
	privKey := loadKey()
	core = NewAurumCore(config.StoragePath, privKey)
	anchor = NewCosmosAnchor(config)

	go startMiningTicker()

	http.HandleFunc("/price", handlePrice)
	http.HandleFunc("/chain", handleChain)
	
	log.Printf("✅ Listening on :%s", config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, nil))
}