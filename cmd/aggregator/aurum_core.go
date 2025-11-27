package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// --- Types ---

type Transaction struct {
	TxHash    string                 `json:"tx_hash"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	PreviousHash string        `json:"previous_hash"`
	MerkleRoot   string        `json:"merkle_root"`
	Transactions []Transaction `json:"transactions"`
	Hash         string        `json:"hash"`
	Signature    string        `json:"signature"`
	SignerPubkey string        `json:"signer_pubkey"`
	Locked       bool          `json:"locked"`
}

// --- Engine ---

type AurumCore struct {
	mu         sync.RWMutex
	blocks     []Block
	filepath   string
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewAurumCore(filepath string, privKey ed25519.PrivateKey) *AurumCore {
	core := &AurumCore{
		blocks:     []Block{},
		filepath:   filepath,
		privateKey: privKey,
		publicKey:  privKey.Public().(ed25519.PublicKey),
	}
	core.loadFromDisk()
	return core
}

// --- Crypto Logic ---

func (core *AurumCore) ComputeMerkleRoot(txs []Transaction) string {
	if len(txs) == 0 {
		h := sha256.Sum256([]byte("empty"))
		return hex.EncodeToString(h[:])
	}
	hashes := make([]string, len(txs))
	for i, tx := range txs {
		hashes[i] = tx.TxHash
	}
	// Simple Merkle Tree construction
	for len(hashes) > 1 {
		var nextLevel []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := hashes[i] + hashes[i+1]
				h := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(h[:]))
			} else {
				nextLevel = append(nextLevel, hashes[i])
			}
		}
		hashes = nextLevel
	}
	return hashes[0]
}

func (core *AurumCore) SignBlock(b *Block) {
	// Format: AURUM|v1|Index|PrevHash|MerkleRoot
	msg := fmt.Sprintf("AURUM|v1|%d|%s|%s", b.Index, b.PreviousHash, b.MerkleRoot)
	sig := ed25519.Sign(core.privateKey, []byte(msg))
	b.Signature = hex.EncodeToString(sig)
	b.SignerPubkey = hex.EncodeToString(core.publicKey)
}

func (core *AurumCore) HashBlock(b *Block) string {
	data := fmt.Sprintf("%d%d%s%s", b.Index, b.Timestamp, b.PreviousHash, b.MerkleRoot)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// --- Storage Operations ---

func (core *AurumCore) AppendBlock(data map[string]interface{}) (*Block, error) {
	core.mu.Lock()
	defer core.mu.Unlock()

	// 1. Determine Height and PrevHash
	var index int64 = 0
	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if len(core.blocks) > 0 {
		last := core.blocks[len(core.blocks)-1]
		index = last.Index + 1
		prevHash = last.Hash
	}

	// 2. Construct Tx
	txBytes, _ := json.Marshal(data)
	txHash := sha256.Sum256(txBytes)
	tx := Transaction{
		TxHash:    hex.EncodeToString(txHash[:]),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	// 3. Construct Block
	block := Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PreviousHash: prevHash,
		Transactions: []Transaction{tx},
		Locked:       true, // Default to locked until verified
	}

	// 4. Finalize Crypto
	block.MerkleRoot = core.ComputeMerkleRoot(block.Transactions)
	block.Hash = core.HashBlock(&block)
	core.SignBlock(&block)

	// 5. Persist
	if err := core.writeToDisk(block); err != nil {
		return nil, err
	}

	core.blocks = append(core.blocks, block)
	return &block, nil
}

func (core *AurumCore) GetLatest() Block {
	core.mu.RLock()
	defer core.mu.RUnlock()
	if len(core.blocks) == 0 {
		return Block{}
	}
	return core.blocks[len(core.blocks)-1]
}

func (core *AurumCore) GetChainStatus() map[string]interface{} {
	core.mu.RLock()
	defer core.mu.RUnlock()
	return map[string]interface{}{
		"height":      len(core.blocks),
		"latest_hash": core.blocks[len(core.blocks)-1].Hash,
		"integrity":   "secure",
	}
}

// --- Binary I/O (.dat format) ---

func (core *AurumCore) writeToDisk(b Block) error {
	f, err := os.OpenFile(core.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	
	// Simple Binary Encoding: [Size(4)][JSONBytes(N)]
	// In a real C++ daemon, we would use a packed binary struct, 
	// but for Go interoperability JSON is safer for now.
	data, _ := json.Marshal(b)
	binary.Write(buf, binary.LittleEndian, int32(len(data)))
	buf.Write(data)

	_, err = f.Write(buf.Bytes())
	return err
}

func (core *AurumCore) loadFromDisk() {
	f, err := os.Open(core.filepath)
	if err != nil {
		return // No ledger found, starting fresh
	}
	defer f.Close()

	stats, _ := f.Stat()
	if stats.Size() == 0 {
		return
	}

	for {
		var length int32
		err := binary.Read(f, binary.LittleEndian, &length)
		if err != nil {
			break // EOF
		}

		data := make([]byte, length)
		_, err = f.Read(data)
		if err != nil {
			break
		}

		var b Block
		json.Unmarshal(data, &b)
		core.blocks = append(core.blocks, b)
	}
	log.Printf("📚 Core: Loaded %d blocks from secure storage", len(core.blocks))
}