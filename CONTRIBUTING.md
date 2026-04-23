# Contributing to Agenzia Scanner

First, thank you for your interest! 🙏

Agenzia Scanner is built by and for the French cybersecurity community.
Every contribution — a bug report, a typo fix, a new NIS2 check — makes
PMEs safer.

---

## 🎯 Ways to contribute

### 1. Report a bug

Found something broken? [Open an issue](https://github.com/agenzia/scanner/issues/new?labels=bug).

Include:
- Your OS + version
- The command you ran
- The output (paste `--verbose` logs if possible)
- What you expected vs what happened

### 2. Suggest a new NIS2 check

Have an idea for a check that would help PMEs? [Open a feature request](https://github.com/agenzia/scanner/issues/new?labels=enhancement).

Describe:
- Which NIS2 measure it relates to (Article 21, measures 1-10)
- What to check
- How to score it (pass/fail or 0-100)

### 3. Write code

Grab a [good first issue](https://github.com/agenzia/scanner/labels/good%20first%20issue) or propose your own.

---

## 🛠️ Development setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/scanner.git
cd scanner

# Install dependencies
go mod download

# Run tests
go test ./...

# Build locally
go build -o agenzia-scan ./cmd/agenzia-scan

# Run
./agenzia-scan
```

### Project layout

```
cmd/agenzia-scan/        # Entry point (CLI)
internal/collectors/     # OS-specific system probes
internal/nis2/           # NIS2 checks (one .go file per measure)
internal/scoring/        # Aggregation + Agenzia Score
internal/reporter/       # Output formats
```

---

## ➕ Adding a new NIS2 check

Let's say you want to add a check for measure 4 (supply chain security): verify that all installed npm/pip/system packages are up to date.

1. Open `internal/nis2/checks.go`
2. Add a new function:

   ```go
   func checkPackageFreshness(d *collectors.Data) Check {
       // Your logic here
       return Check{
           Measure:     4,
           ID:          "4.1-packages-fresh",
           Title:       "Installed packages up to date",
           Score:       80,
           Severity:    "high",
           Passed:      true,
           Finding:     "Your finding here",
           Recommendation: "Run apt upgrade / brew upgrade / choco upgrade weekly",
       }
   }
   ```

3. Register it in `RunAllChecks()` in `nis2.go`
4. Add a unit test in `checks_test.go`
5. Open a PR with a clear description

---

## 🧪 Testing guidelines

- Every new check MUST have at least one unit test
- Tests should be deterministic (no real system calls in test)
- Use table-driven tests where possible
- Aim for >70% coverage on new code

Run tests:

```bash
go test -race -cover ./...
```

---

## 📝 Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(nis2): add check for measure 4 (supply chain security)
fix(linux): handle missing /etc/passwd gracefully
docs(readme): add Windows installation steps
```

Prefixes:
- `feat` — new feature
- `fix` — bug fix
- `docs` — docs only
- `test` — test-only changes
- `refactor` — code cleanup (no behavior change)
- `chore` — tooling / CI / deps

---

## 🧑‍⚖️ Code of Conduct

Be kind. We follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

No harassment, no discrimination, no ego — just good code for a good cause.

---

## 📜 License

By contributing, you agree that your contributions will be licensed
under the Apache License 2.0 (same as the project).

---

## 🎉 Contributors hall of fame

All contributors get listed in our README + on the website.

First 10 meaningful contributors earn a free year of **Agenzia Starter**
(99 €/mo × 12 = 1,188 € value) as a thank-you. 🎁

---

Questions? Open a discussion or email `contribute@agenzia.uk`.
