package models

import "time"

type SystemSetting struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SystemName     string    `gorm:"size:100;not null;default:运维中枢" json:"system_name"`
	SystemSubtitle string    `gorm:"size:200;not null;default:智能运维控制台" json:"system_subtitle"`
	LogoData       string    `gorm:"type:longtext" json:"logo_data"`
	DefaultTheme   string    `gorm:"size:20;not null;default:system" json:"default_theme"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AuthSetting struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	LocalEnabled        bool      `gorm:"not null;default:true" json:"local_enabled"`
	LDAPEnabled         bool      `gorm:"not null;default:false" json:"ldap_enabled"`
	LDAPURL             string    `gorm:"size:500" json:"ldap_url"`
	LDAPBaseDN          string    `gorm:"size:500" json:"ldap_base_dn"`
	LDAPBindDN          string    `gorm:"size:500" json:"ldap_bind_dn"`
	LDAPBindPassword    string    `gorm:"type:longtext" json:"-"`
	LDAPUserFilter      string    `gorm:"size:500" json:"ldap_user_filter"`
	LDAPUsernameAttr    string    `gorm:"size:100" json:"ldap_username_attr"`
	LDAPDisplayNameAttr string    `gorm:"size:100" json:"ldap_display_name_attr"`
	LDAPEmailAttr       string    `gorm:"size:100" json:"ldap_email_attr"`
	OIDCEnabled         bool      `gorm:"column:oidc_enabled;not null;default:false" json:"oidc_enabled"`
	OIDCIssuerURL       string    `gorm:"column:oidc_issuer_url;size:500" json:"oidc_issuer_url"`
	OIDCClientID        string    `gorm:"column:oidc_client_id;size:255" json:"oidc_client_id"`
	OIDCClientSecret    string    `gorm:"column:oidc_client_secret;type:longtext" json:"-"`
	OIDCRedirectURL     string    `gorm:"column:oidc_redirect_url;size:500" json:"oidc_redirect_url"`
	OIDCScopes          string    `gorm:"column:oidc_scopes;size:500" json:"oidc_scopes"`
	OIDCUsernameClaim   string    `gorm:"column:oidc_username_claim;size:100" json:"oidc_username_claim"`
	DefaultRoleID       uint      `gorm:"index" json:"default_role_id"`
	UpdatedAt           time.Time `json:"updated_at"`
}
