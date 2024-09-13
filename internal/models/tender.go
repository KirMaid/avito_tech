package models

import (
	"github.com/google/uuid"
	"time"
)

type Tender struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id,omitempty" :"id"`
	Name            string    `gorm:"type:varchar(100);not null"`
	Description     string    `gorm:"type:text"`
	OrganizationID  uuid.UUID `gorm:"type:uuid;not null" json:"organization_id,omitempty" :"organization_id"`
	ServiceType     string    `gorm:"type:varchar(100)" json:"service_type,omitempty" :"service_type"`
	Status          string    `gorm:"type:varchar(20);not null;default:'CREATED'" json:"status,omitempty" :"status"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at" :"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at" :"updated_at"`
	CreatorUsername string    `json:"creatorUsername" validate:"required"`
}
