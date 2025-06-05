# DOGGOWOOF 🚨🐕🚨

**THE LOCAL-FIRST ALERT TRIAGE SYSTEM THAT ACTUALLY BARKS!**

Your personal guard dog for development alerts. Smart, private, and LOUD when it matters.

> *🗣️ "WOOF! WOOF! HEY HUMAN, THIS ONE ACTUALLY MATTERS!" 🗣️*

## 🔥 THE VISION

**STOP THE NOTIFICATION MADNESS!** DOGGOWOOF transforms information overload into actionable insights with the enthusiasm of a golden retriever and the precision of a border collie.

**PHILOSOPHY**: Local-first, privacy-focused, ZERO BS monitoring with BIG ENERGY when alerts matter!

## 🏗️ ARCHITECTURE (MVP WORKING!)

```
┌─ Go CLI ────────────┐    ┌─ Python Daemon ─────┐    
│ • doggo init        │ ←→ │ • Webhook receiver  │
│ • doggo daemon      │    │ • Basic filtering   │    
│ • doggo status      │    │ • SQLite storage    │    
│ • doggo pet         │    │ • Desktop alerts    │    
└─────────────────────┘    └─────────────────────┘    
```

**What's Working Now**: GitHub CI failure detection, local SQLite storage, desktop notifications  
**What's Next**: Pattern learning, smart filtering, more integrations

## 🚀 QUICK START (GET THAT TAIL WAGGING!)

```bash
# Install your guard dog (Go way!)
go install github.com/QRY91/doggowoof/cmd/doggowoof@latest
ln -sf $(go env GOPATH)/bin/doggowoof $(go env GOPATH)/bin/doggo

# Initialize your guard dog
doggo init

# Start the daemon (your guard dog goes to work!)
doggo daemon start

# Check what your guard dog is up to
doggo status

# Test with a GitHub CI failure webhook
curl -X POST http://localhost:8080/webhook/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: workflow_run" \
  -d '{"action":"completed","workflow_run":{"name":"Tests","conclusion":"failure"},"repository":{"name":"your-repo"}}'

# See the alert in your status report!
doggo status

# Pet the good doggo (secret command!)
doggo pet
```

## 🎯 USE CASES (WHERE DOGGOWOOF IS LEARNING TO SHINE!)

- **Solo Developers**: GitHub CI failure detection (working now!)
- **Local Development**: Basic webhook alerting (working now!)
- **Future**: Smart filtering, team features, broader integrations

## 🔧 INTEGRATION (GROWING PACK!)

**Working Now:**
- **GitHub**: CI failure detection via webhooks
- **Generic Webhooks**: Basic HTTP endpoint receiver

**Coming Soon:**
- **Uroboro**: Development workflow insights
- **Email/IMAP**: Critical service monitoring
- **Discord**: Alert routing

## 📊 LOCAL DATA (YOUR DATA STAYS HOME!)

Everything stays on YOUR machine:
- **SQLite database**: Alert history, basic metrics (working now!)
- **Pattern learning**: Coming soon - will learn what you care about
- **Privacy-first**: NO TELEMETRY, NO CLOUD DEPENDENCIES! (guaranteed!)

## 🎨 COMPONENTS (THE PACK!)

### CLI (`internal/cli/`) - THE QUIET COMPANION ✅
Go-based command interface. Fast, single binary deployment. Working now!

### Daemon (`daemon/`) - THE WORKING DOG ✅  
Python background service. Webhook processing, basic filtering, desktop notifications. Working now!

### Future Components 🔜
- **Smart Filtering**: Pattern learning and intelligent triage
- **Advanced Integrations**: More alert sources and routing options

## 🛠️ DEVELOPMENT (JOIN THE PACK!)

```bash
# Clone and setup
git clone git@github.com:QRY91/doggowoof.git
cd doggowoof

# Install the guard dog
go install ./cmd/doggowoof
ln -sf $(go env GOPATH)/bin/doggowoof $(go env GOPATH)/bin/doggo

# Initialize and test
doggo init
doggo daemon start
doggo status

# Start contributing!
# Python daemon is in daemon/ (working webhook receiver)
# Go CLI is in internal/cli/ (init, daemon, status, pet commands)
# Next: Pattern learning and smart filtering
```

## 🗺️ ROADMAP (THE TRAINING PLAN!)

**MVP ACHIEVED! 🎉**
- [x] Project genesis and architecture design
- [x] Go CLI foundation with complete commands (`init`, `daemon`, `status`)
- [x] Python daemon with webhook receiver (WORKING!)
- [x] SQLite schema and data models (STORING ALERTS!)
- [x] GitHub CI failure integration (BARKING AT BUGS!)
- [x] Desktop notification system (LOUD WHEN NEEDED!)
- [x] Local-first operation (YOUR DATA STAYS HOME!)

**Next Phase - Pattern Learning:**
- [ ] Alert acknowledgment tracking (learn what you care about)
- [ ] Basic pattern recognition (remember your responses)
- [ ] Smart filtering (reduce noise over time)
- [ ] Uroboro integration (development workflow alerts)
- [ ] Email/IMAP monitoring (critical service notifications)

## 📈 INSPIRATION (BRED FROM THE BEST!)

Built from experience with:
- **Logistics monitoring** (Mileviewer/Maersk scale - learned what works at scale)
- **Project dashboards** (Panopticron internal tooling - battle-tested approaches)
- **Developer workflows** (Uroboro content generation - local-first privacy focus)

**THE GOAL**: Bring proven alert intelligence patterns to individual developers with the enthusiasm of a VERY GOOD BOY! 🐕‍🦺

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