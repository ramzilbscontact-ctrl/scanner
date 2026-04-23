package collectors

import (
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

func collectWindows(d *Data, verbose bool) error {
	d.Platform = "Windows"

	// OS version via PowerShell
	if out, err := runPS("(Get-CimInstance Win32_OperatingSystem).Version"); err == nil {
		d.OSVersion = strings.TrimSpace(out)
	}

	d.Accounts = parseWindowsAccounts()
	d.Services = parseWindowsServices()
	d.Patches = parseWindowsPatches()
	d.Firewall = parseWindowsFirewall()
	d.Disk = parseWindowsDisk()
	d.Backup = parseWindowsBackup()

	return nil
}

// runPS executes a PowerShell snippet and returns stdout.
func runPS(script string) (string, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	return string(out), err
}

func parseWindowsAccounts() []Account {
	var accounts []Account
	// Get-LocalUser returns JSON when piped to ConvertTo-Json
	ps := `Get-LocalUser | Select-Object Name, Enabled, @{N='IsAdmin';E={(Get-LocalGroupMember -Group Administrators -ErrorAction SilentlyContinue | Where-Object {$_.Name -match $_.Name}).Count -gt 0}} | ConvertTo-Json -Compress`
	out, err := runPS(ps)
	if err != nil {
		return accounts
	}

	// Output can be a single object (not an array) if only 1 user
	raw := strings.TrimSpace(out)
	if raw == "" {
		return accounts
	}
	if !strings.HasPrefix(raw, "[") {
		raw = "[" + raw + "]"
	}

	type winAccount struct {
		Name    string `json:"Name"`
		Enabled bool   `json:"Enabled"`
		IsAdmin bool   `json:"IsAdmin"`
	}
	var wa []winAccount
	if err := json.Unmarshal([]byte(raw), &wa); err != nil {
		return accounts
	}

	for _, a := range wa {
		accounts = append(accounts, Account{
			Name:     a.Name,
			IsAdmin:  a.IsAdmin,
			Disabled: !a.Enabled,
			// MFA detection requires AAD/Entra context — not covered in v0.1 local scan
			MFAEnabled: false,
		})
	}
	return accounts
}

func parseWindowsServices() []Service {
	var svcs []Service
	relevant := []string{"WinDefend", "MpsSvc", "BITS", "wuauserv", "BFE",
		"EventLog", "Dnscache", "LanmanServer", "LanmanWorkstation",
		"Spooler", "RemoteRegistry", "SSDPSRV", "SharedAccess"}

	for _, name := range relevant {
		out, err := runPS("(Get-Service -Name " + name + " -ErrorAction SilentlyContinue).Status")
		if err != nil || strings.TrimSpace(out) == "" {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(out))
		svcs = append(svcs, Service{
			Name:   name,
			Status: status,
		})
	}
	return svcs
}

func parseWindowsPatches() PatchInfo {
	// Look for the most recent hotfix install date
	ps := `(Get-HotFix | Sort-Object InstalledOn -Descending | Select-Object -First 1).InstalledOn.ToString('yyyy-MM-ddTHH:mm:ssZ')`
	out, err := runPS(ps)
	if err != nil {
		return PatchInfo{DaysSincePatch: 999}
	}
	dateStr := strings.TrimSpace(out)
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return PatchInfo{DaysSincePatch: 999}
	}
	return PatchInfo{
		LastUpdate:     t,
		DaysSincePatch: int(time.Since(t).Hours() / 24),
	}
}

func parseWindowsFirewall() FirewallInfo {
	ps := `(Get-NetFirewallProfile | Where-Object Enabled -eq $true).Count`
	out, err := runPS(ps)
	if err != nil {
		return FirewallInfo{}
	}
	count := strings.TrimSpace(out)
	return FirewallInfo{
		Enabled: count != "0" && count != "",
		Profile: "Windows Defender Firewall",
	}
}

func parseWindowsDisk() DiskInfo {
	ps := `(Get-BitLockerVolume -MountPoint C: -ErrorAction SilentlyContinue).VolumeStatus`
	out, err := runPS(ps)
	if err != nil {
		return DiskInfo{}
	}
	status := strings.TrimSpace(out)
	return DiskInfo{
		Encrypted:      status == "FullyEncrypted",
		EncryptionType: "BitLocker",
	}
}

func parseWindowsBackup() BackupInfo {
	// Windows Backup is mostly deprecated; check for File History or 3rd-party
	ps := `(Get-ScheduledTask -TaskPath '\Microsoft\Windows\FileHistory\' -ErrorAction SilentlyContinue).State`
	out, _ := runPS(ps)
	if strings.Contains(out, "Ready") || strings.Contains(out, "Running") {
		return BackupInfo{Configured: true, Provider: "File History"}
	}
	return BackupInfo{Configured: false}
}
