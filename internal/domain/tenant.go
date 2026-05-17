package domain

import "time"

type Tenant struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	IDCard      string    `json:"idCard,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	IDCardImage string    `json:"idCardImage,omitempty"`
	ExtraFields string    `json:"extraFields,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
