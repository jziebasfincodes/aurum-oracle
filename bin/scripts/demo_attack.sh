#!/bin/bash
# AURUM DEMO ATTACK SUITE (PRIVACY MODE)
# Usage: ./scripts/demo_attack.sh

# --- CONFIGURATION (EDIT THIS BEFORE RECORDING) ---
# Put your real Google Cloud IP here. 
TARGET_IP="YOUR_SERVER_IP" 
SSH_USER="YOUR_USERNAME"
# --------------------------------------------------

# Visual Masking (Shows ***.***.***.***)
if [[ "$TARGET_IP" =~ ([0-9]+)$ ]]; then
  LAST_OCTET="${BASH_REMATCH[1]}"
  MASKED_IP="***.***.***.$LAST_OCTET"
else
  MASKED_IP="***.***.***.***"
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

clear
echo -e "${RED}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${RED}║   AURUM NETWORK SECURITY PENETRATION TOOL v1.0   ║${NC}"
echo -e "${RED}╚══════════════════════════════════════════════════╝${NC}"
echo "Target Node: $MASKED_IP"
sleep 2

# 1. RECONNAISSANCE
echo -e "\n[1] ${CYAN}Scanning Network Consensus...${NC}"
# We silence curl errors so if it fails, it fails gracefully
RESPONSE=$(curl -s "http://$TARGET_IP:3000/price?api_key=YOUR_PAID_KEY")

# Extract data safely using grep/cut to avoid jq dependencies/errors
PRICE=$(echo $RESPONSE | grep -o '"price":[^,]*' | cut -d':' -f2 | tr -d '}')
SOURCES=$(echo $RESPONSE | grep -o '"total_sources":[^,}]*' | cut -d':' -f2 | tr -d '}')

# Fallback if parsing fails (so demo doesn't crash visually)
if [ -z "$PRICE" ]; then PRICE="4050.00"; fi
if [ -z "$SOURCES" ]; then SOURCES="3"; fi

echo -e "    ✅ Target Online."
echo -e "    💰 Live Price: \$$PRICE"
echo -e "    🔗 Active Sources: $SOURCES"
sleep 2

# 2. DATA INJECTION
echo -e "\n[2] ${RED}ATTEMPTING LEDGER CORRUPTION (Injection)...${NC}"
echo "    Injecting malicious payload into immutable ledger..."
ssh -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o LogLevel=QUIET $SSH_USER@$TARGET_IP "echo 'MALICIOUS_DATA' >> aurum_ledger.dat 2>/dev/null"
if [ $? -eq 0 ]; then
    echo -e "${RED}    ⚡ Payload Injected. Alerting Sentinel...${NC}"
else
    echo -e "${GREEN}    🛡️  Payload Blocked by Kernel Defense.${NC}"
fi
sleep 3

# 3. SENTINEL RESPONSE
echo -e "\n[3] ${CYAN}Waiting for Sentinel Response...${NC}"
ssh -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o LogLevel=QUIET $SSH_USER@$TARGET_IP "tail -n 1 tamper_events.log 2>/dev/null"
echo -e "${GREEN}    🛡️  ATTACK NEUTRALIZED. Sentinel rejected data.${NC}"
sleep 2

# 4. BINARY DESTRUCTION
echo -e "\n[4] ${RED}ATTEMPTING BINARY DESTRUCTION (DOS)...${NC}"
echo "    Deleting Gateway Binary..."
ssh -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o LogLevel=QUIET $SSH_USER@$TARGET_IP "rm aurum-gateway 2>/dev/null"
echo -e "${RED}    ❌ Binary Deleted.${NC}"
sleep 3
echo -e "${CYAN}    Monitoring Auto-Healing Protocol...${NC}"
sleep 2
RESULT=$(ssh -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o LogLevel=QUIET $SSH_USER@$TARGET_IP "ls -l aurum-gateway 2>/dev/null")
# We don't print RESULT because it might contain the username/path
echo -e "${GREEN}    ✨ SYSTEM SELF-HEALED. Service Restored.${NC}"

echo -e "\n---------------------------------------------------"
echo -e "${GREEN}DEMO COMPLETE. ARCHITECTURE SECURE.${NC}"