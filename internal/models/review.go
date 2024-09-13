package models

import (
	"github.com/google/uuid"
	"time"
)

type Review struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	BidID          uuid.UUID `gorm:"type:uuid"`
	AuthorUsername string    `json:"authorUsername"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Rating         int       `json:"rating"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
