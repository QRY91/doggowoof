# DOGGOWOOF Python Daemon 🚨🐕🚨

The webhook receiver that never sleeps!

## Quick Start

```bash
# Install dependencies
pip install -r requirements.txt

# Run the daemon
python main.py

# Test it's working
curl http://localhost:8080/
```

## Webhook Endpoints

- `GET /` - Health check (WOOF!)
- `POST /webhook/github` - GitHub webhooks (Actions, Issues, PRs)
- `POST /webhook/uroboro` - Uroboro events
- `POST /webhook/cursor` - Cursor usage/billing
- `POST /webhook/generic` - Any custom webhook

## Database

SQLite at `~/.doggowoof.db` with alerts table:
- Auto-created on first run
- Shared with Go CLI
- Desktop notifications for HIGH/CRITICAL

## Next Steps

1. Set up GitHub webhook pointing to `/webhook/github`
2. Configure Uroboro to send events to `/webhook/uroboro`
3. Build Cursor usage monitoring
4. Train the AI on alert patterns! 