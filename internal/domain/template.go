package domain

import "time"

type Template struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	FilePath     string    `json:"filePath" gorm:"default:''"`
	FieldMap     string    `json:"fieldMap,omitempty" gorm:"type:text"`
	ActiveFields string    `json:"activeFields,omitempty" gorm:"type:text"`
	Validated    bool      `json:"validated" gorm:"default:false"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
