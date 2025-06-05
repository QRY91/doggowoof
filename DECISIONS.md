# DOGGOWOOF ARCHITECTURAL DECISIONS 🚨🐕🚨

*A log of key technical decisions and their reasoning*

---

## Decision #001: Webhook Daemon Language Choice
**Status:** DECIDED - Python first, port to Go later  

### Context
Need to build webhook receiver for monitoring GitHub, Uroboro, Cursor usage, etc. Two options:
1. Build everything in Go for consistency 
2. Build webhook daemon in Python first, port later

### Decision
**Python first, port to Go after core logic is stable**

### Reasoning
**Why Python first:**
- **Iteration speed** - Webhook parsing logic will need tuning
- **Rich ecosystem** - Flask/FastAPI, great JSON handling
- **Testing flexibility** - Easy to mock different webhook formats
- **Notification libraries** - Desktop notifications, email parsing
- **Lower stakes** - Daemon can crash/restart without affecting CLI

**Why not Go immediately:**
- Webhook logic unknown complexity - better to prototype
- Different notification strategies need testing
- JSON schema evolution likely as we add sources

**Port to Go when:**
- Core webhook patterns established
- Notification strategies proven
- Performance becomes important
- Want single binary deployment

### Implementation Plan
1. ~~Python daemon with FastAPI/Flask~~ **SHIPPED: Python stdlib only** (http.server)
2. SQLite shared between Go CLI and Python daemon ✅
3. ~~Well-defined JSON schemas in `/shared`~~ **SHIPPED: Direct webhook parsing**
4. Port to Go once logic stabilizes (est. after 2-3 webhook sources working)

**What we actually built:** Simple Python daemon using only stdlib - http.server, sqlite3, subprocess for notifications. Zero external dependencies. Much simpler than planned!

---

## Decision #002: Pet Command Implementation
**Status:** SHIPPED 🐕  

### Context
Need secret easter egg command for user delight

### Decision
Hidden `doggo pet` command with random responses

### Reasoning
- User engagement and delight
- Hidden from help but discoverable
- Fits brand personality perfectly
- Zero operational impact

**Result:** Users can pet the doggo and get random wholesome responses!

---

## Decision #003: Northstar and Product Identity
**Status:** DECIDED - Core mission defined  

### Context
Need clear product identity to guide all future decisions and prevent scope creep.

### Decision
**"Stop alert fatigue from breaking solodevs"** - Local-first intelligent alert triage

### Reasoning
**Why this mission:**
- **Focused scope** - Solves ONE real problem really well
- **Clear market** - Solo developers drowning in notifications  
- **Unique position** - Local-first learning vs cloud dashboards
- **Measurable success** - Did we save you from missing something important?

**Philosophy alignment with Uroboro:**
- Local-first, privacy-focused
- Simple core workflow (not feature bloat)
- Solves "information overwhelm" (different domain)
- Honest marketing, real value

**Brand identity:**
- Guard dog metaphors (enthusiastic but focused)
- WOOF energy with smart filtering
- Privacy as core architecture, not marketing
- Anti-buzzword: "learning patterns" not "AI-powered"

---

## Decision #004: MVP Success Criteria
**Status:** ACHIEVED - Core functionality working  

### Context
What constitutes a working MVP that validates the core concept?

### Decision
**GitHub CI failure detection with local learning storage**

### Reasoning
**MVP requirements met:**
- ✅ Receives real alerts (GitHub webhooks)
- ✅ Stores locally (SQLite, zero cloud)
- ✅ Smart filtering (CI failures = HIGH priority)
- ✅ Desktop notifications (actual alerting)
- ✅ Zero dependencies (Python stdlib only)

**Validates core hypothesis:** Can we catch real problems and filter noise locally?

**Next validation:** Pattern learning - does it get smarter over time?

---

## Decision #005: Go Project Structure & Binary Management
**Status:** SHIPPED - Proper Go layout with symlinks  

### Context
Started with Makefile approach copying binaries, but wanted both `doggowoof` and `doggo` commands like Uroboro's `uroboro`/`uro` pattern.

### Decision
**Proper Go project structure with symlink approach**

### Reasoning
**What we built:**
- Moved CLI code from `cmd/` to `internal/cli/` (proper Go conventions)
- Created `cmd/doggowoof/main.go` as proper entry point
- Use `go install` + `ln -sf` for `doggo` symlink (like Uroboro)
- No Makefile needed - pure Go workflow

**Why this approach:**
- Follows Go project standards exactly
- Symlinks are cleaner than file copies
- Matches established pattern from Uroboro
- Easier for users: one install command + one symlink

**Learning:** Started with Makefile thinking we needed it, but Go conventions + symlinks are much cleaner!

---

*🐕 Keep this updated as we make architectural choices! Future you will thank present you! 🐕* 