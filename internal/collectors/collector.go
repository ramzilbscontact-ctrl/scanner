// Package collectors gathers system information from the host.
// Each OS has its own collector; the Run function picks the right one at runtime.
package collectors

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// Data holds everything collected about the host.
// It is passed to the NIS2 rule engine for evaluation.
type Data struct {
	Hostname     string            `json:"hostname"`
	OS           string            `json:"os"`
	OSVersion    string            `json:"os_version"`
	Architecture string            `json:"architecture"`
	Platform     string            `json:"platform"`
	CollectedAt  time.Time         `json:"collected_at"`
	Accounts     []Account         `json:"accounts"`
	Services     []Service         `json:"services"`
	Patches      PatchInfo         `json:"patches"`
	Firewall     FirewallInfo      `json:"firewall"`
	Disk         DiskInfo          `json:"disk"`
	Backup       BackupInfo        `json:"backup"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// Account describes a local or domain user.
type Account struct {
	Name       string `json:"name"`
	IsAdmin    bool   `json:"is_admin"`
	MFAEnabled bool   `json:"mfa_enabled"`
	LastLogin  string `json:"last_login,omitempty"`
	Disabled   bool   `json:"disabled"`
}

// Service is a running system service (daemon on Linux, service on Windows, etc.).
type Service struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // running | stopped | disabled
	AutoRun bool   `json:"auto_run"`
}

type PatchInfo struct {
	LastUpdate   time.Time `json:"last_update"`
	DaysSincePatch int     `json:"days_since_patch"`
	MissingCritical int    `json:"missing_critical"`
}

type FirewallInfo struct {
	Enabled     bool   `json:"enabled"`
	Profile     string `json:"profile,omitempty"` // domain | private | public
	OpenInbound int    `json:"open_inbound_rules"`
}

type DiskInfo struct {
	Encrypted      bool   `json:"encrypted"`
	EncryptionType string `json:"encryption_type,omitempty"` // BitLocker | LUKS | FileVault | none
}

type BackupInfo struct {
	Configured    bool      `json:"configured"`
	Provider      string    `json:"provider,omitempty"`
	LastBackup    time.Time `json:"last_backup,omitempty"`
	DaysSinceTest int       `json:"days_since_test"`
}

// Count returns a rough count of collected data points (used for UI progress).
func (d *Data) Count() int {
	n := 6 // base fields
	n += len(d.Accounts)
	n += len(d.Services)
	if d.Patches.LastUpdate.IsZero() {
		n++
	}
	return n
}

// Run picks the right platform-specific collector and runs it.
func Run(verbose bool) (*Data, error) {
	host, _ := os.Hostname()
	d := &Data{
		Hostname:     host,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CollectedAt:  time.Now().UTC(),
		Extra:        map[string]string{},
	}

	var err error
	switch runtime.GOOS {
	case "linux":
		err = collectLinux(d, verbose)
	case "darwin":
		err = collectDarwin(d, verbose)
	case "windows":
		err = collectWindows(d, verbose)
	default:
		err = fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return d, err
}
