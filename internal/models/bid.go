package models

import (
	"github.com/google/uuid"
	"time"
)

type Bid struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name            string    `gorm:"type:varchar(255);not null"`
	Description     string    `gorm:"type:text"`
	Status          string    `gorm:"type:varchar(50);not null"`
	TenderID        uuid.UUID `gorm:"type:uuid;not null"`
	Version         int       `json:"version"`
	OrganizationID  uuid.UUID `gorm:"type:uuid;not null"`
	CreatorUsername string    `gorm:"type:varchar(50);not null"`
	CreatedAt       time.Time `gorm:"default:current_timestamp"`
	UpdatedAt       time.Time `gorm:"default:current_timestamp"`
}
