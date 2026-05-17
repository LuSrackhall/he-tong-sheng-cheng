package domain

import "time"

type Receipt struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ReceiptBookID uint      `json:"receiptBookId" gorm:"not null;index"`
	PaymentID     uint      `json:"paymentId" gorm:"not null;index"`
	SequenceNum   int       `json:"sequenceNum" gorm:"not null"`
	PrintedAt     time.Time `json:"printedAt"`
}
