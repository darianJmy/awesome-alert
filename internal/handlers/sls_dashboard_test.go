package handlers

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildSLSConsoleURL(t *testing.T) {
	raw := buildSLSConsoleURL("demo project", "app/access", "cn-hangzhou", "ticket+value")
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	if !strings.Contains(parsed.EscapedPath(), "/project/demo%20project/logsearch/app%2Faccess") {
		t.Fatalf("path was not escaped: %s", parsed.EscapedPath())
	}
	query := parsed.Query()
	if query.Get("slsRegion") != "cn-hangzhou" || query.Get("sls_ticket") != "ticket+value" {
		t.Fatalf("unexpected query: %v", query)
	}
	if query.Get("hideTopbar") != "true" || query.Get("hideSidebar") != "true" || query.Get("isShare") != "true" {
		t.Fatalf("missing embed options: %v", query)
	}
}

func TestSLSRegionEndpoint(t *testing.T) {
	if got := slsRegionEndpoint("cn-hangzhou"); got != "cn-hangzhou.log.aliyuncs.com" {
		t.Fatalf("region endpoint = %q", got)
	}
}
