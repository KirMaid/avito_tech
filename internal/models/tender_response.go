package models

import (
	"github.com/google/uuid"
	"time"
)

type TenderResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ServiceType    string    `json:"serviceType"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	OrganizationID uuid.UUID `json:"organizationId"`
}
