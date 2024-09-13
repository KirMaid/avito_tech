package models

import (
	"github.com/google/uuid"
	"time"
)

type BidVersion struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	BidID       uuid.UUID `gorm:"type:uuid"` // Ссылка на основное предложение
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}
