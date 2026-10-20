package handlers

import (
	"testing"
	"time"
)

func TestCertificateDomainStatusUsesConfiguredExpiryDays(t *testing.T) {
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.Local)
	tests := []struct {
		name      string
		notAfter  time.Time
		alertDays int
		want      string
	}{
		{name: "outside window", notAfter: now.AddDate(0, 0, 8), alertDays: 7, want: "normal"},
		{name: "at alert boundary", notAfter: now.AddDate(0, 0, 7), alertDays: 7, want: "expiring"},
		{name: "inside alert window", notAfter: now.AddDate(0, 0, 2), alertDays: 7, want: "expiring"},
		{name: "expired", notAfter: now.Add(-time.Minute), alertDays: 7, want: "expired"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := certificateDomainStatus(test.notAfter, now, test.alertDays); got != test.want {
				t.Fatalf("certificateDomainStatus() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeCertificateDomainSeverity(t *testing.T) {
	for input, want := range map[string]string{"P0": "P0", "p1": "P1", " P2 ": "P2", "": "P2"} {
		if got := normalizeCertificateDomainSeverity(input); got != want {
			t.Fatalf("normalizeCertificateDomainSeverity(%q) = %q, want %q", input, got, want)
		}
	}
}
