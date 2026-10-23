package models

import "time"

type Certificate struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Domains           string    `gorm:"type:text" json:"domains"` // JSON array
	Issuer            string    `gorm:"size:255" json:"issuer"`
	Subject           string    `gorm:"size:255" json:"subject"`
	Source            string    `gorm:"size:50;not null;default:manual" json:"source"` // manual, acme, cloud-managed
	SerialNumber      string    `gorm:"size:128;index" json:"serial_number"`
	FingerprintSHA256 string    `gorm:"size:95;index" json:"fingerprint_sha256"`
	NotBefore         time.Time `json:"not_before"`
	NotAfter          time.Time `gorm:"index" json:"not_after"`
	Status            string    `gorm:"size:20;not null;default:active;index" json:"status"`
	Owner             string    `gorm:"size:100" json:"owner"`
	Environment       string    `gorm:"size:50;index" json:"environment"`
	CurrentVersionID  uint      `json:"current_version_id"`
	Description       string    `gorm:"size:500" json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CertificateVersion struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	CertificateID     uint       `gorm:"index;not null" json:"certificate_id"`
	Version           string     `gorm:"size:64;not null;index" json:"version"`
	CertPEM           string     `gorm:"type:longtext" json:"-"`
	PrivateKeyCipher  string     `gorm:"type:longtext" json:"-"`
	Domains           string     `gorm:"type:text" json:"domains"` // JSON array
	Issuer            string     `gorm:"size:255" json:"issuer"`
	Subject           string     `gorm:"size:255" json:"subject"`
	SerialNumber      string     `gorm:"size:128;index" json:"serial_number"`
	FingerprintSHA256 string     `gorm:"size:95;index" json:"fingerprint_sha256"`
	NotBefore         time.Time  `json:"not_before"`
	NotAfter          time.Time  `gorm:"index" json:"not_after"`
	Status            string     `gorm:"size:20;not null;default:active;index" json:"status"`
	ActivatedAt       *time.Time `json:"activated_at"`
	PreviousVersionID uint       `json:"previous_version_id"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CertificateDomain struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	Domain                string     `gorm:"size:255;not null;index" json:"domain"`
	Port                  int        `gorm:"not null;default:443" json:"port"`
	CertificateID         uint       `gorm:"index" json:"certificate_id"`
	ExecutorID            uint       `gorm:"index" json:"executor_id"`
	AgentID               uint       `gorm:"index" json:"agent_id"`
	DeployScriptID        uint       `gorm:"index" json:"deploy_script_id"`
	FaultCenterID         uint       `gorm:"index" json:"fault_center_id"`
	NotificationChannelID uint       `gorm:"index" json:"notification_channel_id"`
	DeployParams          string     `gorm:"type:text" json:"deploy_params"`
	DeployCertPath        string     `gorm:"size:255" json:"deploy_cert_path"`
	DeployKeyPath         string     `gorm:"size:255" json:"deploy_key_path"`
	DeployReloadCommand   string     `gorm:"size:500" json:"deploy_reload_command"`
	Environment           string     `gorm:"size:50;index" json:"environment"`
	Owner                 string     `gorm:"size:100" json:"owner"`
	MonitorEnabled        bool       `gorm:"not null;default:true" json:"monitor_enabled"`
	CheckIntervalMinutes  int        `gorm:"not null;default:360" json:"check_interval_minutes"`
	ExpiryAlertDays       int        `gorm:"not null;default:30" json:"expiry_alert_days"`
	Severity              string     `gorm:"size:20;not null;default:P2;index" json:"severity"`
	AlertBeforeDays       string     `gorm:"size:100;default:'30,15,7,3,1'" json:"alert_before_days"`
	LastObservedVersionID uint       `json:"last_observed_version_id"`
	LastSeenFingerprint   string     `gorm:"size:95" json:"last_seen_fingerprint"`
	ObservedIssuer        string     `gorm:"size:255" json:"observed_issuer"`
	ObservedSubject       string     `gorm:"size:255" json:"observed_subject"`
	ObservedNotBefore     *time.Time `json:"observed_not_before"`
	ObservedExpiresAt     *time.Time `json:"observed_expires_at"`
	DaysRemaining         int        `json:"days_remaining"`
	LastCheckStatus       string     `gorm:"size:20;index" json:"last_check_status"` // success, failed
	LastCheckError        string     `gorm:"type:text" json:"last_check_error"`
	LastCheckedAt         *time.Time `json:"last_checked_at"`
	Status                string     `gorm:"size:30;index" json:"status"` // unknown, normal, expiring, expired, mismatch, failed
	Description           string     `gorm:"size:500" json:"description"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type CertificateDeployScript struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Type        string    `gorm:"size:30;not null;default:shell" json:"type"`
	TargetType  string    `gorm:"size:50;not null;default:custom" json:"target_type"`
	Content     string    `gorm:"type:longtext;not null" json:"content"`
	Description string    `gorm:"size:500" json:"description"`
	Status      int       `gorm:"not null;default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CertificateExecutor struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128;not null;index" json:"name"`
	Type         string    `gorm:"size:30;not null;default:ssh" json:"type"` // ssh
	Host         string    `gorm:"size:255;not null" json:"host"`
	Port         int       `gorm:"not null;default:22" json:"port"`
	Username     string    `gorm:"size:100;not null" json:"username"`
	AuthType     string    `gorm:"size:30;not null;default:password" json:"auth_type"` // password, private_key
	SecretCipher string    `gorm:"type:longtext" json:"-"`
	Status       int       `gorm:"not null;default:1" json:"status"`
	Description  string    `gorm:"size:500" json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CertificateDeploymentRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	DomainID       uint      `gorm:"index;not null" json:"domain_id"`
	CertificateID  uint      `gorm:"index;not null" json:"certificate_id"`
	ExecutorID     uint      `gorm:"index;not null" json:"executor_id"`
	AgentID        uint      `gorm:"index" json:"agent_id"`
	DeployScriptID uint      `gorm:"index" json:"deploy_script_id"`
	TargetType     string    `gorm:"size:50;not null;default:linux-nginx" json:"target_type"`
	Status         string    `gorm:"size:20;not null;index" json:"status"`
	Output         string    `gorm:"type:text" json:"output"`
	DurationMs     int64     `json:"duration_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

type CertificateRotationRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Provider  string    `gorm:"size:50;not null;index" json:"provider"`
	Product   string    `gorm:"size:50;not null;index" json:"product"`
	Domain    string    `gorm:"size:255;not null;index" json:"domain"`
	CertName  string    `gorm:"size:128" json:"cert_name"`
	Status    string    `gorm:"size:20;not null;index" json:"status"`
	RequestID string    `gorm:"size:100" json:"request_id"`
	Error     string    `gorm:"type:text" json:"error"`
	CreatedAt time.Time `json:"created_at"`
}
