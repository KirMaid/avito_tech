package models

import (
	"github.com/google/uuid"
	"time"
)

type TenderResponse struct {
	ID             uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id,omitempty" :"id"`
	Name           string           `gorm:"type:varchar(100);not null" json:"name,omitempty"`
	Description    string           `gorm:"type:text" json:"description,omitempty"`
	ServiceType    string           `gorm:"type:varchar(100)" json:"serviceType,omitempty" :"service_type"`
	Status         TenderStatusType `gorm:"type:varchar(20);not null;default:'CREATED'" json:"status,omitempty" :"status"`
	OrganizationID uuid.UUID        `gorm:"type:uuid;not null" json:"organizationId,omitempty" :"organization_id"`
	CreatedAt      time.Time        `gorm:"autoCreateTime" json:"createdAt" :"created_at"`
	Version        int              `gorm:"type:int;not null" json:"version"`
}
