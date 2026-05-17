package domain

import "time"

type ReceiptBook struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Prefix     string    `json:"prefix" gorm:"not null"`
	StartNum   int       `json:"startNum" gorm:"not null"`
	CurrentNum int       `json:"currentNum" gorm:"not null"`
	TotalPages int       `json:"totalPages" gorm:"not null"`
	Status     string    `json:"status" gorm:"default:'active'"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
