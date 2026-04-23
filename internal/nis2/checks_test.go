package nis2

import (
	"testing"

	"github.com/agenzia/scanner/internal/collectors"
)

func TestCheckMFAOnAdmins(t *testing.T) {
	tests := []struct {
		name       string
		accounts   []collectors.Account
		wantPassed bool
		wantScore  int
	}{
		{
			name: "all admins have MFA",
			accounts: []collectors.Account{
				{Name: "alice", IsAdmin: true, MFAEnabled: true},
				{Name: "bob", IsAdmin: true, MFAEnabled: true},
			},
			wantPassed: true,
			wantScore:  100,
		},
		{
			name: "one admin without MFA",
			accounts: []collectors.Account{
				{Name: "alice", IsAdmin: true, MFAEnabled: true},
				{Name: "bob", IsAdmin: true, MFAEnabled: false},
			},
			wantPassed: false,
			wantScore:  10,
		},
		{
			name:       "no admins",
			accounts:   []collectors.Account{{Name: "alice", IsAdmin: false, MFAEnabled: true}},
			wantPassed: true,
			wantScore:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &collectors.Data{Accounts: tt.accounts}
			got := checkMFAOnAdmins(data)
			if got.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", got.Passed, tt.wantPassed)
			}
			if got.Score != tt.wantScore {
				t.Errorf("Score = %d, want %d", got.Score, tt.wantScore)
			}
		})
	}
}

func TestCheckPatchLevel(t *testing.T) {
	tests := []struct {
		name     string
		days     int
		wantPass bool
	}{
		{"up to date (5 days)", 5, true},
		{"slightly old (20 days)", 20, true},
		{"stale (45 days)", 45, false},
		{"critical (100 days)", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &collectors.Data{
				Patches: collectors.PatchInfo{DaysSincePatch: tt.days},
			}
			got := checkPatchLevel(data)
			if got.Passed != tt.wantPass {
				t.Errorf("Passed = %v, want %v (days=%d, score=%d)",
					got.Passed, tt.wantPass, tt.days, got.Score)
			}
		})
	}
}

func TestCheckAdminCount(t *testing.T) {
	tests := []struct {
		name     string
		accounts []collectors.Account
		wantPass bool
	}{
		{
			name: "healthy ratio (1 admin out of 10)",
			accounts: func() []collectors.Account {
				a := []collectors.Account{{Name: "admin", IsAdmin: true}}
				for i := 0; i < 9; i++ {
					a = append(a, collectors.Account{Name: "user"})
				}
				return a
			}(),
			wantPass: true,
		},
		{
			name: "too many admins (6 out of 10)",
			accounts: func() []collectors.Account {
				a := []collectors.Account{}
				for i := 0; i < 6; i++ {
					a = append(a, collectors.Account{Name: "admin", IsAdmin: true})
				}
				for i := 0; i < 4; i++ {
					a = append(a, collectors.Account{Name: "user"})
				}
				return a
			}(),
			wantPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &collectors.Data{Accounts: tt.accounts}
			got := checkAdminCount(data)
			if got.Passed != tt.wantPass {
				t.Errorf("Passed = %v, want %v (score=%d)", got.Passed, tt.wantPass, got.Score)
			}
		})
	}
}
