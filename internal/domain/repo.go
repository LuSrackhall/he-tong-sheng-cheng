package domain

import "time"

type AssetRepo interface {
	Create(a *Asset) error
	GetByID(id uint) (*Asset, error)
	List(search string, assetType string, offset, limit int) ([]Asset, int64, error)
	Update(a *Asset) error
}

type TenantRepo interface {
	Create(t *Tenant) error
	GetByID(id uint) (*Tenant, error)
	List(search string, offset, limit int) ([]Tenant, int64, error)
	Update(t *Tenant) error
}

type ContractRepo interface {
	Create(c *Contract) error
	GetByID(id uint) (*Contract, error)
	List(search string, status string, offset, limit int) ([]Contract, int64, error)
	Update(c *Contract) error
	ListByAssetID(assetID uint) ([]Contract, error)
	ListByTenantID(tenantID uint) ([]Contract, error)
	ListActive() ([]Contract, error)
	ListUnpaid() ([]Contract, error)
	ListUnpaidPaging(offset, limit int) ([]Contract, int64, error)
	CheckOverlap(assetID, tenantID uint, start, end time.Time) (bool, error)
}

type PaymentRepo interface {
	Create(p *Payment) error
	GetByID(id uint) (*Payment, error)
	ListByContractID(contractID uint) ([]Payment, error)
}

type ReceiptRepo interface {
	Create(r *Receipt) error
	GetByID(id uint) (*Receipt, error)
	GetByPaymentID(paymentID uint) (*Receipt, error)
	ListByReceiptBookID(bookID uint) ([]Receipt, error)
	List(offset, limit int) ([]Receipt, int64, error)
}

type ReceiptBookRepo interface {
	Create(rb *ReceiptBook) error
	GetByID(id uint) (*ReceiptBook, error)
	List() ([]ReceiptBook, error)
	Update(rb *ReceiptBook) error
	GetActive() (*ReceiptBook, error)
	AllocateSequence(bookID uint) (int, error)
}

type TemplateRepo interface {
	Create(t *Template) error
	GetByID(id uint) (*Template, error)
	List() ([]Template, error)
	Update(t *Template) error
	Delete(id uint) error
	IsUsedByContract(id uint) (bool, error)
}

type UserRepo interface {
	Create(u *User) error
	GetByUsername(username string) (*User, error)
	GetByID(id uint) (*User, error)
	List() ([]User, error)
	Update(u *User) error
	Delete(id uint) error
	Count() (int64, error)
}

type ArrearsRecordRepo interface {
	Create(r *ArrearsRecord) error
	ListByContractID(contractID uint) ([]ArrearsRecord, error)
	ListByDateAndLevel(date time.Time, level int) ([]ArrearsRecord, error)
}

type ArrearsRecord struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ContractID uint      `json:"contractId" gorm:"not null;index"`
	Level      int       `json:"level" gorm:"not null"`
	RecordDate time.Time `json:"recordDate" gorm:"not null;index"`
	CreatedAt  time.Time `json:"createdAt"`
}
