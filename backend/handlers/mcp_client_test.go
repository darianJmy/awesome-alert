package handlers

import (
	"aiops-backend/models"
	"testing"
	"time"
)

func TestMCPCompatibilityStatus(t *testing.T) {
	now := time.Now()
	for _, test := range []struct {
		name, protocol, want string
		lastSeen             *time.Time
	}{
		{name: "never connected", want: "never_connected"},
		{name: "current protocol", protocol: currentMCPProtocolVersion, lastSeen: &now, want: "compatible"},
		{name: "supported older protocol", protocol: "2025-06-18", lastSeen: &now, want: "compatible"},
		{name: "unknown protocol", protocol: "2024-11-05", lastSeen: &now, want: "incompatible"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := mcpCompatibilityStatus(models.MCPClient{ProtocolVersion: test.protocol, LastSeenAt: test.lastSeen})
			if got != test.want {
				t.Fatalf("compatibility = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeMCPClientReq(t *testing.T) {
	req := normalizeMCPClientReq(mcpClientReq{Name: "  laptop  ", ClientType: " Codex ", APITokenID: " token ", Status: 9})
	if req.Name != "laptop" || req.ClientType != "codex" || req.APITokenID != "token" || req.Status != 1 {
		t.Fatalf("unexpected normalized request: %#v", req)
	}
}
