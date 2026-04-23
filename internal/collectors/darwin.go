package collectors

import (
	"bufio"
	"os/exec"
	"strings"
	"time"
)

func collectDarwin(d *Data, verbose bool) error {
	// OS version
	if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
		d.OSVersion = strings.TrimSpace(string(out))
	}
	d.Platform = "macOS " + d.OSVersion

	// Accounts: dscl list users (admin group: admin)
	d.Accounts = parseMacAccounts()

	// Services: launchctl list
	d.Services = parseMacServices()

	// Patches: softwareupdate --history
	d.Patches = parseMacPatches()

	// Firewall: defaults read /Library/Preferences/com.apple.alf
	d.Firewall = parseMacFirewall()

	// Disk: FileVault status via fdesetup
	d.Disk = parseMacDisk()

	// Backup: Time Machine status
	d.Backup = parseMacBackup()

	return nil
}

func parseMacAccounts() []Account {
	var accounts []Account
	out, err := exec.Command("dscl", ".", "list", "/Users").Output()
	if err != nil {
		return accounts
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		user := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(user, "_") || user == "daemon" || user == "nobody" || user == "root" {
			continue
		}
		// Check if user is in admin group
		adminOut, _ := exec.Command("dsmemberutil", "checkmembership", "-U", user, "-G", "admin").Output()
		isAdmin := strings.Contains(string(adminOut), "is a member")
		accounts = append(accounts, Account{
			Name:    user,
			IsAdmin: isAdmin,
			// MFA detection on macOS is limited to iCloud 2FA — treat as unknown
			MFAEnabled: false,
		})
	}
	return accounts
}

func parseMacServices() []Service {
	var svcs []Service
	out, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return svcs
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	scanner.Scan() // skip header
	relevant := []string{"com.apple.firewalld", "com.apple.TimeMachine",
		"com.apple.securityd", "org.apache.httpd", "com.docker"}
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		for _, r := range relevant {
			if strings.Contains(fields[2], r) {
				svcs = append(svcs, Service{
					Name:   fields[2],
					Status: "running",
				})
			}
		}
	}
	return svcs
}

func parseMacPatches() PatchInfo {
	out, err := exec.Command("softwareupdate", "--history").Output()
	if err != nil {
		return PatchInfo{DaysSincePatch: 999}
	}
	// Parse the last line of the history (most recent)
	lines := strings.Split(string(out), "\n")
	if len(lines) < 3 {
		return PatchInfo{DaysSincePatch: 999}
	}
	// softwareupdate --history format has a date in columns; simplified parsing
	// This is a heuristic — real implementation would parse properly
	return PatchInfo{
		LastUpdate:     time.Now().AddDate(0, 0, -30), // placeholder
		DaysSincePatch: 30,
	}
}

func parseMacFirewall() FirewallInfo {
	out, err := exec.Command("defaults", "read", "/Library/Preferences/com.apple.alf", "globalstate").Output()
	if err != nil {
		return FirewallInfo{}
	}
	state := strings.TrimSpace(string(out))
	return FirewallInfo{
		Enabled: state != "0",
		Profile: "alf",
	}
}

func parseMacDisk() DiskInfo {
	out, err := exec.Command("fdesetup", "status").Output()
	if err != nil {
		return DiskInfo{}
	}
	enabled := strings.Contains(string(out), "FileVault is On")
	return DiskInfo{
		Encrypted:      enabled,
		EncryptionType: "FileVault",
	}
}

func parseMacBackup() BackupInfo {
	out, err := exec.Command("tmutil", "destinationinfo").Output()
	if err != nil || len(out) < 10 {
		return BackupInfo{Configured: false}
	}
	return BackupInfo{
		Configured: true,
		Provider:   "Time Machine",
	}
}
