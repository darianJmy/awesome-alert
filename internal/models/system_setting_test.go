package models

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestAuthSettingUsesCanonicalOIDCColumnNames(t *testing.T) {
	parsed, err := schema.Parse(&AuthSetting{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse AuthSetting schema: %v", err)
	}
	want := map[string]string{
		"OIDCEnabled":       "oidc_enabled",
		"OIDCIssuerURL":     "oidc_issuer_url",
		"OIDCClientID":      "oidc_client_id",
		"OIDCClientSecret":  "oidc_client_secret",
		"OIDCRedirectURL":   "oidc_redirect_url",
		"OIDCScopes":        "oidc_scopes",
		"OIDCUsernameClaim": "oidc_username_claim",
	}
	for fieldName, columnName := range want {
		field := parsed.LookUpField(fieldName)
		if field == nil {
			t.Fatalf("field %s not found", fieldName)
		}
		if field.DBName != columnName {
			t.Fatalf("%s column = %q, want %q", fieldName, field.DBName, columnName)
		}
	}
}
