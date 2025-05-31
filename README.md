# DOGGOWOOF 🚨🐕🚨

**THE LOCAL-FIRST ALERT TRIAGE SYSTEM THAT ACTUALLY BARKS!**

Your personal guard dog for development alerts. Smart, private, and LOUD when it matters.

> *🗣️ "WOOF! WOOF! HEY HUMAN, THIS ONE ACTUALLY MATTERS!" 🗣️*

## 🔥 THE VISION

**STOP THE NOTIFICATION MADNESS!** DOGGOWOOF transforms information overload into actionable insights with the enthusiasm of a golden retriever and the precision of a border collie.

**PHILOSOPHY**: Local-first, privacy-focused, ZERO BS monitoring with BIG ENERGY when alerts matter!

## 🏗️ ARCHITECTURE (IT'S LOUD BUT ORGANIZED!)

```
┌─ Go CLI ────────────┐    ┌─ Python Daemon ─────┐    ┌─ Svelte Dashboard ─┐
│ • doggo init        │ ←→ │ • Webhook receiver  │ ←→ │ • Live monitoring  │
│ • doggo daemon      │    │ • AI triage         │    │ • Training UI      │
│ • doggo train       │    │ • SQLite storage    │    │ • Alert history    │
│ • doggo status      │    │ • Desktop alerts    │    │ • Local only       │
└─────────────────────┘    └─────────────────────┘    └────────────────────┘
```

## 🚀 QUICK START (GET THAT TAIL WAGGING!)

```bash
# Install your new best friend
go install github.com/QRY91/doggowoof@latest

# Initialize (creates config, DB, starts daemon)
doggo init

# Add alert sources (FEED THE DOGGO!)
doggo watch --discord --webhook "https://discord.com/api/webhooks/..."
doggo watch --github --repo "username/repo"

# Train on your patterns (GOOD BOY LEARNS!)
doggo train --from-resolved

# Check what's happening (STATUS CHECK!)
doggo status
```

## 🎯 USE CASES (WHERE DOGGOWOOF SHINES!)

- **Solo Developers**: NO MORE MISSED DEPLOY FAILURES! 
- **Small Teams**: SHARED ALERT INTELLIGENCE THAT ACTUALLY WORKS!
- **DevOps**: INFRASTRUCTURE ALERTS WITH PERSONALITY!
- **Researchers**: PUBLICATION ALERTS WITHOUT THE SPAM!

## 🔧 INTEGRATION (PLAYS WELL WITH OTHERS!)

- **Discord**: Webhook monitoring with SMART FILTERING!
- **GitHub**: Issue/PR/CI alert prioritization with ENERGY!
- **Email**: IMAP monitoring for critical services!
- **Webhooks**: Generic HTTP endpoint monitoring!
- **Uroboro**: Cross-pollinate insights for content generation!

## 📊 LOCAL DATA (YOUR DATA STAYS HOME!)

Everything stays on YOUR machine:
- **SQLite database**: Alert history, training data
- **Local AI models**: Pattern recognition, triage decisions  
- **Privacy-first**: NO TELEMETRY, NO CLOUD DEPENDENCIES!

## 🎨 COMPONENTS (THE PACK!)

### CLI (`cli/`) - THE QUIET COMPANION
Go-based command interface. Fast, single binary deployment. Respectful Unix citizen.

### Daemon (`daemon/`) - THE WORKING DOG
Python background service. Webhook processing, AI analysis, LOUD notifications when needed.

### Dashboard (`dashboard/`) - THE DOGGY DASHBOARD  
Svelte web interface. Real-time monitoring, training, configuration. LOCAL ONLY!

### Shared (`shared/`) - THE SHARED KNOWLEDGE
Common schemas, database models, analysis logic.

## 🛠️ DEVELOPMENT (JOIN THE PACK!)

```bash
# Clone and setup
git clone git@github.com:QRY91/doggowoof.git
cd doggowoof

# Install dependencies
make install

# Run development suite  
make dev

# Build all components
make build
```

## 🗺️ ROADMAP (THE TRAINING PLAN!)

- [x] Project genesis and architecture design
- [x] Go CLI foundation with basic commands
- [ ] Python daemon with webhook receiver
- [ ] SQLite schema and data models
- [ ] Svelte dashboard with real-time updates
- [ ] Discord integration (WOOF AT DISCORD!)
- [ ] GitHub integration (BARK AT BUGS!)
- [ ] Local AI training pipeline (SMART DOGGO!)
- [ ] Desktop notification system (LOUD WHEN NEEDED!)
- [ ] Documentation and examples

## 📈 INSPIRATION (BRED FROM THE BEST!)

Built from experience with:
- **Logistics monitoring** (Mileviewer/Maersk scale - ENTERPRISE GRADE!)
- **Project dashboards** (Panopticron internal tooling - BATTLE TESTED!)
- **Developer workflows** (Uroboro content generation - PRODUCTIVITY FOCUSED!)

**THE GOAL**: Bring enterprise-grade alert intelligence to individual developers with the enthusiasm of a VERY GOOD BOY! 🐕‍🦺

---

## 🚨 WHY DOGGOWOOF? 🚨

Because your alerts deserve a guard dog that:
- ✅ **STAYS AWAKE** while you code
- ✅ **BARKS LOUD** when something's actually wrong  
- ✅ **STAYS QUIET** when it's just noise
- ✅ **LEARNS YOUR PATTERNS** like a loyal companion
- ✅ **PROTECTS YOUR PRIVACY** (no data leaves home!)

---

***DOGGOWOOF: BECAUSE YOUR ALERTS DESERVE A VERY GOOD BOY WATCHING OVER THEM!*** 🚨🐕‍🦺🚨 