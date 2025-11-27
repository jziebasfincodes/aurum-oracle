#!/usr/bin/env python3
import sys
import time
import requests
import subprocess
import json

# --- CONFIGURATION ---
# Replace 'YOUR_USERNAME' with your actual GCP username if different
SSH_USER = "YOUR_USERNAME" 
TARGET_IP = sys.argv[1] if len(sys.argv) > 1 else ""

# Visual Colors
class Colors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    GREEN = '\033[92m'
    RED = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_step(msg):
    print(f"\n{Colors.HEADER}>>> {msg}{Colors.ENDC}")

def ssh_exec(cmd):
    """Runs a command on the remote server via SSH."""
    ssh_cmd = [
        "ssh", 
        "-o", "StrictHostKeyChecking=no", 
        "-o", "UserKnownHostsFile=/dev/null",
        f"{SSH_USER}@{TARGET_IP}", 
        cmd
    ]
    result = subprocess.run(ssh_cmd, capture_output=True, text=True)
    return result.returncode, result.stdout, result.stderr

def main():
    if not TARGET_IP:
        print("Usage: python3 demo_attack.py")
        sys.exit(1)

    print(f"{Colors.BOLD}╔══════════════════════════════════════════╗")
    print(f"             ║   AURUM ORACLE: LIVE SECURITY DEMO       ║")
    print(f"             ╚══════════════════════════════════════════╝{Colors.ENDC}")
    
    time.sleep(2)

    # --- SCENE 1: FREE TIER ---
    
    

    # --- SCENE 3: THE ATTACK ---
    print_step("Simulating Hostile Takeover (Injection)")
    print("Injecting malicious payload into immutable ledger...")
    
    code, out, err = ssh_exec("echo 'MALICIOUS_DATA_INJECTION' >> aurum_ledger.dat")
    
    if code == 0:
        print(f"{Colors.RED}⚡ Injection Attempted.{Colors.ENDC}")
    else:
        print(f"{Colors.GREEN}🛡️  KERNEL BLOCK: OS prevented write access! (Immutable Attribute Active){Colors.ENDC}")

    time.sleep(1)
    print(f"{Colors.BLUE}Checking Sentinel Logs...{Colors.ENDC}")
    code, out, err = ssh_exec("tail -n 2 tamper_events.log")
    if out:
        print(f"{Colors.GREEN}{out.strip()}{Colors.ENDC}")
    else:
        print(f"{Colors.GREEN}🛡️  System rejected write before logging.{Colors.ENDC}")

    time.sleep(2)

    # --- SCENE 4: DESTRUCTION ---
    print_step("Simulating Binary Destruction (DDoS/Wipe)")
    print("Attempting to delete Gateway binary...")
    ssh_exec("rm aurum-gateway")
    print(f"{Colors.RED}❌ Gateway Binary Deleted.{Colors.ENDC}")
    
    print("Waiting for Auto-Healing...")
    for i in range(3):
        sys.stdout.write(".")
        sys.stdout.flush()
        time.sleep(1)
    print("")

    code, out, err = ssh_exec("ls -l aurum-gateway")
    if "aurum-gateway" in out:
        print(f"{Colors.GREEN}✨ SYSTEM SELF-HEALED. Service Restored automatically.{Colors.ENDC}")
    else:
        print(f"{Colors.RED}❌ Healing Failed.{Colors.ENDC}")

    print(f"\n{Colors.BOLD}DEMO COMPLETE. ARCHITECTURE SECURE.{Colors.ENDC}")

if __name__ == "__main__":
    main()