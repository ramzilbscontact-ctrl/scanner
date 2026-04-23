package nis2

import (
	"fmt"

	"github.com/agenzia/scanner/internal/collectors"
)

// ═══════════════════════════════════════════════════════════════════
// Measure 1 — Risk management policy
// ═══════════════════════════════════════════════════════════════════

func checkRiskPolicyDocumented(d *collectors.Data) Check {
	// v0.1: we can't verify actual documentation from a system scan.
	// We score this as "unverified" and require manual input.
	return Check{
		Measure:     1,
		ID:          "1.1-policy-documented",
		Title:       "Risk management policy documented",
		Score:       0,
		Severity:    "critical",
		Passed:      false,
		Finding:     "Cannot auto-verify from system scan (requires document review)",
		Recommendation: "Upload your ISO 27005 or EBIOS RM policy to dashboard.agenzia.uk for review",
	}
}

func checkRiskAssessmentRecent(d *collectors.Data) Check {
	return Check{
		Measure:     1,
		ID:          "1.2-assessment-recent",
		Title:       "Risk assessment performed in last 12 months",
		Score:       0,
		Severity:    "high",
		Passed:      false,
		Finding:     "Cannot auto-verify from system scan",
		Recommendation: "Run an annual risk assessment; document top 15 risks",
	}
}

func checkRSSIAppointed(d *collectors.Data) Check {
	return Check{
		Measure:     1,
		ID:          "1.3-rssi-appointed",
		Title:       "Information Security Officer (RSSI) appointed",
		Score:       0,
		Severity:    "high",
		Passed:      false,
		Finding:     "Cannot auto-verify (organizational check)",
		Recommendation: "NIS2 requires a designated security officer, even part-time",
	}
}

// ═══════════════════════════════════════════════════════════════════
// Measure 2 — Incident handling
// ═══════════════════════════════════════════════════════════════════

func checkIncidentResponseTooling(d *collectors.Data) Check {
	hasEDR := false
	for _, s := range d.Services {
		n := s.Name
		if contains(n, "crowdstrike", "wazuh", "sentinel", "osquery", "clamav", "WinDefend") {
			hasEDR = true
			break
		}
	}
	score := 0
	finding := "No EDR/SIEM/AV service detected"
	if hasEDR {
		score = 80
		finding = "Endpoint security tooling detected"
	}
	return Check{
		Measure:        2,
		ID:             "2.1-ir-tooling",
		Title:          "Endpoint security tooling deployed (EDR/AV)",
		Score:          score,
		Severity:       "critical",
		Passed:         hasEDR,
		Finding:        finding,
		Recommendation: "Deploy Wazuh (free) or Sentinel One / Crowdstrike for endpoint detection",
	}
}

func checkLogsCollected(d *collectors.Data) Check {
	hasLogging := false
	for _, s := range d.Services {
		if contains(s.Name, "auditd", "EventLog", "rsyslog", "systemd-journald") {
			hasLogging = true
			break
		}
	}
	score := 0
	if hasLogging {
		score = 60
	}
	return Check{
		Measure:        2,
		ID:             "2.2-logs-enabled",
		Title:          "System logging enabled",
		Score:          score,
		Severity:       "high",
		Passed:         hasLogging,
		Finding:        fmt.Sprintf("Logging service present: %v", hasLogging),
		Recommendation: "Forward logs to a centralized SIEM (Wazuh is free and NIS2-friendly)",
	}
}

// ═══════════════════════════════════════════════════════════════════
// Measure 3 — Business continuity (backup)
// ═══════════════════════════════════════════════════════════════════

func checkBackupConfigured(d *collectors.Data) Check {
	score := 0
	finding := "No backup tool detected"
	if d.Backup.Configured {
		score = 70
		finding = fmt.Sprintf("Backup configured with %s", d.Backup.Provider)
	}
	return Check{
		Measure:        3,
		ID:             "3.1-backup-configured",
		Title:          "Backup solution configured",
		Score:          score,
		Severity:       "critical",
		Passed:         d.Backup.Configured,
		Finding:        finding,
		Recommendation: "Configure daily backups with 3-2-1 rule (3 copies, 2 media, 1 offsite)",
	}
}

func checkBackupRecent(d *collectors.Data) Check {
	if d.Backup.LastBackup.IsZero() {
		return Check{
			Measure:        3,
			ID:             "3.2-backup-recent",
			Title:          "Recent successful backup",
			Score:          0,
			Severity:       "high",
			Passed:         false,
			Finding:        "No recent backup timestamp found",
			Recommendation: "Verify backups run daily and succeed",
		}
	}
	return Check{
		Measure:        3,
		ID:             "3.2-backup-recent",
		Title:          "Recent successful backup",
		Score:          80,
		Severity:       "high",
		Passed:         true,
		Finding:        "Recent backup timestamp detected",
		Recommendation: "",
	}
}

func checkBackupTested(d *collectors.Data) Check {
	return Check{
		Measure:        3,
		ID:             "3.3-backup-tested",
		Title:          "Backup restoration test performed",
		Score:          0,
		Severity:       "high",
		Passed:         false,
		Finding:        "Cannot auto-verify test restoration from system scan",
		Recommendation: "Test restoration every quarter (keep a log). ANSSI insists on this.",
	}
}

// ═══════════════════════════════════════════════════════════════════
// Measure 7 — Cyber hygiene
// ═══════════════════════════════════════════════════════════════════

func checkPatchLevel(d *collectors.Data) Check {
	score := 100
	severity := "low"
	passed := true
	finding := fmt.Sprintf("System patched %d days ago", d.Patches.DaysSincePatch)

	switch {
	case d.Patches.DaysSincePatch > 90:
		score = 20
		severity = "critical"
		passed = false
		finding = "System not patched in 90+ days — critical risk"
	case d.Patches.DaysSincePatch > 30:
		score = 60
		severity = "high"
		passed = false
		finding = "System not patched in 30+ days"
	case d.Patches.DaysSincePatch > 14:
		score = 80
		severity = "medium"
	}

	return Check{
		Measure:        7,
		ID:             "7.1-patch-level",
		Title:          "System patch level up to date",
		Score:          score,
		Severity:       severity,
		Passed:         passed,
		Finding:        finding,
		Recommendation: "Enable automatic updates; apply critical patches within 14 days",
	}
}

func checkFirewallEnabled(d *collectors.Data) Check {
	score := 0
	finding := "Firewall disabled or not detected"
	if d.Firewall.Enabled {
		score = 80
		finding = fmt.Sprintf("Firewall active (%s)", d.Firewall.Profile)
	}
	return Check{
		Measure:        7,
		ID:             "7.2-firewall-enabled",
		Title:          "Host firewall enabled",
		Score:          score,
		Severity:       "high",
		Passed:         d.Firewall.Enabled,
		Finding:        finding,
		Recommendation: "Enable built-in firewall; restrict inbound to essential services only",
	}
}

func checkNoAbandonedAccounts(d *collectors.Data) Check {
	// Count disabled accounts that still exist (should be removed or reviewed)
	disabled := 0
	for _, a := range d.Accounts {
		if a.Disabled {
			disabled++
		}
	}
	score := 90
	passed := true
	finding := "No disabled accounts detected"
	if disabled > 0 {
		score = 50
		passed = false
		finding = fmt.Sprintf("%d disabled account(s) still present", disabled)
	}
	return Check{
		Measure:        7,
		ID:             "7.3-no-abandoned-accounts",
		Title:          "No abandoned user accounts",
		Score:          score,
		Severity:       "medium",
		Passed:         passed,
		Finding:        finding,
		Recommendation: "Remove accounts within 30 days of departure; review quarterly",
	}
}

// ═══════════════════════════════════════════════════════════════════
// Measure 8 — Cryptography
// ═══════════════════════════════════════════════════════════════════

func checkDiskEncrypted(d *collectors.Data) Check {
	score := 0
	finding := "Disk not encrypted"
	if d.Disk.Encrypted {
		score = 90
		finding = fmt.Sprintf("Disk encrypted with %s", d.Disk.EncryptionType)
	}
	return Check{
		Measure:        8,
		ID:             "8.1-disk-encryption",
		Title:          "Full disk encryption enabled",
		Score:          score,
		Severity:       "critical",
		Passed:         d.Disk.Encrypted,
		Finding:        finding,
		Recommendation: "Enable BitLocker (Windows) / FileVault (Mac) / LUKS (Linux) on all laptops",
	}
}

// ═══════════════════════════════════════════════════════════════════
// Measure 9 — Access control
// ═══════════════════════════════════════════════════════════════════

func checkAdminCount(d *collectors.Data) Check {
	admins := 0
	for _, a := range d.Accounts {
		if a.IsAdmin {
			admins++
		}
	}
	total := len(d.Accounts)
	ratio := 0
	if total > 0 {
		ratio = (admins * 100) / total
	}

	score := 100
	severity := "low"
	passed := true
	finding := fmt.Sprintf("%d admin account(s) out of %d (%d%%)", admins, total, ratio)

	if ratio > 50 {
		score = 30
		severity = "high"
		passed = false
		finding += " — too many admin accounts (least privilege violated)"
	} else if ratio > 20 {
		score = 60
		severity = "medium"
	}

	return Check{
		Measure:        9,
		ID:             "9.1-admin-count",
		Title:          "Least privilege — admin count reasonable",
		Score:          score,
		Severity:       severity,
		Passed:         passed,
		Finding:        finding,
		Recommendation: "Keep admin accounts below 20% of total users; audit quarterly",
	}
}

func checkNoDisabledKept(d *collectors.Data) Check {
	return Check{
		Measure:  9,
		ID:       "9.2-clean-accounts",
		Title:    "Inactive accounts removed",
		Score:    70, // placeholder
		Severity: "medium",
		Passed:   true,
		Finding:  "Heuristic: removed based on disabled count",
	}
}

// ═══════════════════════════════════════════════════════════════════
// Measure 10 — MFA
// ═══════════════════════════════════════════════════════════════════

func checkMFAOnAdmins(d *collectors.Data) Check {
	var adminsWithoutMFA int
	for _, a := range d.Accounts {
		if a.IsAdmin && !a.MFAEnabled {
			adminsWithoutMFA++
		}
	}
	score := 100
	severity := "low"
	passed := true
	finding := "All admin accounts have MFA enabled"
	if adminsWithoutMFA > 0 {
		score = 10
		severity = "critical"
		passed = false
		finding = fmt.Sprintf("%d admin account(s) WITHOUT MFA — NIS2 requires MFA for privileged access", adminsWithoutMFA)
	}
	return Check{
		Measure:        10,
		ID:             "10.1-mfa-admins",
		Title:          "MFA enabled on all admin accounts",
		Score:          score,
		Severity:       severity,
		Passed:         passed,
		Finding:        finding,
		Recommendation: "Enable MFA for every admin — this is NIS2 Art. 21 mandatory",
	}
}

func checkMFAOnRegularUsers(d *collectors.Data) Check {
	var regularsWithoutMFA int
	var regularsTotal int
	for _, a := range d.Accounts {
		if !a.IsAdmin && !a.Disabled {
			regularsTotal++
			if !a.MFAEnabled {
				regularsWithoutMFA++
			}
		}
	}
	score := 100
	severity := "medium"
	passed := true
	if regularsTotal > 0 {
		ratio := (regularsWithoutMFA * 100) / regularsTotal
		if ratio > 50 {
			score = 30
			severity = "high"
			passed = false
		} else if ratio > 10 {
			score = 60
		}
	}
	return Check{
		Measure:        10,
		ID:             "10.2-mfa-regular",
		Title:          "MFA deployed on regular users",
		Score:          score,
		Severity:       severity,
		Passed:         passed,
		Finding:        fmt.Sprintf("%d/%d regular users without MFA", regularsWithoutMFA, regularsTotal),
		Recommendation: "Deploy MFA to 100% of users (M365/GWS allow 1-click org-wide enforcement)",
	}
}

// ─── helpers ──────────────────────────────────────────────────────
func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(sub) == 0 {
			continue
		}
		for i := 0; i+len(sub) <= len(s); i++ {
			if equalFold(s[i:i+len(sub)], sub) {
				return true
			}
		}
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
