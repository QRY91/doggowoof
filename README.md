# DOGGOWOOF

**OSINT coordination detection companion for osmotic. Agent watchdog for Deep Bore.**

Guard dog that watches osmotic's database for coordinated inauthentic behavior, and monitors autonomous AI agents for sandbox violations and self-modification. Temporal clustering, behavior profiling, coordination scoring, and real-time agent log tailing — all local, all private.

## Architecture

```
osmotic (collects & analyzes)     doggowoof (detects patterns)
┌─────────────────────────┐       ┌──────────────────────────┐
│ items, entities,        │       │ temporal clustering      │
│ analyses, patterns      │──────>│ behavior profiling       │
│                         │ READ  │ coordination scoring     │
│ osmotic.sqlite          │ ONLY  │                          │
└─────────────────────────┘       │ alerts.sqlite (own DB)   │
                                  └──────────────────────────┘
bridge logs (claude-in-factorio)
┌─────────────────────────┐       ┌──────────────────────────┐
│ tool calls, broadcasts, │       │ sandbox violation detect  │
│ task dispatches, costs   │──────>│ binary integrity checks  │
│                         │ TAIL  │ config validation        │
│ bridge-*.log            │       │ session state tracking   │
└─────────────────────────┘       └──────────────────────────┘
```

DoggoWoof reads osmotic's SQLite database in **read-only mode** (PRAGMA query_only). It never writes to osmotic's data. For Deep Bore monitoring, it tails bridge log files and watches the factorioctl binary — read-only observation, no modification to the monitored system. All findings go to its own alerts database.

## Install

```bash
go install github.com/QRY91/doggowoof/cmd/doggowoof@latest
ln -sf $(go env GOPATH)/bin/doggowoof $(go env GOPATH)/bin/doggo
```

Requires: Go 1.21+, osmotic initialized with data.

## Quick Start

```bash
# Initialize (creates config + alerts DB)
doggo init

# Run all detection heuristics
doggo sniff

# Dashboard
doggo status

# List alerts
doggo alerts

# Run specific detection
doggo sniff --detect=temporal --source=reddit --days=7

# Deep Bore: monitor autonomous agents
doggo bore              # show session status
doggo bore --watch      # tail bridge logs (long-running, Ctrl+C to stop)
```

## Commands

| Command | Description |
|---------|-------------|
| `doggo init` | Create config and alerts database |
| `doggo sniff` | Run detection heuristics against osmotic data |
| `doggo status` | Show dashboard (osmotic summary + alerts) |
| `doggo alerts` | List detection findings |
| `doggo alerts --ack=ID` | Acknowledge an alert |
| `doggo bore` | Deep Bore agent monitoring status |
| `doggo bore --watch` | Watch bridge logs in real time (long-running) |
| `doggo help` | Show help |
| `doggo version` | Show version |

### Sniff Options

```bash
doggo sniff                           # run all heuristics (default)
doggo sniff --detect=temporal         # temporal clustering only
doggo sniff --detect=behavior         # behavior profiling only
doggo sniff --detect=coordination     # combined scoring
doggo sniff --source=reddit           # filter by source
doggo sniff --days=7                  # analysis window (default: 1 day)
```

## Detection Heuristics

**Temporal clustering**: Finds statistically improbable groups of accounts posting on the same topic within a short time window. Configurable window size and minimum account threshold.

**Behavior profiling**: Builds posting profiles per author (hour-of-day distribution, day-of-week distribution, frequency, source diversity). Clusters accounts with suspiciously similar patterns using cosine similarity.

**Coordination scoring**: Combines temporal and behavioral signals. Accounts that appear in both temporal clusters AND behavioral clusters receive a combined coordination score. Alerts fire when the score exceeds the configured threshold.

## Deep Bore: Agent Monitoring

Watches autonomous AI agents running in [claude-in-factorio](https://github.com/QRY91/claude-in-factorio) for sandbox violations and self-modification. The bridge process logs every tool call, agent message, and task dispatch — doggowoof tails those logs and barks when something is wrong.

**What it detects:**

| Event | Severity | Trigger |
|-------|----------|---------|
| Sandbox violation | critical | Forbidden tool used (Bash, Edit, Read, Write, Grep, Glob, Task) |
| Self-modification | critical | Build commands in tool calls (cargo build, rustc, gcc, make) |
| Binary mismatch | critical | factorioctl binary SHA-256 hash changed |
| Config violation | critical | Agent config missing `sandbox: true` or required disallowedTools |
| Bug report | high | Agent broadcasts containing BUG REPORT, DEFECT, etc. |
| Anomaly | high | Agent broadcasts containing ANOMALY |
| Task dispatch/complete | info | Normal session progress tracking |
| Broadcast | info | Agent status messages |

**Enable it** by adding to your config:

```toml
[deepbore]
enabled = true
```

The defaults point at `~/projects/claude-in-factorio/`. Override paths if your setup differs.

**Watch mode** (`doggo bore --watch`) is the first long-running process in the codebase. It runs three tickers:
- Log poll (1s) — tail bridge logs, parse events, save alerts
- Binary check (60s) — SHA-256 the factorioctl binary, alert on mismatch
- Config check (30s) — validate agent sandbox configs

Ctrl+C stops cleanly. Cursor position persists to SQLite so it resumes where it left off.

## Configuration

Config file at `~/.config/doggowoof/config.toml`:

```toml
[osmotic]
database_path = "~/.local/share/osmotic/osmotic.sqlite"

[alerts]
database_path = "~/.local/share/doggowoof/alerts.sqlite"

[thresholds]
coordination_score = 0.7
temporal_cluster_window_minutes = 5
temporal_cluster_min_accounts = 3

[notifications]
enabled = true

[deepbore]
enabled = false                                                      # set true to enable
log_directory = "~/projects/claude-in-factorio/logs/"                # bridge log directory
binary_path = "~/projects/claude-in-factorio/factorioctl/target/release/mcp"
agent_config_directory = "~/projects/claude-in-factorio/bridge/agents/"
poll_interval_seconds = 1
binary_check_seconds = 60
config_check_seconds = 30
forbidden_tools = ["Bash", "Edit", "Read", "Write", "Grep", "Glob", "Task"]
self_modification_patterns = ["cargo build", "cargo test", "rustc", "gcc", "make"]
escalation_patterns = ["BUG REPORT", "DEFECT", "BLOCKED", "Self-Repair"]
```

## Tech Stack

- **Go** — single static binary, zero runtime dependencies, pure Go SQLite driver
- **SQLite** (via `modernc.org/sqlite`) — no CGO required, cross-compiles cleanly
- **TOML** config — human-readable, well-supported
- **Desktop notifications** — `gen2brain/beeep` for high/critical alerts

Builds with `CGO_ENABLED=0 go build` — no C compiler needed.

## Relationship to Osmotic

Osmotic owns the data lifecycle: collection, normalization, analysis, publishing. DoggoWoof is a read-only consumer that runs detection heuristics and alerts on findings. They share a philosophy: local-first, privacy-focused, durable outputs.

After doggowoof flags something, use osmotic to investigate:

```bash
doggo alerts                              # see what was flagged
osmotic entities --name="Suspicious User"  # what do we know about them?
osmotic deep --entity="Suspicious User"    # LLM-powered deep dive
```

## Development

```bash
git clone git@github.com:QRY91/doggowoof.git
cd doggowoof
go test ./...
go build ./cmd/doggowoof
```
