#!/usr/bin/env python3
"""
DOGGOWOOF Simple Daemon 🚨🐕🚨
A minimal webhook receiver using only Python standard library
"""

import json
import sqlite3
import subprocess
import sys
from datetime import datetime
from http.server import HTTPServer, BaseHTTPRequestHandler
from os.path import expanduser
from urllib.parse import urlparse

class DoggoHandler(BaseHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        self.db_path = expanduser("~/.doggowoof/alerts.db")
        super().__init__(*args, **kwargs)
    
    def do_GET(self):
        """Health check endpoint"""
        if self.path == "/":
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {
                "status": "WOOF! Doggo is listening 🐕",
                "timestamp": datetime.now().isoformat()
            }
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def do_POST(self):
        """Handle webhook posts"""
        try:
            # Parse content
            content_length = int(self.headers.get('content-length', 0))
            post_data = self.rfile.read(content_length)
            payload = json.loads(post_data.decode())
            
            # Determine source from path
            source = "unknown"
            if "/webhook/github" in self.path:
                source = "github"
                self._handle_github_webhook(payload)
            elif "/webhook/uroboro" in self.path:
                source = "uroboro"
                self._handle_uroboro_webhook(payload)
            elif "/webhook/generic" in self.path:
                source = payload.get("source", "generic")
                self._handle_generic_webhook(payload)
            
            print(f"🐕 WOOF! Received {source} webhook: {datetime.now()}")
            
            # Respond with success
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {"status": "received", "source": source}
            self.wfile.write(json.dumps(response).encode())
            
        except Exception as e:
            print(f"❌ WOOF! Error handling webhook: {e}")
            self.send_response(500)
            self.end_headers()
    
    def _handle_github_webhook(self, payload):
        """Handle GitHub webhooks - focus on CI failures"""
        event_type = self.headers.get("X-GitHub-Event", "unknown")
        
        # GitHub Actions failure detection
        if event_type == "workflow_run" and payload.get("action") == "completed":
            if payload["workflow_run"]["conclusion"] == "failure":
                self._store_alert(
                    source="github",
                    severity="HIGH",
                    title=f"🚨 CI FAILED: {payload['workflow_run']['name']}",
                    message=f"Workflow failed in {payload['repository']['name']}",
                    raw_payload=json.dumps(payload)
                )
                return
        
        # Store other GitHub events as low priority
        self._store_alert(
            source="github",
            severity="LOW",
            title=f"GitHub {event_type}",
            message=f"Event in {payload.get('repository', {}).get('name', 'unknown')}",
            raw_payload=json.dumps(payload)
        )
    
    def _handle_uroboro_webhook(self, payload):
        """Handle Uroboro webhooks"""
        self._store_alert(
            source="uroboro", 
            severity="MEDIUM",
            title="Uroboro Activity",
            message=f"Event: {payload.get('event_type', 'unknown')}",
            raw_payload=json.dumps(payload)
        )
    
    def _handle_generic_webhook(self, payload):
        """Handle generic webhooks"""
        self._store_alert(
            source=payload.get("source", "generic"),
            severity=payload.get("severity", "MEDIUM"),
            title=payload.get("title", "Generic Alert"),
            message=payload.get("message", ""),
            raw_payload=json.dumps(payload)
        )
    
    def _store_alert(self, source, severity, title, message="", raw_payload=""):
        """Store alert in SQLite database"""
        try:
            conn = sqlite3.connect(self.db_path)
            
            # Ensure table exists
            conn.execute("""
                CREATE TABLE IF NOT EXISTS alerts (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    source TEXT NOT NULL,
                    severity TEXT NOT NULL,
                    title TEXT NOT NULL,
                    message TEXT,
                    raw_payload TEXT,
                    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                    acknowledged BOOLEAN DEFAULT FALSE
                )
            """)
            
            conn.execute(
                "INSERT INTO alerts (source, severity, title, message, raw_payload) VALUES (?, ?, ?, ?, ?)",
                (source, severity, title, message, raw_payload)
            )
            conn.commit()
            conn.close()
            
            print(f"📊 STORED: {severity} alert from {source}: {title}")
            
            # Desktop notification for high priority
            if severity in ["CRITICAL", "HIGH"]:
                self._desktop_notify(title, message)
                
        except Exception as e:
            print(f"❌ Database error: {e}")
    
    def _desktop_notify(self, title, message):
        """Send desktop notification using system tools"""
        try:
            # Try notify-send on Linux
            full_title = f"🚨 DOGGOWOOF: {title}"
            subprocess.run([
                "notify-send", 
                "--urgency=critical",
                "--app-name=DOGGOWOOF", 
                full_title, 
                message or "Alert detected!"
            ], check=False)
            print(f"🔔 Desktop notification sent: {title}")
        except:
            # Fallback - just print to console
            print(f"🔔 ALERT: {title} - {message}")

def main():
    print("🚨🐕 DOGGOWOOF Simple Daemon starting up! 🐕🚨")
    print("📡 Listening on http://localhost:8080")
    print("🐾 Ready to catch webhooks and BARK at important alerts!")
    
    server = HTTPServer(('localhost', 8080), DoggoHandler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n🛑 DOGGOWOOF daemon shutting down. Good boy, rest well! 🐕")
        server.server_close()

if __name__ == "__main__":
    main() 