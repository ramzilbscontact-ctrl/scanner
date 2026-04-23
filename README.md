# 🛡️ Agenzia Scanner

> **Open-source NIS2 compliance scanner for French SMBs.**
> Run it locally. Get a compliance score in 60 seconds. No data leaves your infrastructure.

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev/)
[![Build](https://github.com/getagenzia/scanner/actions/workflows/ci.yml/badge.svg)](https://github.com/getagenzia/scanner/actions)
[![Release](https://img.shields.io/github/v/release/getagenzia/scanner)](https://github.com/getagenzia/scanner/releases)

---

## ⚡ Quick start

```bash
# macOS / Linux
curl -sSL https://agenzia.uk/scanner/install.sh | sh

# Windows PowerShell
iwr -useb https://agenzia.uk/scanner/install.ps1 | iex

# Or download the binary
# https://github.com/getagenzia/scanner/releases/latest
```

Then:

```bash
agenzia-scan
```

You'll see in 60 seconds:

```
🛡️  Agenzia Scanner v0.1.0

Running NIS2 scan on hostname-xyz (linux)...

┌─ Results ──────────────────────────────────────────┐
│  Overall NIS2 Score : 68/100   🟡 Medium risk      │
├────────────────────────────────────────────────────┤
│  ✓  Measure 01 — Risk management policy    80/100  │
│  ✗  Measure 02 — Incident handling          30/100 │
│  ✓  Measure 03 — Business continuity       75/100  │
│  ⚠️  Measure 07 — Cyber hygiene             55/100 │
│  ✗  Measure 10 — MFA everywhere             20/100 │
│  ...                                               │
├────────────────────────────────────────────────────┤
│  Top 3 critical gaps :                             │
│  1. MFA not enabled on 4 admin accounts            │
│  2. No documented incident response plan           │
│  3. Backup tests not performed in 6+ months        │
└────────────────────────────────────────────────────┘

📄 Full report saved to: ./agenzia-report.json
🌐 Optional: upload to dashboard.agenzia.uk for history + remediation
```

---

## 🎯 Why does this exist?

The EU **NIS2 directive** requires 15,000–18,000 French SMBs to demonstrate
compliance with 10 mandatory cybersecurity measures by **October 17, 2026**.

Penalties: **up to €10M or 2% global revenue + personal liability for directors.**

Commercial NIS2 audit tools are:
- 🇺🇸 US-based (not adapted to ANSSI expectations)
- 💰 €15,000–50,000/year (out of reach for SMBs)
- 🔒 Closed-source (you can't inspect what they do with your data)

**Agenzia Scanner is different**:
- 🇫🇷 Built in France, following ANSSI recommendations
- 🆓 Fully free and open-source (Apache 2.0)
- 🔍 Auditable: read the code, see exactly what's collected
- 💻 Runs **on your machine** — no data sent anywhere unless you opt-in
- ⚡ Fast: 60-second scan, zero configuration

---

## 🚀 What it scans

| Area | Windows | Linux | macOS | M365 | Google Workspace |
|------|---------|-------|-------|------|------------------|
| **User accounts & MFA** | ✅ | ✅ | ✅ | ✅ | 🟡 planned |
| **Patch level** | ✅ | ✅ | ✅ | — | — |
| **Local firewall config** | ✅ | ✅ | ✅ | — | — |
| **Backup status** | ✅ | 🟡 | 🟡 | ✅ | 🟡 planned |
| **Shared folders / permissions** | ✅ | ✅ | ✅ | ✅ | — |
| **Admin privilege review** | ✅ | ✅ | ✅ | ✅ | — |
| **Endpoint protection** | ✅ | 🟡 | 🟡 | — | — |

Legend: ✅ supported · 🟡 partial · — N/A

---

## 🧪 The 10 NIS2 measures we check (Art. 21)

1. **Risk analysis policy** — Do you have a documented ISO 27005/EBIOS RM policy?
2. **Incident handling** — Do you have a formal IR process + escalation?
3. **Business continuity** — Backups, DRP, tested restoration
4. **Supply chain security** — Vendor NIS2 compliance tracking
5. **Acquisition, development, maintenance** — Secure SDLC practices
6. **Effectiveness assessment** — Are your measures actually working?
7. **Cyber hygiene & training** — User training + best practices
8. **Cryptography** — Data encryption at rest + in transit
9. **Access control & asset management** — Least privilege, asset inventory
10. **MFA everywhere + SSO** — Strong authentication for all users

Each measure has **10–30 technical checks** that run automatically.

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────┐
│  agenzia-scan (single Go binary, 10 MB)              │
├──────────────────────────────────────────────────────┤
│  ┌── collectors/ ─────────────────────────────┐      │
│  │  Platform-specific system probes           │      │
│  │  • windows.go (WMI, PowerShell)            │      │
│  │  • linux.go  (/etc, systemd, journalctl)   │      │
│  │  • macos.go  (defaults, plist)             │      │
│  │  • m365.go   (Microsoft Graph API)         │      │
│  │  • gws.go    (Google Admin SDK)            │      │
│  └────────────────────────────────────────────┘      │
│                      ↓                                │
│  ┌── nis2/ ───────────────────────────────────┐      │
│  │  Checks organized by NIS2 measure          │      │
│  │  • measure_01_risk_policy.go              │      │
│  │  • measure_07_cyber_hygiene.go            │      │
│  │  • measure_10_mfa.go                      │      │
│  │  ... (10 files, ~150 individual checks)   │      │
│  └────────────────────────────────────────────┘      │
│                      ↓                                │
│  ┌── scoring/ + reporter/ ────────────────────┐      │
│  │  Aggregate checks into 0-100 scores         │      │
│  │  Output: JSON, pretty terminal, PDF (opt)   │      │
│  └────────────────────────────────────────────┘      │
└──────────────────────────────────────────────────────┘
```

**Why Go**: fast compile, static binary, zero runtime dependency, works on Windows/Linux/macOS from a single codebase, accessible to contributors.

---

## 📦 Installation

### Option 1 — One-liner (recommended)

**macOS / Linux**:
```bash
curl -sSL https://agenzia.uk/scanner/install.sh | sh
```

**Windows PowerShell**:
```powershell
iwr -useb https://agenzia.uk/scanner/install.ps1 | iex
```

### Option 2 — Download binary

Grab the latest release: https://github.com/getagenzia/scanner/releases/latest

### Option 3 — From source

```bash
git clone https://github.com/getagenzia/scanner.git
cd scanner
go build -o agenzia-scan ./cmd/agenzia-scan
./agenzia-scan
```

### Option 4 — Docker

```bash
docker run --rm -v /:/host:ro ghcr.io/getagenzia/scanner:latest
```

---

## 🎮 Usage examples

### Basic scan (local machine)

```bash
agenzia-scan
```

### Scan + upload to Agenzia dashboard (optional, signup required)

```bash
agenzia-scan --upload --api-key <your-key>
```

### Scan specific measures only

```bash
agenzia-scan --measures 2,7,10
```

### Run only checks that require admin/root

```bash
sudo agenzia-scan --privileged
```

### Output formats

```bash
agenzia-scan --format json       # JSON (for pipelines)
agenzia-scan --format pretty     # Pretty terminal (default)
agenzia-scan --format markdown   # Markdown (for GitHub)
agenzia-scan --format pdf        # PDF report (saves to ./agenzia-report.pdf)
```

### Scan remote Microsoft 365 tenant

```bash
agenzia-scan --m365 --tenant contoso.onmicrosoft.com
```

### Continuous monitoring (daemon mode)

```bash
agenzia-scan daemon --interval 24h --upload
```

---

## 🔐 Data privacy

**What's collected**:
- System config (OS version, patches, firewall rules)
- User/account metadata (names, admin status, MFA status)
- Service status (backup, EDR, disk encryption)
- File permissions on critical paths

**What's NEVER collected**:
- File contents (we read permissions, not data)
- Passwords (even hashed)
- Network traffic / PCAPs
- Browsing history
- Personal user data

**Where data goes**:
- By default: **nowhere**. Report saved locally as JSON.
- With `--upload`: sent over HTTPS to `dashboard.agenzia.uk` (your account only).
- You can self-host the dashboard (see `docker-compose.yml`).

---

## 🤝 Contributing

We welcome contributions! Agenzia Scanner is built by and for the French
cybersecurity community.

### Quick contribution guide

```bash
# Fork, clone
git clone https://github.com/YOUR_USERNAME/scanner.git
cd scanner

# Install dependencies
go mod download

# Run tests
make test

# Run the scanner locally
go run ./cmd/agenzia-scan

# Create a new NIS2 check
make new-check MEASURE=7 NAME=password-policy
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### 🎯 Good first issues

Check [issues labeled `good first issue`](https://github.com/getagenzia/scanner/labels/good%20first%20issue) — we've curated 15+ beginner-friendly tasks.

---

## 🌍 Community

- 🐦 Twitter: [@agenzia_fr](https://twitter.com/agenzia_fr)
- 💼 LinkedIn: [Agenzia](https://linkedin.com/company/agenzia)
- 📧 Email: hello@agenzia.uk
- 💬 Discord: [discord.gg/agenzia](https://discord.gg/agenzia) (coming soon)

---

## 🗺️ Roadmap

### v0.1 — current (April 2026)
- ✅ Core NIS2 scanner (3 measures: risk policy, cyber hygiene, MFA)
- ✅ Windows/Linux/macOS collectors
- ✅ JSON + pretty terminal output
- ✅ GitHub Actions CI

### v0.2 — May 2026
- [ ] All 10 NIS2 measures covered
- [ ] M365 + Google Workspace collectors
- [ ] PDF report generation (native, no external service)
- [ ] ISO 27001 mapping (Annex A, 93 controls)

### v1.0 — June 2026
- [ ] Daemon mode (continuous monitoring)
- [ ] Web UI (self-hostable dashboard)
- [ ] Plugin system for custom rules
- [ ] Multi-tenant architecture
- [ ] DORA + GDPR scanners

### v1.5 — Q3 2026
- [ ] Plugin: AWS / OVH / Scaleway / Azure cost scanner (FinOps)
- [ ] Plugin: LLM traffic analyzer (AI Security)
- [ ] Auto-remediation scripts (opt-in)

---

## 🏢 Commercial offering

The scanner is **free forever** under Apache 2.0.

For businesses that want **more than a scan**, Agenzia also provides:

| Need | Solution | Price |
|------|----------|-------|
| Hosted dashboard + history | **Agenzia Starter** | 99 €/mo |
| +Human remediation 24/7 | **Agenzia Pro** | 49 €/poste/mo |
| +RSSI + SOC managed | **Agenzia Sovereign** | 79 €/poste/mo |

Learn more: [agenzia.uk/saas](https://agenzia.uk/saas)

---

## 📜 License

Apache License 2.0. See [LICENSE](LICENSE).

**TL;DR**: do whatever you want with this code, including commercial use.
Just don't sue us, and include attribution.

---

## ⭐ Star history

If this tool helped you, please give it a star. It helps others discover it.

[![Star History Chart](https://api.star-history.com/svg?repos=agenzia/scanner&type=Date)](https://star-history.com/#agenzia/scanner&Date)

---

<div align="center">

**Built with ❤️ in Paris by [Agenzia](https://agenzia.uk)**

*Making cybersecurity compliance accessible to every French SMB.*

</div>
