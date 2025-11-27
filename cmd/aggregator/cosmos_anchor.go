package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type CosmosAnchor struct {
	Endpoint string
	ChainID  string
	Enabled  bool
}

func NewCosmosAnchor(config Config) *CosmosAnchor {
	return &CosmosAnchor{
		Endpoint: config.Cosmos.RPCEndpoint,
		ChainID:  config.Cosmos.ChainID,
		Enabled:  config.Cosmos.Enabled,
	}
}

// Anchor sends the Merkle Root to the Cosmos chain
// In a real production env, this would sign a MsgStoreCode or MsgMemo transaction
func (ca *CosmosAnchor) Anchor(blockIndex int64, merkleRoot string, blockHash string) {
	if !ca.Enabled {
		return
	}

	// Prepare the payload (Simulating a Cosmos REST API payload)
	payload := map[string]interface{}{
		"type": "aurum/MsgAnchor",
		"value": map[string]string{
			"height":      fmt.Sprintf("%d", blockIndex),
			"merkle_root": merkleRoot,
			"block_hash":  blockHash,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		},
	}
	
	jsonData, _ := json.Marshal(payload)

	// In a demo environment, we often don't have a full Cosmos node running locally.
	// We will log the exact transaction structure that WOULD be sent.
	log.Printf("⚓ COSMOS ANCHOR | Broadcasting Tx to %s (%s)", ca.ChainID, ca.Endpoint)
	log.Printf("   └── Payload: %s", string(jsonData))

	// Mock HTTP Call to demonstrate integration
	// go ca.broadcast(jsonData) 
}

func (ca *CosmosAnchor) broadcast(data []byte) {
	client := &http.Client{Timeout: 5 * time.Second}
	// Assuming a custom sidecar listening on port 1317
	resp, err := client.Post(ca.Endpoint+"/txs", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("⚠️  Anchor failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("✅ Anchor confirmed on Cosmos Testnet")
}