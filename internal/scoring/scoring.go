// Package scoring aggregates check results into the Agenzia Score (0-100).
package scoring

import (
	"github.com/agenzia/scanner/internal/nis2"
)

type Score struct {
	Overall      int            `json:"overall"`
	RiskLevel    string         `json:"risk_level"` // low | medium | high | critical
	Verdict      string         `json:"verdict"`    // one-line summary
	ByMeasure    map[int]int    `json:"by_measure"`
	TopGaps      []string       `json:"top_gaps"`
	QuickWins    []QuickWin     `json:"quick_wins"`
}

type QuickWin struct {
	Action        string `json:"action"`
	Impact        string `json:"impact"`
	Effort        string `json:"effort"`
	ScoreGain     int    `json:"score_gain"`
}

// Compute builds the Score from NIS2 results.
// Weights per measure (sum = 100): critical measures weighted higher.
var weights = map[int]int{
	1:  10, // Risk policy
	2:  10, // Incident handling
	3:  15, // Business continuity — critical
	4:  8,  // Supply chain
	5:  8,  // Secure SDLC
	6:  5,  // Effectiveness
	7:  15, // Cyber hygiene — critical
	8:  8,  // Cryptography
	9:  10, // Access control
	10: 11, // MFA — critical
}

func Compute(r *nis2.Results) *Score {
	total := 0
	totalWeight := 0
	byMeasure := map[int]int{}

	for i := 1; i <= 10; i++ {
		score := r.MeasureScores[i]
		if score < 0 {
			continue // measure not evaluated
		}
		byMeasure[i] = score
		total += score * weights[i]
		totalWeight += weights[i]
	}

	overall := 0
	if totalWeight > 0 {
		overall = total / totalWeight
	}

	s := &Score{
		Overall:   overall,
		ByMeasure: byMeasure,
		TopGaps:   r.CriticalGaps,
	}

	// Risk level + verdict
	switch {
	case overall >= 90:
		s.RiskLevel = "low"
		s.Verdict = "🟢 Excellent — you are NIS2-ready. Maintain discipline."
	case overall >= 75:
		s.RiskLevel = "low"
		s.Verdict = "🟡 Good — a few gaps remain. You should be ready in 1-2 months."
	case overall >= 60:
		s.RiskLevel = "medium"
		s.Verdict = "🟠 Medium risk — expect 3-4 months of work to close the gaps."
	case overall >= 40:
		s.RiskLevel = "high"
		s.Verdict = "🔴 High risk — you will not make the October 17 deadline without help."
	default:
		s.RiskLevel = "critical"
		s.Verdict = "⚫ Critical — immediate action required. Multiple NIS2 obligations unmet."
	}

	// Quick wins from failed checks, sorted by (severity, effort)
	for _, c := range r.Checks {
		if c.Passed || c.Recommendation == "" {
			continue
		}
		impact := "Low"
		gain := 3
		switch c.Severity {
		case "critical":
			impact = "Critical"
			gain = 15
		case "high":
			impact = "High"
			gain = 8
		case "medium":
			impact = "Medium"
			gain = 5
		}
		s.QuickWins = append(s.QuickWins, QuickWin{
			Action:    c.Recommendation,
			Impact:    impact,
			Effort:    "Low", // heuristic v0.1
			ScoreGain: gain,
		})
		if len(s.QuickWins) >= 5 {
			break
		}
	}

	return s
}
