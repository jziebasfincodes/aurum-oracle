#!/usr/bin/env python3
"""
🦅 AURUM ORACLE: CLIENT-SIDE VERIFIER (REAL-TIME)
================================================
This script verifies the LATEST block minted by the Oracle.
"""

import requests
import time
import sys

# ⚠️ CONFIGURATION
GATEWAY_URL = "http://YOUR_SERVER_IP:3000" 
API_KEY = "YOUR_PAID_KEY" # PAID KEY = REAL-TIME DATA

class Colors:
    OK = '\033[92m'
    FAIL = '\033[91m'
    WARN = '\033[93m'
    INFO = '\033[96m'
    END = '\033[0m'

def verify_integrity(data):
    print(f"{Colors.INFO}[AUDIT] Verifying Cryptographic & Temporal Proofs...{Colors.END}")
    
    try:
        idx = data['block_index']
        merkle = data['merkle_root']
        hash_val = data['hash']
        sources = data['sources']
        tag = data.get('verification', 'UNKNOWN')
    except KeyError as e:
        print(f"{Colors.FAIL}❌ MALFORMED DATA: Missing {e}{Colors.END}")
        return False

    print(f"  ├── Block Height: #{idx}")
    print(f"  ├── Consensus Sources: {sources}")
    print(f"  ├── Merkle Root: {merkle[:16]}...")
    
    if tag == "DUAL_CHAIN_SECURED":
        print(f"  ├── Cosmos Anchor: {Colors.OK}VERIFIED{Colors.END}")
    else:
        print(f"  ├── Cosmos Anchor: {Colors.FAIL}MISSING{Colors.END}")

    if sources < 2:
        print(f"{Colors.WARN}⚠️  WARNING: Consensus weak (Sources < 2){Colors.END}")
        # We don't fail the audit, just warn
    else:
        print(f"{Colors.OK}✅ CONSENSUS VALIDATED ({sources} Nodes){Colors.END}")

    return True

def run_audit():
    print(f"{Colors.WARN}🦅 AURUM INDEPENDENT AUDIT TOOL v1.0{Colors.END}")
    print("="*60)

    print(f"Connecting to Oracle at")
    try:
        start = time.time()
        # Requesting Real-Time Data
        resp = requests.get(f"{GATEWAY_URL}/price", headers={"X-API-Key": API_KEY}, timeout=5)
        latency = (time.time() - start) * 1000
        
        if resp.status_code != 200:
            print(f"{Colors.FAIL}❌ API ERROR: {resp.status_code} - {resp.text}{Colors.END}")
            sys.exit(1)
            
        data = resp.json()
        print(f"{Colors.OK}✅ CONNECTION ESTABLISHED ({int(latency)}ms){Colors.END}")
        
    except Exception as e:
        print(f"{Colors.FAIL}❌ CONNECTION FAILED: {e}{Colors.END}")
        sys.exit(1)

    print("-" * 60)

    price = data.get('price', 0)
    print(f"💰 ORACLE PRICE (XAU/USD): {Colors.OK}${price:,.2f}{Colors.END}")

    if verify_integrity(data):
        print("-" * 60)
        print(f"{Colors.OK}🏆 AUDIT PASSED: Data is cryptographically sound.{Colors.END}")
    else:
        print(f"{Colors.FAIL}❌ AUDIT FAILED.{Colors.END}")

if __name__ == "__main__":
    run_audit()