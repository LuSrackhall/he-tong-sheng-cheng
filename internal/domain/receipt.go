package domain

import "time"

type Receipt struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ReceiptBookID uint      `json:"receiptBookId" gorm:"not null;index"`
	PaymentID     uint      `json:"paymentId" gorm:"not null;uniqueIndex"`
	SequenceNum   int       `json:"sequenceNum" gorm:"not null"`
	Amount        float64   `json:"amount" gorm:"not null"`
	PrintedAt     time.Time `json:"printedAt"`
}
