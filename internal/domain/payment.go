package domain

import "time"

type Payment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ContractID uint      `json:"contractId" gorm:"not null;index"`
	Amount     float64   `json:"amount" gorm:"not null"`
	PaidAt     time.Time `json:"paidAt" gorm:"not null"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}
