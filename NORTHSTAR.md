# DOGGOWOOF North Star

**Last updated**: April 2026

## The One Problem We Solve

**Local-first observability for autonomous AI agents.**

Watch agent log streams, binaries, and configs. Detect sandbox violations, self-modification, integrity drift, and config regressions. Alert loud and local when something matters; stay quiet otherwise.

## Focus Guardrails

### Always Yes
- Read agent logs and detect safety violations
- Checksum monitored binaries and alert on drift
- Validate agent sandbox configs
- Surface alerts via desktop notification + SQLite-backed history
- Maintain a clean read-only boundary with monitored systems
- Stay a single Go binary, no daemons, no cloud

### Always No Until Core Is Perfect
- Modifying the monitored system in any way (read-only, always)
- Cloud services, APIs, telemetry, accounts
- LLM-powered analysis (other tools in the ecosystem can layer on top)
- Multi-user features or team dashboards
- Webhooks (we tail; we don't receive HTTP)

## The Guard Dog Standard

**"Good dogs stay alert, bark at real threats, ignore squirrels."**

- **Local-first over cloud-dependent** — your data stays home
- **Observation over participation** — we read, we don't intervene
- **Precision over noise** — fewer, higher-quality alerts
- **Privacy over features** — no telemetry, no exfiltration
- **Subtraction over accumulation** — every feature must earn its keep

## Success Metric

**"Did doggowoof catch something the user wouldn't have noticed?"**

Secondary:
- Zero false-positive notifications in a typical session
- Cursor persistence works across restarts
- Watch mode runs indefinitely without leaks or drift

## Core Workflow

```bash
doggo bore --watch     # run alongside an agent session (terminal A)
doggo bore             # check status anytime (terminal B)
doggo alerts           # review what was caught
doggo alerts --ack=42  # acknowledge handled findings
```

## Anti-Feature-Creep Rules

Before adding anything:
1. **Does this catch a real failure mode in real agent systems?**
2. **Does it preserve read-only observation?**
3. **Does it work fully offline?**
4. **Would it have caught the Deep Bore self-modification incident?**

## Marketing Honesty Principle

Underpromise, overdeliver.

- "Tails agent logs and barks" not "AI-powered safety platform"
- "SHA-256 binary integrity" not "Cryptographic supply-chain security"
- "Local SQLite" not "Enterprise observability backend"
- Dog metaphors remain non-negotiable

## Roadmap

### Phase 1: Agent Observability Foundation (complete)
- Log tailing with regex-based event parsing
- Sandbox violation detection (forbidden tool calls)
- Self-modification detection (build commands inside tool calls)
- Binary SHA-256 integrity checking
- Agent config validation (sandbox + disallowedTools)
- Session state tracking (task progress, cost, turns)
- Cursor persistence + log rotation handling
- Desktop notifications for critical/high

### Phase 2: Sharper signal
- Configurable rule profiles (multiple watch targets per install)
- Cost spike detection (alert when a task burns more than expected)
- Token-budget tracking
- Better dedup for repeat broadcasts

### Phase 3: Ecosystem integration
- uroboro integration: watch capture stream for stale blockers, undeclared decisions
- Generic file/log watcher mode (define rules in TOML, point at any log)
- Optional cross-tool alert routing

## Place in the QRY Ecosystem

doggowoof is the **alerting / observability layer** in a small set of personally-built tools:

- **uroboro** — context capture (decisions, blockers, questions)
- **osmotic** — OSINT collection and analysis
- **miqro** — local voice transcription
- **doggowoof** — watching layer; barks when patterns appear
- **openclaw** — sandbox for AI assistant experimentation

Each tool is a deliberately scoped exercise in AI integration. doggowoof's exercise is: **agent safety observability without participating in the agent.**
