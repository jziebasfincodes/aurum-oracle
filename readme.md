# AURUM Oracle 🦅

**The World's First Multi-Cloud, Dual-Chain Gold Oracle.**

AURUM utilizes a proprietary Active Defense mechanism to secure Real-World Asset (RWA) data against 51% attacks and data corruption. We aggregate price data across AWS, Azure, and Google Cloud, anchoring every update to the Cosmos Blockchain for thermodynamic certainty.

---

## 🏗 Architecture

AURUM moves beyond the single-server oracle model. We shard consensus across physically separated cloud providers to eliminate single points of failure.

```
graph TD
    subgraph "Layer 1: The Truth Sources (Workers)"
        A[AWS Node] -->|Price Data| D[Aggregator]
        B[Azure Node] -->|Price Data| D
        C[GCP Node] -->|Price Data| D
    end

    subgraph "Layer 2: The Brain (GCP Leader)"
        D[Aggregator] -- Mint Block --> E[(AurumDB Ledger)]
        D -- Sign & Hash --> E
        E -- Anchor Hash --> F[Cosmos Blockchain]
    end

    subgraph "Layer 3: The Gatekeeper"
        G[API Gateway] -->|Auth & Rate Limit| D
        H[Sentinel Watcher] -.->|Monitors| D
        H -.->|Heals| E
    end

    subgraph "Layer 4: The World"
        User[Client App] -->|Request Price| G
        Verifier[Audit Script] -->|Verify Proof| G
    end

    style D fill:#f9f,stroke:#333,stroke-width:4px
    style E fill:#ff9,stroke:#333,stroke-width:2px
    style F fill:#9cf,stroke:#333,stroke-width:2px
```

---

## 🚀 Core Features

- **Multi-Cloud Consensus**: Nodes run on physically separated infrastructure to eliminate single points of failure.
- **Dual-Chain Verification**: High-frequency internal ledger for sub-second latency, anchored to Cosmos for public finality.
- **Active Defense**: Automated sentinels detect tampering and self-heal corrupted binaries in real-time.
- **Tiered Access**: Free historical data for backtesting; Paid real-time streams for institutional trading.

---

## ⚡ Quick Start

### Prerequisites

- Go 1.21+
- Linux Server (Ubuntu 20.04+)

### Installation

**1. Clone the repo**

```bash
git clone https://github.com/jziebasfincodes/aurum-oracle.git
cd aurum-oracle
```

**2. Build Binaries**

```bash
# Builds Node, Aggregator, and Gateway
go build -o bin/aurum-node cmd/oracle_node/main.go
go build -o bin/aurum-aggregator cmd/aggregator/main.go cmd/aggregator/aurum_core.go cmd/aggregator/cosmos_anchor.go
go build -o bin/aurum-gateway cmd/gateway/main.go
```

**3. Run Verification**

```bash
python3 aurum_verifier.py
```

---

## 🛡️ Security Demo

to simulate an attack or hostile takeover, run the included penetration test suite:

```bash
./scripts/demo_attack.sh <TARGET_IP>
```

This script will:
1. Verify Consensus.
2. Attempt to inject malicious blocks (Blocked by Kernel).
3. Attempt to delete binaries (Restored by Sentinel).

---

## 🔮 The Vision (Roadmap)

This repository represents **AURUM V1**. We are currently raising funds to develop:

- **V2 (Privacy)**: Zero-Knowledge Proofs for private price validation. As well as multi-sopurce aggregation. As well as better data speeds and quality. 
- **V3 (The Living Chain)**: A proprietary 51% Immunity Protocol that uses elastic cloud scaling to physically dilute hostile voting power in real-time.

We are seeking **Validators and Partners**. Contact Us to join the testnet.

---

## 📜 License

**PolyForm Noncommercial License 1.0.0**

- ✅ You **MAY** use the **DATA** from this oracle for any purpose (including commercial trading).
- ❌ You **MAY NOT** resell, fork, or host this **CODE** as a competing commercial service.

For commercial licensing inquiries, contact: **jeremyzieba@pm.me**