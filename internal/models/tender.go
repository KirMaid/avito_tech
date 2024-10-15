package models

import (
	"github.com/google/uuid"
	"time"
)

type TenderStatusType string

const (
	TenderStatusCreated   TenderStatusType = "CREATED"
	TenderStatusPublished TenderStatusType = "PUBLISHED"
	TenderStatusClosed    TenderStatusType = "CLOSED"
)

type Tender struct {
	ID              uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id,omitempty"`
	Name            string           `gorm:"type:varchar(100);not null"`
	Description     string           `gorm:"type:text"`
	ServiceType     string           `gorm:"type:varchar(100)" json:"serviceType,omitempty"`
	Status          TenderStatusType `gorm:"type:varchar(20);not null;default:'CREATED'" json:"status,omitempty" `
	OrganizationID  uuid.UUID        `gorm:"type:uuid;not null" json:"organizationId,omitempty"`
	CreatedAt       time.Time        `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	CreatorUsername string           `json:"creatorUsername" validate:"required"`
	Version         int              `json:"version"`
}
