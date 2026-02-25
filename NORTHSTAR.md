# DOGGOWOOF North Star

**Last updated**: February 2026

## The One Problem We Solve

**Guard dog for OSINT data integrity and autonomous agent safety.**

Watch osmotic's collected data for suspicious patterns: temporal clustering, behavioral uniformity, coordination signals. Alert when public information flows show signs of manipulation. Watch autonomous AI agents for sandbox violations, self-modification, and integrity failures.

## Focus Guardrails

### Always Yes
- Read osmotic's data and detect coordination patterns
- Profile account behavior across sources
- Monitor autonomous agents for safety violations
- Alert loud on high-confidence findings
- Keep everything private and local-first
- Maintain clear read-only boundary with monitored systems

### Always No Until Core Is Perfect
- Writing to osmotic's database or modifying monitored systems (read-only, always)
- LLM-powered analysis (that's osmotic's job)
- Cloud services, APIs, or SaaS features
- Team collaboration or multi-user features
- Complex dashboards

## The Guard Dog Standard

**"Good dogs stay alert, bark at real threats, ignore squirrels."**

- **Local-first over cloud-dependent** — your data stays home
- **Detection over collection** — osmotic collects, we detect
- **Precision over noise** — fewer, higher-quality alerts
- **Privacy over features** — no telemetry, no cloud dependencies
- **Read-only over read-write** — we never modify source data

## Success Metric

**"Did DoggoWoof find something real this week?"**

Secondary metrics:
- False positive rate decreasing over time
- Detection accuracy improving with more data
- Zero writes to osmotic's database (enforced by PRAGMA query_only)

## Core Workflows

### OSINT Coordination Detection
```bash
osmotic collect                              # 1. osmotic collects data
doggo sniff                                  # 2. doggowoof sniffs for patterns
doggo status                                 # 3. review findings
doggo alerts --ack=42                        # 4. acknowledge or investigate
osmotic deep --entity="Suspicious Account"
```

### Deep Bore Agent Monitoring
```bash
doggo bore --watch    # watch bridge logs (run alongside agent session)
doggo bore            # check session status from another terminal
doggo alerts          # review agent-related alerts
```

## Anti-Feature-Creep Rules

Before adding anything:
1. **Does this improve detection accuracy?**
2. **Does it maintain read-only access to osmotic?**
3. **Does this keep data local and private?**
4. **Would an investigative journalist want this?**

## Marketing Honesty Principle

Underpromise, overdeliver.

- "Detects temporal clusters" not "AI-powered threat intelligence"
- "Profiles posting patterns" not "Advanced behavioral analytics platform"
- "Local SQLite" not "Enterprise-grade data warehouse"
- Dog metaphors remain non-negotiable

## Roadmap

### Phase 1: Detection Foundation (complete)
- Temporal clustering detection
- Account behavior profiling
- Combined coordination scoring
- Alert management (list, acknowledge)
- Desktop notifications for high/critical
- Deep Bore agent monitoring (log tailing, sandbox violation detection, binary integrity, config validation)

### Phase 2: Refinement
- Historical pattern tracking (clusters that persist across runs)
- Improved topic extraction (leverage osmotic's analysis passes)
- Alert deduplication and merging
- Configurable alert thresholds per source

### Phase 3: Advanced Detection
- Network analysis (who replies to whom)
- Narrative tracking (same talking points across accounts)
- Anomaly detection (accounts that change behavior suddenly)
- Cross-source coordination (same patterns across reddit + hn + rss)

## Relationship to Osmotic

**Complementary, not competing:**
- **Osmotic**: Collects, normalizes, analyzes, publishes
- **DoggoWoof**: Reads osmotic data, detects coordination, alerts. Watches agents.

**Shared philosophy**: Local-first, privacy-focused, durable outputs, minimal dependencies.

**Integration points**: osmotic.sqlite (read-only), bridge logs (tail-only). The file system IS the API. Dog watches. Dog does not touch.
