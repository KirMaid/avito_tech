package models

import (
	"github.com/google/uuid"
	"time"
)

type TenderVersion struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TenderID    uuid.UUID        `gorm:"type:uuid;not null" json:"tenderId"`
	Version     int              `gorm:"type:int;not null" json:"version"`
	Name        string           `gorm:"type:varchar(100);not null" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	ServiceType string           `gorm:"type:varchar(100)" json:"serviceType"`
	Status      TenderStatusType `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt   time.Time        `gorm:"autoCreateTime" json:"createdAt"`
}
