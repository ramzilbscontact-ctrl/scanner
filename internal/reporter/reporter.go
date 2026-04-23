// Package reporter formats a Report into the various output formats.
package reporter

import (
	"fmt"
	"io"
	"time"

	"github.com/agenzia/scanner/internal/collectors"
	"github.com/agenzia/scanner/internal/nis2"
	"github.com/agenzia/scanner/internal/scoring"
	"github.com/fatih/color"
)

type Report struct {
	ID        string              `json:"id"`
	Version   string              `json:"scanner_version"`
	Host      *collectors.Data    `json:"host"`
	Results   *nis2.Results       `json:"results"`
	Score     *scoring.Score      `json:"score"`
	GeneratedAt time.Time         `json:"generated_at"`
}

func New(d *collectors.Data, r *nis2.Results, s *scoring.Score, version string) *Report {
	return &Report{
		ID:          generateID(d),
		Version:     version,
		Host:        d,
		Results:     r,
		Score:       s,
		GeneratedAt: time.Now().UTC(),
	}
}

func generateID(d *collectors.Data) string {
	return fmt.Sprintf("%s-%d", d.Hostname, time.Now().Unix())
}

// ─── Pretty terminal output ───────────────────────────────────────
func (r *Report) Pretty(w io.Writer) {
	hdr := color.New(color.FgHiCyan, color.Bold)
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	grey := color.New(color.FgHiBlack)

	fmt.Fprintln(w)
	hdr.Fprintln(w, "═══════════════════════════════════════════════════════════")
	hdr.Fprintln(w, "                  AGENZIA NIS2 SCAN REPORT                 ")
	hdr.Fprintln(w, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(w)

	// Host
	bold.Fprintln(w, "Host")
	fmt.Fprintf(w, "  %-14s %s\n", "Hostname:", r.Host.Hostname)
	fmt.Fprintf(w, "  %-14s %s\n", "Platform:", r.Host.Platform)
	fmt.Fprintf(w, "  %-14s %s\n", "Scanned at:", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintln(w)

	// Overall score
	bold.Fprint(w, "Overall Agenzia Score : ")
	scoreColor := green
	switch {
	case r.Score.Overall < 40:
		scoreColor = red
	case r.Score.Overall < 75:
		scoreColor = yellow
	}
	scoreColor.Fprintf(w, "%d/100", r.Score.Overall)
	fmt.Fprintf(w, "   %s\n\n", r.Score.Verdict)

	// Per-measure table
	bold.Fprintln(w, "NIS2 Art. 21 — Scores by measure")
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────")
	for m := 1; m <= 10; m++ {
		sc, ok := r.Score.ByMeasure[m]
		if !ok {
			grey.Fprintf(w, "  %d. %-35s not evaluated\n", m, measureName(m))
			continue
		}
		status := "✓"
		statusColor := green
		if sc < 60 {
			status = "✗"
			statusColor = red
		} else if sc < 80 {
			status = "⚠"
			statusColor = yellow
		}
		statusColor.Fprintf(w, "  %s ", status)
		fmt.Fprintf(w, "%d. %-35s ", m, measureName(m))
		scoreColor := green
		switch {
		case sc < 40:
			scoreColor = red
		case sc < 75:
			scoreColor = yellow
		}
		scoreColor.Fprintf(w, "%d/100\n", sc)
	}
	fmt.Fprintln(w)

	// Critical gaps
	if len(r.Score.TopGaps) > 0 {
		red.Fprintln(w, "⚠ Critical gaps (requires immediate attention)")
		fmt.Fprintln(w, "───────────────────────────────────────────────────────────")
		for i, gap := range r.Score.TopGaps {
			if i >= 5 {
				break
			}
			fmt.Fprintf(w, "  %d. %s\n", i+1, gap)
		}
		fmt.Fprintln(w)
	}

	// Quick wins
	if len(r.Score.QuickWins) > 0 {
		green.Fprintln(w, "🚀 Quick wins (top actions to run this week)")
		fmt.Fprintln(w, "───────────────────────────────────────────────────────────")
		for i, q := range r.Score.QuickWins {
			fmt.Fprintf(w, "  %d. %s\n", i+1, q.Action)
			grey.Fprintf(w, "       └─ Impact: %s · Effort: %s · Gain: +%d pts\n",
				q.Impact, q.Effort, q.ScoreGain)
		}
		fmt.Fprintln(w)
	}

	// Footer
	grey.Fprintln(w, "───────────────────────────────────────────────────────────")
	grey.Fprintln(w, "This is a v0.1 scan. For continuous monitoring, upload to ")
	grey.Fprintln(w, "dashboard.agenzia.uk (free) or run with --upload --api-key ")
	fmt.Fprintln(w)
}

// ─── Markdown output ──────────────────────────────────────────────
func (r *Report) Markdown(w io.Writer) {
	fmt.Fprintf(w, "# Agenzia NIS2 Scan Report — `%s`\n\n", r.Host.Hostname)
	fmt.Fprintf(w, "> **Overall score:** `%d/100` — %s\n\n", r.Score.Overall, r.Score.Verdict)
	fmt.Fprintf(w, "- **Host:** %s\n", r.Host.Hostname)
	fmt.Fprintf(w, "- **Platform:** %s\n", r.Host.Platform)
	fmt.Fprintf(w, "- **Scanned at:** %s\n\n", r.GeneratedAt.Format(time.RFC3339))

	fmt.Fprintln(w, "## NIS2 measure breakdown")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| # | Measure | Score |")
	fmt.Fprintln(w, "|---|---------|-------|")
	for m := 1; m <= 10; m++ {
		if sc, ok := r.Score.ByMeasure[m]; ok {
			fmt.Fprintf(w, "| %d | %s | `%d/100` |\n", m, measureName(m), sc)
		} else {
			fmt.Fprintf(w, "| %d | %s | _n/a_ |\n", m, measureName(m))
		}
	}
	fmt.Fprintln(w)

	if len(r.Score.TopGaps) > 0 {
		fmt.Fprintln(w, "## ⚠️ Critical gaps")
		fmt.Fprintln(w)
		for _, g := range r.Score.TopGaps {
			fmt.Fprintf(w, "- %s\n", g)
		}
		fmt.Fprintln(w)
	}

	if len(r.Score.QuickWins) > 0 {
		fmt.Fprintln(w, "## 🚀 Quick wins")
		fmt.Fprintln(w)
		for i, q := range r.Score.QuickWins {
			fmt.Fprintf(w, "%d. **%s** (+%d pts — %s impact)\n", i+1, q.Action, q.ScoreGain, q.Impact)
		}
	}
}

func measureName(n int) string {
	names := []string{
		"", // index 0 unused
		"Risk management policy",
		"Incident handling",
		"Business continuity (backup)",
		"Supply chain security",
		"Secure development & maintenance",
		"Effectiveness assessment",
		"Cyber hygiene & training",
		"Cryptography",
		"Access control & assets",
		"MFA & strong authentication",
	}
	if n < 1 || n > 10 {
		return "Unknown"
	}
	return names[n]
}
