package handlers

import (
	"testing"
	"time"
)

func TestAPITokenExpiresAtSupportsOnlyPresetDurations(t *testing.T) {
	now := time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC)
	for _, days := range []int{7, 30, 60, 90} {
		expiresAt, ok := apiTokenExpiresAt(days, now)
		if !ok || !expiresAt.Equal(now.AddDate(0, 0, days)) {
			t.Fatalf("days=%d expiresAt=%v ok=%v", days, expiresAt, ok)
		}
	}
	permanent, ok := apiTokenExpiresAt(0, now)
	if !ok || permanent.Year() != permanentAPITokenYear {
		t.Fatalf("permanent expiresAt=%v ok=%v", permanent, ok)
	}
	for _, days := range []int{-1, 1, 365} {
		if _, ok := apiTokenExpiresAt(days, now); ok {
			t.Fatalf("unexpected supported duration: %d", days)
		}
	}
}
