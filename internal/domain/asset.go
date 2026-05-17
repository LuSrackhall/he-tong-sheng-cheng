package domain

import "time"

// Asset represents a leased property (shop, parking space, booth, equipment, etc.)
type Asset struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	AssetType   string    `json:"assetType" gorm:"not null"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status" gorm:"default:'idle'"`
	ExtraFields string    `json:"extraFields,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
