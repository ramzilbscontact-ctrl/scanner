// Package nis2 evaluates collected system data against the 10 NIS2 Article 21 measures.
// Each measure has multiple checks. Each check returns a 0-100 score + a verdict.
package nis2

import (
	"github.com/agenzia/scanner/internal/collectors"
)

// Check is a single compliance verification.
type Check struct {
	Measure     int    `json:"measure"`     // 1-10, NIS2 Art. 21
	ID          string `json:"id"`          // e.g., "10.1-mfa-admins"
	Title       string `json:"title"`       // Human-readable
	Score       int    `json:"score"`       // 0-100
	Severity    string `json:"severity"`    // critical | high | medium | low
	Passed      bool   `json:"passed"`
	Finding     string `json:"finding"`     // What we found
	Recommendation string `json:"recommendation"`
}

// Results holds all checks + a per-measure aggregate.
type Results struct {
	Checks         []Check      `json:"checks"`
	MeasureScores  [11]int      `json:"measure_scores"` // index 0 unused, 1-10 are real
	CriticalGaps   []string     `json:"critical_gaps"`
}

// RunAllChecks runs the full NIS2 rule engine on collected data.
func RunAllChecks(d *collectors.Data) *Results {
	r := &Results{}

	// Measure 1: Risk management policy (3 checks)
	r.Checks = append(r.Checks, checkRiskPolicyDocumented(d))
	r.Checks = append(r.Checks, checkRiskAssessmentRecent(d))
	r.Checks = append(r.Checks, checkRSSIAppointed(d))

	// Measure 2: Incident handling (2 checks)
	r.Checks = append(r.Checks, checkIncidentResponseTooling(d))
	r.Checks = append(r.Checks, checkLogsCollected(d))

	// Measure 3: Business continuity (3 checks)
	r.Checks = append(r.Checks, checkBackupConfigured(d))
	r.Checks = append(r.Checks, checkBackupRecent(d))
	r.Checks = append(r.Checks, checkBackupTested(d))

	// Measure 7: Cyber hygiene (3 checks)
	r.Checks = append(r.Checks, checkPatchLevel(d))
	r.Checks = append(r.Checks, checkFirewallEnabled(d))
	r.Checks = append(r.Checks, checkNoAbandonedAccounts(d))

	// Measure 8: Cryptography (1 check)
	r.Checks = append(r.Checks, checkDiskEncrypted(d))

	// Measure 9: Access control (2 checks)
	r.Checks = append(r.Checks, checkAdminCount(d))
	r.Checks = append(r.Checks, checkNoDisabledKept(d))

	// Measure 10: MFA (2 checks)
	r.Checks = append(r.Checks, checkMFAOnAdmins(d))
	r.Checks = append(r.Checks, checkMFAOnRegularUsers(d))

	// Aggregate per measure
	r.MeasureScores = aggregateByMeasure(r.Checks)

	// Collect critical gaps (severity=critical AND !passed)
	for _, c := range r.Checks {
		if c.Severity == "critical" && !c.Passed {
			r.CriticalGaps = append(r.CriticalGaps, c.Title)
		}
	}

	return r
}

func aggregateByMeasure(checks []Check) [11]int {
	var scores [11]int
	counts := [11]int{}
	sums := [11]int{}
	for _, c := range checks {
		if c.Measure < 1 || c.Measure > 10 {
			continue
		}
		counts[c.Measure]++
		sums[c.Measure] += c.Score
	}
	for i := 1; i <= 10; i++ {
		if counts[i] > 0 {
			scores[i] = sums[i] / counts[i]
		} else {
			scores[i] = -1 // not evaluated
		}
	}
	return scores
}
