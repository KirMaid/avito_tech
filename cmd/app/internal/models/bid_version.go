package models

import (
	"github.com/google/uuid"
	"time"
)

type BidVersion struct {
	ID          uuid.UUID     `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	BidID       uuid.UUID     `gorm:"type:uuid;not null" json:"bidId"`
	Version     int           `gorm:"not null" json:"version"`
	Name        string        `gorm:"type:varchar(255);not null" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	Status      BidStatusType `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt   time.Time     `gorm:"default:current_timestamp" json:"createdAt"`
}
