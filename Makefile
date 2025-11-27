# Makefile for AURUM Oracle

.PHONY: all aggregator node gateway clean

all: aggregator node gateway

aggregator:
	@echo "Building Aggregator (Leader)..."
	@mkdir -p bin
	go build -o bin/aurum-aggregator ./cmd/aggregator/main.go ./cmd/aggregator/aurum_core.go ./cmd/aggregator/cosmos_anchor.go 
	@cp cmd/aggregator/aurum_config.json bin/

node:
	@echo "Building Oracle Node (Worker)..."
	@mkdir -p bin
	go build -o bin/aurum-node ./cmd/oracle_node/main.go

gateway:
	@echo "Building API Gateway..."
	@mkdir -p bin
	go build -o bin/aurum-gateway ./cmd/gateway/main.go

clean:
	rm -rf bin/
	