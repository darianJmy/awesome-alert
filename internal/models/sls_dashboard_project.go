package models

import "time"

type SLSDashboardProject struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	Provider              string    `gorm:"size:20;not null;default:aliyun_sls;index" json:"provider"`
	Name                  string    `gorm:"size:100;not null" json:"name"`
	Project               string    `gorm:"size:100;not null;index" json:"project"`
	Region                string    `gorm:"size:50;not null;index" json:"region"`
	RoleArn               string    `gorm:"size:255" json:"role_arn"`
	AccessKeyIDCipher     string    `gorm:"type:text;not null" json:"-"`
	AccessKeySecretCipher string    `gorm:"type:text;not null" json:"-"`
	Description           string    `gorm:"size:500" json:"description"`
	Sort                  int       `gorm:"default:0;index" json:"sort"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
