package models

import (
	"github.com/google/uuid"
	"time"
)

type OrganizationType string

const (
	OrganizationTypeIE  OrganizationType = "IE"
	OrganizationTypeLLC OrganizationType = "LLC"
	OrganizationTypeJSC OrganizationType = "JSC"
)

type Organization struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string           `gorm:"type:varchar(100);not null"`
	Description string           `gorm:"type:text"`
	Type        OrganizationType `gorm:"type:organization_type"`
	CreatedAt   time.Time        `gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `gorm:"autoUpdateTime"`
}

func (Organization) TableName() string {
	return "organization"
}
