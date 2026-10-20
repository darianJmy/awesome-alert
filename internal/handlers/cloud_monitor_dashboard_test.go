package handlers

import (
	"net/url"
	"strings"
	"testing"
)

func TestValidateCloudMonitorWorkspaceReq(t *testing.T) {
	valid := cloudMonitorWorkspaceReq{Name: "生产环境监控", Workspace: "production-observe", Region: "cn-shanghai", AccessKeyID: "id", AccessKeySecret: "secret"}
	if message := validateCloudMonitorWorkspaceReq(valid, true); message != "" {
		t.Fatalf("valid request rejected: %s", message)
	}
	invalid := valid
	invalid.Name = " "
	if message := validateCloudMonitorWorkspaceReq(invalid, true); message == "" {
		t.Fatal("blank name should be rejected")
	}
	invalid = valid
	invalid.Region = "unknown-region"
	if message := validateCloudMonitorWorkspaceReq(invalid, true); message == "" {
		t.Fatal("unsupported region should be rejected")
	}
	invalid = valid
	invalid.Workspace = "bad/workspace"
	if message := validateCloudMonitorWorkspaceReq(invalid, true); message == "" {
		t.Fatal("workspace containing a path separator should be rejected")
	}
}

func TestBuildCloudMonitorConsoleURL(t *testing.T) {
	value := buildCloudMonitorConsoleURL("cn-shanghai", "production", "ticket+/=")
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	if !strings.Contains(parsed.Path, "/region/cn-shanghai/workspace/production/") {
		t.Fatalf("unexpected path: %s", parsed.Path)
	}
	if parsed.Query().Get("sls_ticket") != "ticket+/=" {
		t.Fatalf("ticket was not encoded safely: %q", parsed.Query().Get("sls_ticket"))
	}
}
