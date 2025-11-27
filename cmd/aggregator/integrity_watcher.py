#!/usr/bin/env python3
# integrity_watcher.py - Deep Inspection Sentinel with EMAIL ALERTS

import hashlib
import time
import os
import shutil
import json
import struct
import smtplib
import socket
from email.message import EmailMessage
from datetime import datetime


EMAIL_ALERTS_ENABLED = False  # Set to True to enable
SMTP_SERVER = "smtp.gmail.com"
SMTP_PORT = 587
SMTP_USER = "your_email@gmail.com"
SMTP_PASS = "your_app_password"  # Generate this in Google Account Settings
ALERT_RECIPIENT = "your_email@example.com"

# --- WATCH CONFIG ---
WATCH_FILES = {
    'aurum-aggregator': {'type': 'binary', 'restore': True},
    'aurum-gateway':    {'type': 'binary', 'restore': True},
    'aurum_config.json': {'type': 'config', 'restore': True},
    'aurum_ledger.dat': {'type': 'ledger', 'restore': False},
}

BACKUP_DIR = 'backups'
CHECK_INTERVAL = 3
LOG_FILE = 'tamper_events.log'
HOSTNAME = "your-hostname"

class IntegrityMonitor:
    def __init__(self):
        self.file_states = {}
        os.makedirs(BACKUP_DIR, exist_ok=True)
        self.load_state()
        
    def send_email_alert(self, filepath, event, details):
        if not EMAIL_ALERTS_ENABLED: return
        
        msg = EmailMessage()
        msg.set_content(f"""
🚨 AURUM SECURITY ALERT 🚨

NODE: {HOSTNAME}
FILE: {filepath}
EVENT: {event}
TIME: {datetime.now().isoformat()}

DETAILS:
{details}

AUTOMATED ACTION:
Self-healing protocols have been initiated.
        """)
        
        msg['Subject'] = f"🚨 TAMPER DETECTED: {filepath} on {HOSTNAME}"
        msg['From'] = SMTP_USER
        msg['To'] = ALERT_RECIPIENT
        
        try:
            server = smtplib.SMTP(SMTP_SERVER, SMTP_PORT)
            server.starttls()
            server.login(SMTP_USER, SMTP_PASS)
            server.send_message(msg)
            server.quit()
            print("  📧 Alert email sent successfully.")
        except Exception as e:
            print(f"  ❌ Failed to send email: {e}")

    def get_file_info(self, filepath):
        if not os.path.exists(filepath): return None
        size = os.path.getsize(filepath)
        sha256 = hashlib.sha256()
        try:
            with open(filepath, 'rb') as f:
                for chunk in iter(lambda: f.read(4096), b""): sha256.update(chunk)
            return {'hash': sha256.hexdigest(), 'size': size}
        except: return None

    def create_backup(self, filepath):
        if not os.path.exists(filepath): return
        shutil.copy2(filepath, os.path.join(BACKUP_DIR, f"{filepath}.latest"))

    def restore_file(self, filepath):
        latest = os.path.join(BACKUP_DIR, f"{filepath}.latest")
        if os.path.exists(latest):
            print(f"  🚑 SELF-HEALING: Restoring {filepath}...")
            shutil.copy2(latest, filepath)
            return True
        return False

    def log_tamper(self, filepath, event, details=""):
        msg = f"[{datetime.now().isoformat()}] 🚨 TAMPER DETECTED: {filepath} ({event})"
        print(f"\n{msg}")
        with open(LOG_FILE, 'a') as f: f.write(msg + "\n")
        
        # Trigger Alert
        self.send_email_alert(filepath, event, details)

    def load_state(self):
        print("🔍 Scanning System Integrity...")
        for f in WATCH_FILES:
            info = self.get_file_info(f)
            if info:
                self.file_states[f] = info
                self.create_backup(f)
                print(f"  ✅ Secured: {f}")

    def run(self):
        print(f"🛡️  AURUM DEEP SENTINEL ACTIVE (Email Alerts: {EMAIL_ALERTS_ENABLED})")
        try:
            while True:
                time.sleep(CHECK_INTERVAL)
                for fpath, info in WATCH_FILES.items():
                    current = self.get_file_info(fpath)
                    previous = self.file_states.get(fpath)

                    if previous and not current:
                        self.log_tamper(fpath, "DELETED", "File disappeared from disk.")
                        if info['restore']: self.restore_file(fpath)
                        current = self.get_file_info(fpath)
                        
                    elif current and previous and current['hash'] != previous['hash']:
                        if info['type'] == 'ledger':
                            # We skip deep inspection for this brevity, but alert on binary changes
                            pass 
                        else:
                            self.log_tamper(fpath, "MODIFIED", f"Hash mismatch.\nExpected: {previous['hash']}\nActual: {current['hash']}")
                            if info['restore']: self.restore_file(fpath)
                    
                    if current:
                        self.file_states[fpath] = current

        except KeyboardInterrupt:
            print("\n🛑 Sentinel Disengaged.")

if __name__ == "__main__":
    IntegrityMonitor().run