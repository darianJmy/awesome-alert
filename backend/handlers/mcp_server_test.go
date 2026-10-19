package handlers

import "testing"

func TestNegotiateMCPProtocol(t *testing.T) {
	for requested, want := range map[string]string{
		"2025-03-26": "2025-03-26",
		"2025-06-18": "2025-06-18",
		"2025-11-25": "2025-11-25",
		"unknown":    currentMCPProtocolVersion,
	} {
		if got := negotiateMCPProtocol(requested); got != want {
			t.Fatalf("negotiateMCPProtocol(%q) = %q, want %q", requested, got, want)
		}
	}
}

func TestMCPToolsHaveUniqueNamesAndPermissions(t *testing.T) {
	seen := map[string]bool{}
	for _, tool := range mcpTools {
		if tool.Name == "" || tool.Description == "" || tool.Permission == "" {
			t.Fatalf("incomplete tool definition: %#v", tool)
		}
		if seen[tool.Name] {
			t.Fatalf("duplicate tool name %q", tool.Name)
		}
		seen[tool.Name] = true
		if tool.InputSchema["type"] != "object" {
			t.Fatalf("tool %q does not expose an object input schema", tool.Name)
		}
	}
}

func TestMCPPaginationBounds(t *testing.T) {
	page, pageSize := mcpPagination(map[string]interface{}{"page": float64(-1), "page_size": float64(1000)})
	if page != 1 || pageSize != 100 {
		t.Fatalf("pagination = %d/%d, want 1/100", page, pageSize)
	}
}
