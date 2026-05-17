package domain

import "time"

type Template struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	FilePath    string    `json:"filePath" gorm:"not null"`
	FieldMap    string    `json:"fieldMap,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
