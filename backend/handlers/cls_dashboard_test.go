package handlers

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestValidateSLSDashboardProjectReqSupportsTencentCLS(t *testing.T) {
	valid := slsDashboardProjectReq{
		Provider:        "tencent_cls",
		Name:            "生产日志",
		Region:          "ap-shanghai",
		RoleArn:         "qcs::cam::uin/10001:roleName/CLSReadOnly",
		AccessKeyID:     "AKIDexample",
		AccessKeySecret: "secret",
	}
	if message := validateSLSDashboardProjectReq(valid, true); message != "" {
		t.Fatalf("valid Tencent CLS config rejected: %s", message)
	}

	missingRole := valid
	missingRole.RoleArn = ""
	if message := validateSLSDashboardProjectReq(missingRole, true); !strings.Contains(message, "RoleArn") {
		t.Fatalf("missing RoleArn error = %q", message)
	}
}

func TestBuildTencentCLSRoleLoginURL(t *testing.T) {
	credentials := tencentTemporaryCredentials{
		SecretID:  "TMP-ID",
		SecretKey: "TMP-KEY",
		Token:     "TMP-TOKEN",
	}
	raw := buildTencentCLSRoleLoginURL(credentials, "ap-shanghai", "topic-123", time.Unix(1_700_000_000, 0), 12345)
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "cloud.tencent.cn" || parsed.Path != "/login/roleAccessCallback" {
		t.Fatalf("unexpected callback URL: %s", raw)
	}
	query := parsed.Query()
	if query.Get("secretId") != "TMP-ID" || query.Get("token") != "TMP-TOKEN" {
		t.Fatalf("temporary credentials missing from URL: %s", raw)
	}
	if query.Get("timestamp") != "1700000000" || query.Get("nonce") != "12345" || query.Get("algorithm") != "sha1" {
		t.Fatalf("role login metadata is wrong: %s", raw)
	}
	if query.Get("signature") == "" {
		t.Fatalf("signature is missing: %s", raw)
	}
	target, err := url.Parse(query.Get("s_url"))
	if err != nil {
		t.Fatal(err)
	}
	if target.Host != "console.cloud.tencent.cn" || target.Path != "/cls/search" {
		t.Fatalf("unexpected CLS target URL: %s", target.String())
	}
	if target.Query().Get("region") != "ap-shanghai" || target.Query().Get("topic_id") != "topic-123" {
		t.Fatalf("CLS target parameters are wrong: %s", target.String())
	}
	for _, key := range []string{"hideTopNav", "hideLeftNav", "hideConfigMenu", "hideWidget"} {
		if target.Query().Get(key) != "true" {
			t.Fatalf("%s was not hidden in target URL: %s", key, target.String())
		}
	}
}
