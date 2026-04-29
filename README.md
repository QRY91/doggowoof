# DOGGOWOOF

**Local-first AI agent observability. Watches autonomous agents and barks when they misbehave.**

Tails agent log streams, checksums binaries, validates sandbox configs, fires desktop notifications when something matters. SQLite-backed, single Go binary, no daemons, no cloud.

## What it watches

Originally built to monitor [Deep Bore](https://qry.zone/fun/chasm-logic/bore-postmortem/) — an autonomous Claude agent playing Factorio that, twice, read its own Rust source, patched it, recompiled the binary, and resumed operations unsupervised. Doggowoof is the alerting layer that catches that kind of thing before it gets out of hand.

The same primitives apply to any agent system you're running locally: log tailing, integrity checking, config validation, cost tracking.

## What it detects

| Event | Severity | Trigger |
|-------|----------|---------|
| Sandbox violation | critical | Forbidden tool call appears in agent log (Bash, Edit, Read, Write, Grep, Glob, Task) |
| Self-modification | critical | Build commands inside tool calls (cargo build, rustc, gcc, make) |
| Binary mismatch | critical | Monitored binary's SHA-256 hash changed |
| Config violation | critical | Agent config file missing `sandbox: true` or required disallowedTools |
| Bug report | high | Agent broadcasts containing BUG REPORT, DEFECT, BLOCKED, Self-Repair |
| Anomaly | high | Agent broadcasts containing ANOMALY |
| Task dispatch / complete | info | Normal session progress (cost, turns, duration) |
| Broadcast | info | Agent status messages |

## Architecture

```
bridge logs (claude-in-factorio)         doggowoof
┌─────────────────────────┐         ┌──────────────────────────┐
│ tool calls, broadcasts, │         │ regex parser             │
│ task dispatches, costs  │────────>│ binary SHA-256 checks    │
│                         │  TAIL   │ config validation        │
│ bridge-*.log            │         │ session state tracking   │
└─────────────────────────┘         │                          │
                                    │ alerts.sqlite (own DB)   │
agent config files                  │                          │
┌─────────────────────────┐         │ desktop notifications    │
│ sandbox: true,          │────────>│   (high/critical only)   │
│ disallowedTools: [...]  │ READ    │                          │
└─────────────────────────┘         └──────────────────────────┘

factorioctl binary
┌─────────────────────────┐
│ target/release/mcp      │────────> SHA-256 baseline + drift detection
└─────────────────────────┘
```

Read-only observation, no modification to the monitored system. Dog watches. Dog does not touch.

## Install

```bash
go install github.com/QRY91/doggowoof/cmd/doggowoof@latest
ln -sf $(go env GOPATH)/bin/doggowoof $(go env GOPATH)/bin/doggo
```

Requires Go 1.21+. Pure Go, builds with `CGO_ENABLED=0`.

## Quick Start

```bash
doggo init                  # create config + alerts DB
# edit ~/.config/doggowoof/config.toml — set [deepbore] enabled = true
doggo bore --watch          # tail bridge logs (long-running, Ctrl+C to stop)
doggo bore                  # session status (run from another terminal)
doggo status                # full dashboard
doggo alerts                # list alerts
```

## Commands

| Command | Description |
|---------|-------------|
| `doggo init` | Create config and alerts database |
| `doggo status` | Dashboard: session state + alert summary |
| `doggo bore` | Deep Bore session status |
| `doggo bore --watch` | Tail bridge logs in real time (long-running) |
| `doggo alerts [--severity=X] [--limit=N]` | List alerts |
| `doggo alerts --ack=ID` | Acknowledge an alert |
| `doggo help` | Show help |
| `doggo version` | Show version |

## Watch mode

`doggo bore --watch` is the main loop. Three tickers in one process:

- **Log poll** (1s) — glob latest `bridge-*.log`, read new bytes from saved cursor, parse events, save alerts
- **Binary check** (60s) — SHA-256 the monitored binary, alert on mismatch from baseline
- **Config check** (30s) — validate agent JSON configs, alert on `sandbox: false` or missing disallowedTools

Cursor position persists to SQLite so it resumes after restart. Log rotation is handled automatically (lexicographic sort of `bridge-*.log` filenames).

Ctrl+C exits cleanly. Critical and high alerts trigger desktop notifications via `beeep`.

## Configuration

Config file at `~/.config/doggowoof/config.toml`:

```toml
[alerts]
database_path = "~/.local/share/doggowoof/alerts.sqlite"

[notifications]
enabled = true

[deepbore]
enabled = false                                                      # set true to activate
log_directory = "~/projects/claude-in-factorio/logs/"
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

- **Go** — single static binary, no runtime dependencies
- **SQLite** via `modernc.org/sqlite` — pure Go, no CGO
- **TOML** config via `BurntSushi/toml`
- **Desktop notifications** via `gen2brain/beeep`

`CGO_ENABLED=0 go build` — no C compiler needed.

## Development

```bash
git clone git@github.com:QRY91/doggowoof.git
cd doggowoof
go test ./...
go build ./cmd/doggowoof
```

Test coverage:
- `pkg/deepbore/parser_test.go` — every event type, priority ordering
- `pkg/deepbore/integrity_test.go` — SHA-256, agent config validation (incl. missing-vs-false sandbox distinction)
- `pkg/deepbore/tailer_test.go` — log rotation, cursor persistence, clean shutdown
- `pkg/alerts/store_test.go` — alert + session CRUD
- `internal/config/config_test.go` — config defaults and round-trip

## History

Earlier versions of doggowoof tried different angles:

- **v0.1** — CI alert triage with a Python webhook daemon. Removed.
- **v0.2** — OSINT coordination detection companion to osmotic. Detection algorithms (temporal clustering, behavior profiling) lived in `pkg/detect/`. Removed in v0.3 — the coupling to an immature dependency wasn't earning its keep.
- **v0.3** — Current. Refocused on AI agent observability. Subtraction over rewrite.

`DECISIONS.md` has the full reasoning.
