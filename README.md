# DoggoWoof 🐕

**Local-First Alert Triage System**

Your personal guard dog for development alerts. Smart, private, and always watching.

> *"Like having a loyal dog that knows which alerts actually matter"*

## ✨ Vision

Transform information overload into actionable insights. DoggoWoof learns your priorities and guards against notification fatigue.

**Philosophy**: Local-first, privacy-focused, zero BS monitoring.

## 🏗️ Architecture

```
┌─ Go CLI ────────────┐    ┌─ Python Daemon ─────┐    ┌─ Svelte Dashboard ─┐
│ • doggo init        │ ←→ │ • Webhook receiver  │ ←→ │ • Live monitoring  │
│ • doggo daemon      │    │ • AI triage         │    │ • Training UI      │
│ • doggo train       │    │ • SQLite storage    │    │ • Alert history    │
│ • doggo status      │    │ • Desktop alerts    │    │ • Local only       │
└─────────────────────┘    └─────────────────────┘    └────────────────────┘
```

## 🚀 Quick Start

```bash
# Install
go install github.com/QRY91/doggowoof@latest

# Initialize (creates config, DB, starts daemon)
doggo init

# Add alert sources  
doggo watch --discord --webhook "https://discord.com/api/webhooks/..."
doggo watch --github --repo "username/repo"

# Train on your patterns
doggo train --from-resolved

# Check what's happening
doggo status
```

## 🎯 Use Cases

- **Solo Developers**: Filter GitHub notifications, deployment alerts
- **Small Teams**: Discord channel monitoring, shared alert intelligence  
- **DevOps**: Infrastructure alerts, CI/CD pipeline monitoring
- **Researchers**: Publication alerts, collaboration updates

## 🔧 Integration

- **Discord**: Webhook monitoring with smart filtering
- **GitHub**: Issue/PR/CI alert prioritization
- **Email**: IMAP monitoring for critical services
- **Webhooks**: Generic HTTP endpoint monitoring
- **Uroboro**: Cross-pollinate insights for content generation

## 📊 Local Data

Everything stays on your machine:
- **SQLite database**: Alert history, training data
- **Local AI models**: Pattern recognition, triage decisions
- **Privacy-first**: No telemetry, no cloud dependencies

## 🎨 Components

### CLI (`cli/`)
Go-based command interface. Fast, single binary deployment.

### Daemon (`daemon/`) 
Python background service. Webhook processing, AI analysis, notifications.

### Dashboard (`dashboard/`)
Svelte web interface. Real-time monitoring, training, configuration.

### Shared (`shared/`)
Common schemas, database models, analysis logic.

## 🛠️ Development

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

## 🗺️ Roadmap

- [x] Project genesis and architecture design
- [ ] Go CLI foundation with basic commands
- [ ] Python daemon with webhook receiver
- [ ] SQLite schema and data models
- [ ] Svelte dashboard with real-time updates
- [ ] Discord integration
- [ ] GitHub integration
- [ ] Local AI training pipeline
- [ ] Desktop notification system
- [ ] Documentation and examples

## 📈 Inspiration

Built from experience with:
- **Logistics monitoring** (Mileviewer/Maersk scale)
- **Project dashboards** (Panopticron internal tooling)
- **Developer workflows** (uroboro content generation)

The goal: Bring enterprise-grade alert intelligence to individual developers.

---

*DoggoWoof: Because your alerts deserve a good boy watching over them* 🐕‍🦺 