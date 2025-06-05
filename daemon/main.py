#!/usr/bin/env python3
"""
DOGGOWOOF Webhook Daemon 🚨🐕🚨
The listening ear that never sleeps!
"""

import os
import sqlite3
from datetime import datetime
from typing import Dict, Any, Optional

import structlog
from fastapi import FastAPI, Request, HTTPException
from pydantic import BaseModel
import uvicorn
from plyer import notification

# Configure logging
structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.dev.ConsoleRenderer()
    ]
)
logger = structlog.get_logger()

app = FastAPI(title="DOGGOWOOF Webhook Receiver", version="0.1.0")

# Database setup
DB_PATH = os.path.expanduser("~/.doggowoof.db")

class WebhookPayload(BaseModel):
    source: str
    payload: Dict[str, Any]
    timestamp: Optional[datetime] = None

class AlertManager:
    def __init__(self, db_path: str):
        self.db_path = db_path
        self._init_db()
    
    def _init_db(self):
        """Initialize SQLite tables"""
        conn = sqlite3.connect(self.db_path)
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
        conn.commit()
        conn.close()
    
    def store_alert(self, source: str, severity: str, title: str, 
                   message: str = None, raw_payload: str = None):
        """Store alert in database"""
        conn = sqlite3.connect(self.db_path)
        conn.execute(
            "INSERT INTO alerts (source, severity, title, message, raw_payload) VALUES (?, ?, ?, ?, ?)",
            (source, severity, title, message, raw_payload)
        )
        conn.commit()
        conn.close()
        
        # Desktop notification for high priority
        if severity in ["CRITICAL", "HIGH"]:
            self._desktop_notify(title, message or f"Alert from {source}")
    
    def _desktop_notify(self, title: str, message: str):
        """Send desktop notification"""
        try:
            notification.notify(
                title=f"🚨 DOGGOWOOF: {title}",
                message=message,
                timeout=10
            )
            logger.info("Desktop notification sent", title=title)
        except Exception as e:
            logger.error("Failed to send notification", error=str(e))

alert_manager = AlertManager(DB_PATH)

@app.get("/")
async def health_check():
    """Health check endpoint - WOOF!"""
    return {"status": "WOOF! Doggo is listening 🐕", "timestamp": datetime.now()}

@app.post("/webhook/github")
async def github_webhook(request: Request):
    """Handle GitHub webhooks (Actions, Issues, PRs)"""
    payload = await request.json()
    event_type = request.headers.get("X-GitHub-Event", "unknown")
    
    logger.info("GitHub webhook received", event_type=event_type)
    
    # GitHub Actions failure
    if event_type == "workflow_run" and payload.get("action") == "completed":
        if payload["workflow_run"]["conclusion"] == "failure":
            alert_manager.store_alert(
                source="github",
                severity="HIGH",
                title=f"CI Failed: {payload['workflow_run']['name']}",
                message=f"Workflow failed in {payload['repository']['name']}",
                raw_payload=str(payload)
            )
    
    # Store all GitHub events for analysis
    alert_manager.store_alert(
        source="github",
        severity="LOW",
        title=f"GitHub {event_type}",
        raw_payload=str(payload)
    )
    
    return {"status": "received", "event": event_type}

@app.post("/webhook/uroboro")
async def uroboro_webhook(request: Request):
    """Handle Uroboro webhooks"""
    payload = await request.json()
    
    logger.info("Uroboro webhook received", payload_keys=list(payload.keys()))
    
    # Uroboro-specific logic here
    alert_manager.store_alert(
        source="uroboro",
        severity="MEDIUM",
        title="Uroboro Event",
        message=f"Event: {payload.get('event_type', 'unknown')}",
        raw_payload=str(payload)
    )
    
    return {"status": "received", "source": "uroboro"}

@app.post("/webhook/cursor")
async def cursor_webhook(request: Request):
    """Handle Cursor/billing webhooks"""
    payload = await request.json()
    
    logger.info("Cursor webhook received")
    
    # Cursor usage/billing alerts
    if "usage" in payload:
        usage = payload["usage"]
        if usage.get("percentage", 0) > 80:  # 80% usage warning
            alert_manager.store_alert(
                source="cursor",
                severity="HIGH",
                title="Cursor Usage Warning",
                message=f"Usage at {usage.get('percentage', 0)}%",
                raw_payload=str(payload)
            )
    
    return {"status": "received", "source": "cursor"}

@app.post("/webhook/generic")
async def generic_webhook(request: Request):
    """Handle any generic webhook"""
    payload = await request.json()
    source = payload.get("source", "unknown")
    
    logger.info("Generic webhook received", source=source)
    
    alert_manager.store_alert(
        source=source,
        severity=payload.get("severity", "MEDIUM"),
        title=payload.get("title", f"Alert from {source}"),
        message=payload.get("message"),
        raw_payload=str(payload)
    )
    
    return {"status": "received", "source": source}

if __name__ == "__main__":
    logger.info("🚨 DOGGOWOOF Daemon starting up! 🐕")
    uvicorn.run(app, host="127.0.0.1", port=8080, log_level="info") 