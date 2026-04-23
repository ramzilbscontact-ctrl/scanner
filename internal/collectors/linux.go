package collectors

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func collectLinux(d *Data, verbose bool) error {
	// OS version via /etc/os-release
	d.OSVersion, d.Platform = parseOSRelease()

	// Local accounts from /etc/passwd (UID >= 1000 typically = real users)
	d.Accounts = parseLinuxAccounts()

	// Running services via systemctl
	d.Services = parseLinuxServices()

	// Patch info via package manager last-update timestamp
	d.Patches = parseLinuxPatches()

	// Firewall via ufw status or iptables
	d.Firewall = parseLinuxFirewall()

	// Disk encryption via lsblk
	d.Disk = parseLinuxDisk()

	// Backup: heuristic — look for cron jobs mentioning backup/borg/restic
	d.Backup = parseLinuxBackup()

	return nil
}

func parseOSRelease() (version, platform string) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return "unknown", "linux"
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			platform = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		}
	}
	if platform == "" {
		platform = "Linux (unknown distro)"
	}
	return
}

func parseLinuxAccounts() []Account {
	var accounts []Account
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return accounts
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) < 7 {
			continue
		}
		uid, _ := strconv.Atoi(parts[2])
		// Skip system accounts (UID < 1000, except root)
		if uid != 0 && uid < 1000 {
			continue
		}
		accounts = append(accounts, Account{
			Name:       parts[0],
			IsAdmin:    uid == 0 || isInSudoGroup(parts[0]),
			MFAEnabled: checkPAMMFA(parts[0]), // heuristic
			Disabled:   strings.HasPrefix(parts[1], "!") || parts[1] == "*",
		})
	}
	return accounts
}

func isInSudoGroup(user string) bool {
	out, err := exec.Command("groups", user).Output()
	if err != nil {
		return false
	}
	s := string(out)
	return strings.Contains(s, "sudo") || strings.Contains(s, "wheel") || strings.Contains(s, "admin")
}

func checkPAMMFA(user string) bool {
	// Very rough heuristic: check if pam_google_authenticator or pam_u2f is installed.
	// A real check would parse /etc/pam.d/sshd and friends.
	if _, err := os.Stat("/etc/pam.d/sshd"); err != nil {
		return false
	}
	data, _ := os.ReadFile("/etc/pam.d/sshd")
	content := string(data)
	return strings.Contains(content, "pam_google_authenticator") ||
		strings.Contains(content, "pam_u2f") ||
		strings.Contains(content, "pam_duo")
}

func parseLinuxServices() []Service {
	var svcs []Service
	out, err := exec.Command("systemctl", "list-unit-files", "--type=service", "--no-pager", "--no-legend").Output()
	if err != nil {
		return svcs
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimSuffix(fields[0], ".service")
		// Only keep relevant services (skip 100s of stock systemd units)
		if !isRelevantService(name) {
			continue
		}
		status, _ := exec.Command("systemctl", "is-active", name).Output()
		svcs = append(svcs, Service{
			Name:    name,
			Status:  strings.TrimSpace(string(status)),
			AutoRun: fields[1] == "enabled",
		})
	}
	return svcs
}

func isRelevantService(name string) bool {
	relevant := []string{"ssh", "firewalld", "ufw", "fail2ban", "auditd", "snort", "suricata",
		"clamav", "osquery", "wazuh", "borg", "restic", "bacula", "cron", "crowdsec",
		"docker", "containerd", "postgresql", "mysql", "redis"}
	for _, r := range relevant {
		if strings.Contains(name, r) {
			return true
		}
	}
	return false
}

func parseLinuxPatches() PatchInfo {
	// Check last modified time of package db
	// apt: /var/lib/apt/lists/
	// dnf/yum: /var/cache/dnf or /var/cache/yum
	// pacman: /var/lib/pacman/local/
	paths := []string{
		"/var/lib/apt/lists/partial",
		"/var/lib/apt/periodic/update-success-stamp",
		"/var/cache/dnf",
		"/var/lib/pacman/sync",
	}
	var last time.Time
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil {
			if info.ModTime().After(last) {
				last = info.ModTime()
			}
		}
	}
	days := 999
	if !last.IsZero() {
		days = int(time.Since(last).Hours() / 24)
	}
	return PatchInfo{
		LastUpdate:     last,
		DaysSincePatch: days,
	}
}

func parseLinuxFirewall() FirewallInfo {
	// Try ufw first
	if out, err := exec.Command("ufw", "status").Output(); err == nil {
		enabled := strings.Contains(string(out), "Status: active")
		return FirewallInfo{Enabled: enabled, Profile: "ufw"}
	}
	// Try firewalld
	if out, err := exec.Command("systemctl", "is-active", "firewalld").Output(); err == nil {
		enabled := strings.TrimSpace(string(out)) == "active"
		return FirewallInfo{Enabled: enabled, Profile: "firewalld"}
	}
	// Fall back to iptables
	if out, err := exec.Command("iptables", "-L", "-n").Output(); err == nil {
		return FirewallInfo{
			Enabled:     len(out) > 100, // very rough check
			Profile:     "iptables",
			OpenInbound: strings.Count(string(out), "ACCEPT"),
		}
	}
	return FirewallInfo{Enabled: false}
}

func parseLinuxDisk() DiskInfo {
	// Look for LUKS-encrypted volumes
	out, err := exec.Command("lsblk", "-o", "NAME,TYPE", "-n").Output()
	if err != nil {
		return DiskInfo{}
	}
	if strings.Contains(string(out), "crypt") {
		return DiskInfo{Encrypted: true, EncryptionType: "LUKS"}
	}
	return DiskInfo{Encrypted: false}
}

func parseLinuxBackup() BackupInfo {
	// Heuristic: look for known backup tools in /etc/cron.*
	tools := []string{"borg", "restic", "duplicity", "rclone", "rsnapshot", "bacula"}
	cronDirs := []string{"/etc/cron.daily", "/etc/cron.weekly", "/etc/cron.hourly"}

	for _, dir := range cronDirs {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			for _, t := range tools {
				if strings.Contains(strings.ToLower(e.Name()), t) {
					return BackupInfo{Configured: true, Provider: t}
				}
			}
		}
	}
	return BackupInfo{Configured: false}
}

// Avoid unused import warnings
var _ = fmt.Sprintf
