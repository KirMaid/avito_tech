package models

import (
	"github.com/google/uuid"
	"time"
)

type TenderVersion struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenderID    uuid.UUID `gorm:"type:uuid;not null"`
	Version     int       `gorm:"not null"`
	Name        string    `gorm:"type:varchar(255)"`
	Description string    `gorm:"type:text"`
	ServiceType string    `gorm:"type:varchar(255)"`
	Status      string    `gorm:"type:varchar(50)"`
	CreatedAt   time.Time
}
