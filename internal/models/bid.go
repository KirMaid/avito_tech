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

type AuthorType string

const (
	AuthorTypeUser         AuthorType = "User"
	AuthorTypeOrganization AuthorType = "Organization"
)

type Bid struct {
	ID          uuid.UUID     `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string        `gorm:"type:varchar(255);not null" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	Status      BidStatusType `gorm:"type:varchar(20);not null;default:'Created'" json:"status"`
	TenderID    uuid.UUID     `gorm:"type:uuid;not null" json:"tenderId"`
	Version     int           `gorm:"default:1" json:"version"`
	AuthorType  AuthorType    `gorm:"type:varchar(20);not null" json:"authorType"`
	AuthorID    uuid.UUID     `gorm:"type:uuid;not null" json:"authorId"`
	CreatedAt   time.Time     `gorm:"default:current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time     `gorm:"default:current_timestamp" json:"updatedAt"`
}
