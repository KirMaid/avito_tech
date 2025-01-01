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
	ID              uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id,omitempty" :"id"`
	Name            string           `gorm:"type:varchar(100);not null"`
	Description     string           `gorm:"type:text"`
	ServiceType     string           `gorm:"type:varchar(100)" json:"serviceType,omitempty" :"service_type"`
	Status          TenderStatusType `gorm:"type:varchar(20);not null;default:'CREATED'" json:"status,omitempty" :"status"`
	OrganizationID  uuid.UUID        `gorm:"type:uuid;not null" json:"organizationId,omitempty" :"organization_id"`
	CreatedAt       time.Time        `gorm:"autoCreateTime" json:"createdAt" :"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime" json:"updated_at" :"updated_at"`
	CreatorUsername string           `json:"creatorUsername" validate:"required"`
	Version         int              `json:"version"`
}
