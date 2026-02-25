# DOGGOWOOF ARCHITECTURAL DECISIONS

*A log of key technical decisions and their reasoning*

---

## Decision #001: Webhook Daemon Language Choice
**Status:** SUPERSEDED by #006

### Context
Need to build webhook receiver for monitoring GitHub, Uroboro, Cursor usage, etc. Two options:
1. Build everything in Go for consistency
2. Build webhook daemon in Python first, port later

### Decision
**Python first, port to Go after core logic is stable**

### What we actually built
Simple Python daemon using only stdlib - http.server, sqlite3, subprocess for notifications. Zero external dependencies.

---

## Decision #002: Pet Command Implementation
**Status:** PRESERVED

### Decision
Hidden `doggo pet` command with random responses

### Result
15 wholesome responses + 10% secret message. Ported from Cobra to standalone function. All responses preserved verbatim.

---

## Decision #003: Northstar and Product Identity
**Status:** SUPERSEDED by #007

### Original Decision
"Stop alert fatigue from breaking solodevs" - Local-first intelligent alert triage

### Updated
Pivoted to OSINT coordination detection companion for osmotic.

---

## Decision #004: MVP Success Criteria
**Status:** ACHIEVED (v0.1.0), SUPERSEDED

### Original Criteria
GitHub CI failure detection with local learning storage — achieved.

---

## Decision #005: Go Project Structure & Binary Management
**Status:** PRESERVED

### Decision
Proper Go layout: `cmd/doggowoof/main.go`, `internal/cli/`, `pkg/`. Symlink `doggo` -> `doggowoof`.

---

## Decision #006: Pivot from CI Alert Triage to OSINT Detection
**Status:** DECIDED — February 2026

### Context
DoggoWoof v0.1.0 was a CI alert triage tool with Python webhook daemon. Osmotic needed a companion for detecting coordinated inauthentic behavior in collected OSINT data.

### Decision
**Full rewrite as osmotic's detection companion**

### Reasoning
- Osmotic collects and analyzes public content, but has no coordination detection
- DoggoWoof's guard dog metaphor fits perfectly for "watching" data for threats
- Read-only consumer pattern keeps concerns cleanly separated
- Reuses osmotic's patterns (config, store, CLI) for consistency
- Python daemon was unused; webhook-based approach wasn't getting traction

### What changed
- Removed: Python daemon, Cobra CLI, webhook receiver, GitHub CI integration
- Added: osmotic DB reader, temporal clustering, behavior profiling, coordination scoring
- Preserved: pet command, beeep notifications, guard dog identity, Go project structure

---

## Decision #007: Mission Pivot
**Status:** DECIDED

### Decision
From "stop alert fatigue from breaking solodevs" to "guard dog for OSINT data integrity — detect coordinated inauthentic behavior"

### Reasoning
Better alignment with osmotic ecosystem. Clearer value proposition. Unique capability (no other local-first coordination detection tool exists).

---

## Decision #008: Raw os.Args over Cobra
**Status:** DECIDED

### Decision
Replace Cobra with raw `os.Args` switch dispatcher

### Reasoning
DoggoWoof has 5 commands. Cobra brought 14 transitive dependencies for command routing we don't need. os.Args matches osmotic's pattern and is simpler to maintain.

### What we gained
Removed: spf13/cobra, spf13/viper, spf13/pflag, spf13/afero, spf13/cast, fsnotify, hcl, mapstructure, locafero, slog-shim, conc, gotenv, ini.v1, yaml.v3, go-sqlite3 (CGO)

---

## Decision #009: Pure Go SQLite (modernc.org/sqlite) over CGO (mattn/go-sqlite3)
**Status:** DECIDED

### Decision
Switch from `mattn/go-sqlite3` (requires CGO) to `modernc.org/sqlite` (pure Go)

### Reasoning
Matches osmotic's choice. Enables `CGO_ENABLED=0` builds. Cross-compiles cleanly. No C compiler needed. Aligns with permacomputing principle of minimal dependencies.

---

## Decision #010: TOML Config over Viper/YAML
**Status:** DECIDED

### Decision
Single TOML config file via `BurntSushi/toml` instead of Viper with YAML

### Reasoning
Matches osmotic's config approach. TOML is human-readable and well-supported. One dependency instead of Viper's dependency tree. Config file at `~/.config/doggowoof/config.toml`.

---

## Decision #011: Deep Bore as Go-native Log Tailer
**Status:** DECIDED — February 2026

### Context
PLAN_DEEPBORE_MONITORING.md originally specified a Python webhook daemon to tail bridge logs and POST events to doggowoof. By the time implementation started, all Python/daemon/webhook code had been removed. The project was pure Go CLI.

### Decision
**Go-native log tailer with direct SQLite writes, no webhooks**

### Reasoning
- Python daemon and webhook handlers no longer exist — nothing to hook into
- Direct SQLite writes eliminate the HTTP middleman
- Keeps the project as a single Go binary
- `doggo bore --watch` is the first long-running process in the codebase — simpler architecture means fewer failure modes
- Ticker-based poll (1s) over fsnotify — zero dependencies, handles log rotation naturally

### What was built
- `pkg/deepbore/parser.go` — regex-based log event parser (10 event types, priority ordering)
- `pkg/deepbore/integrity.go` — SHA-256 binary checksumming, agent config validation
- `pkg/deepbore/tailer.go` — watch loop with three tickers (poll, binary, config)
- `internal/cli/bore.go` — `doggo bore` status display and `--watch` mode
- `deepbore_sessions` table in alerts.sqlite for mutable session state
- `[deepbore]` config section with all monitoring parameters

---

*Keep this updated as we make architectural choices. Future you will thank present you.*
