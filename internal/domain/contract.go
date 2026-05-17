package domain

import "time"

type Contract struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	AssetID         uint      `json:"assetId" gorm:"not null;index"`
	Asset           *Asset    `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
	TenantID        uint      `json:"tenantId" gorm:"not null;index"`
	Tenant          *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	StartDate       time.Time `json:"startDate" gorm:"not null"`
	EndDate         time.Time `json:"endDate" gorm:"not null"`
	MonthlyRent     float64   `json:"monthlyRent" gorm:"not null"`
	TotalReceivable float64   `json:"totalReceivable" gorm:"not null"`
	TotalReceived   float64   `json:"totalReceived" gorm:"default:0"`
	Deposit         float64   `json:"deposit" gorm:"default:0"`
	Status          string    `json:"status" gorm:"default:'active'"`
	TemplateID      *uint     `json:"templateId,omitempty"`
	Notes           string    `json:"notes,omitempty" gorm:"type:text"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
