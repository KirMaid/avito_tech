package models

import (
	"github.com/google/uuid"
	"time"
)

type TenderVersion struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id,omitempty"`
	Name        string           `gorm:"type:varchar(100);not null"`
	Description string           `gorm:"type:text"`
	ServiceType string           `gorm:"type:varchar(100)" json:"serviceType,omitempty"`
	Status      TenderStatusType `gorm:"type:varchar(20);not null;default:'CREATED'" json:"status,omitempty"`
	CreatedAt   time.Time        `gorm:"autoCreateTime" json:"createdAt"`
	TenderID    uuid.UUID        `gorm:"type:uuid;not null" json:"tenderId"`
	Version     int              `gorm:"type:int;not null" json:"version"`
}
