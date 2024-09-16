package models

import (
	"github.com/google/uuid"
	"time"
)

type BidStatusType string

const (
	BidStatusCreated   BidStatusType = "Created"
	BidStatusPublished BidStatusType = "Published"
	BidStatusCanceled  BidStatusType = "Canceled"
)

type Bid struct {
	ID              uuid.UUID     `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name            string        `gorm:"type:varchar(255);not null" json:"name"`
	Description     string        `gorm:"type:text" json:"description"`
	Status          BidStatusType `gorm:"type:varchar(20);not null;default:'Created'" json:"status"`
	TenderID        uuid.UUID     `gorm:"type:uuid;not null" json:"tenderId"`
	OrganizationID  uuid.UUID     `gorm:"type:uuid;not null" json:"organizationId"`
	Version         int           `gorm:"default:1" json:"version"`
	CreatorUsername string        `json:"creatorUsername" validate:"required"`
	CreatedAt       time.Time     `gorm:"default:current_timestamp" json:"createdAt"`
	UpdatedAt       time.Time     `gorm:"default:current_timestamp" json:"updatedAt"`
}
