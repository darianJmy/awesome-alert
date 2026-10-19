package handlers

import (
	"aiops-backend/models"
	"strings"
	"testing"
)

func TestBuildAuthSettingRejectsMultipleExternalProviders(t *testing.T) {
	_, err := buildAuthSettingForTest(saveAuthSettingReq{
		LDAPEnabled: true,
		OIDCEnabled: true,
	}, models.AuthSetting{})
	if err == nil || !strings.Contains(err.Error(), "只能启用一种") {
		t.Fatalf("expected mutually exclusive provider error, got %v", err)
	}
}

func TestBuildAuthSettingParsesLDAPYAML(t *testing.T) {
	item, err := buildAuthSettingForTest(saveAuthSettingReq{
		LDAPEnabled: true,
		LDAPConfig: `connection:
  url: ldaps://ldap.example.com:636
  bind_dn: ""
  bind_password: ""
user_search:
  base_dn: dc=example,dc=com
  filter: (uid={username})
attributes:
  username: uid
  display_name: cn
  email: mail
`,
	}, models.AuthSetting{})
	if err != nil {
		t.Fatalf("expected valid LDAP YAML, got %v", err)
	}
	if item.LDAPURL != "ldaps://ldap.example.com:636" || item.LDAPBaseDN != "dc=example,dc=com" {
		t.Fatalf("LDAP YAML was not applied: %#v", item)
	}
}

func TestBuildAuthSettingStillAcceptsFlatLDAPYAML(t *testing.T) {
	item, err := buildAuthSettingForTest(saveAuthSettingReq{
		LDAPEnabled: true,
		LDAPConfig: `url: ldap://ldap.example.com:389
base_dn: dc=example,dc=com
user_filter: (uid={username})
username_attribute: uid
display_name_attribute: cn
email_attribute: mail
`,
	}, models.AuthSetting{})
	if err != nil {
		t.Fatalf("expected legacy flat LDAP YAML to remain valid, got %v", err)
	}
	if item.LDAPURL != "ldap://ldap.example.com:389" {
		t.Fatalf("legacy LDAP YAML was not applied: %#v", item)
	}
}

func TestBuildAuthSettingRejectsUnknownLDAPYAMLField(t *testing.T) {
	_, err := buildAuthSettingForTest(saveAuthSettingReq{
		LDAPConfig: "url: ldap://ldap.example.com:389\nunknown: value\n",
	}, models.AuthSetting{})
	if err == nil || !strings.Contains(err.Error(), "YAML") {
		t.Fatalf("expected LDAP YAML validation error, got %v", err)
	}
}

func TestBuildAuthSettingParsesOIDCYAML(t *testing.T) {
	item, err := buildAuthSettingForTest(saveAuthSettingReq{
		OIDCEnabled: true,
		OIDCConfig: `connection:
  issuer_url: https://accounts.example.com
  client_id: aiops
  client_secret: ""
  redirect_url: https://aiops.example.com/api/auth/oidc/callback
authorization:
  scopes:
    - openid
    - profile
    - email
claims:
  username: preferred_username
`,
	}, models.AuthSetting{OIDCClientSecret: "configured"})
	if err != nil {
		t.Fatalf("expected valid OIDC YAML, got %v", err)
	}
	if item.OIDCIssuerURL != "https://accounts.example.com" || item.OIDCScopes != "openid profile email" || item.OIDCUsernameClaim != "preferred_username" {
		t.Fatalf("OIDC YAML was not applied: %#v", item)
	}
}

func TestBuildAuthSettingRejectsUnknownOIDCYAMLField(t *testing.T) {
	_, err := buildAuthSettingForTest(saveAuthSettingReq{
		OIDCConfig: "connection:\n  issuer: https://accounts.example.com\n",
	}, models.AuthSetting{})
	if err == nil || !strings.Contains(err.Error(), "OIDC YAML") {
		t.Fatalf("expected OIDC YAML validation error, got %v", err)
	}
}
