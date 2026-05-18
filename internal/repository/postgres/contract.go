package postgres

import (
	"asset-leasing-system/internal/domain"
	"time"
)

func (r *ContractRepo) Create(c *domain.Contract) error {
	return r.db.Create(c).Error
}

func (r *ContractRepo) GetByID(id uint) (*domain.Contract, error) {
	var c domain.Contract
	err := r.db.Preload("Asset").Preload("Tenant").First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContractRepo) List(search string, status string, offset, limit int) ([]domain.Contract, int64, error) {
	var contracts []domain.Contract
	var total int64
	q := r.db.Model(&domain.Contract{}).Preload("Asset").Preload("Tenant")
	if search != "" {
		q = q.Joins("JOIN tenants ON tenants.id = contracts.tenant_id").
			Joins("JOIN assets ON assets.id = contracts.asset_id").
			Where("tenants.name LIKE ? OR assets.name LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		q = q.Where("contracts.status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Order("contracts.created_at desc").Find(&contracts).Error; err != nil {
		return nil, 0, err
	}
	return contracts, total, nil
}

func (r *ContractRepo) Update(c *domain.Contract) error {
	return r.db.Save(c).Error
}

func (r *ContractRepo) ListByAssetID(assetID uint) ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.Where("asset_id = ?", assetID).Order("created_at desc").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListByTenantID(tenantID uint) ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListActive() ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.Where("status IN ?", []string{"active", "arrears"}).Find(&contracts).Error
	return contracts, err
}

func (r *PaymentRepo) Create(p *domain.Payment) error {
	return r.db.Create(p).Error
}

func (r *PaymentRepo) GetByID(id uint) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) ListByContractID(contractID uint) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.db.Where("contract_id = ?", contractID).Order("paid_at desc").Find(&payments).Error
	return payments, err
}

func (r *ReceiptRepo) Create(rc *domain.Receipt) error {
	return r.db.Create(rc).Error
}

func (r *ReceiptRepo) GetByID(id uint) (*domain.Receipt, error) {
	var rc domain.Receipt
	err := r.db.First(&rc, id).Error
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReceiptRepo) GetByPaymentID(paymentID uint) (*domain.Receipt, error) {
	var rc domain.Receipt
	err := r.db.Where("payment_id = ?", paymentID).First(&rc).Error
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReceiptBookRepo) Create(rb *domain.ReceiptBook) error {
	return r.db.Create(rb).Error
}

func (r *ReceiptBookRepo) GetByID(id uint) (*domain.ReceiptBook, error) {
	var rb domain.ReceiptBook
	err := r.db.First(&rb, id).Error
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

func (r *ReceiptBookRepo) List() ([]domain.ReceiptBook, error) {
	var books []domain.ReceiptBook
	err := r.db.Order("created_at desc").Find(&books).Error
	return books, err
}

func (r *ReceiptBookRepo) Update(rb *domain.ReceiptBook) error {
	return r.db.Save(rb).Error
}

func (r *ReceiptBookRepo) GetActive() (*domain.ReceiptBook, error) {
	var rb domain.ReceiptBook
	err := r.db.Where("status = ?", "active").First(&rb).Error
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

func (r *TemplateRepo) Create(t *domain.Template) error {
	return r.db.Create(t).Error
}

func (r *TemplateRepo) GetByID(id uint) (*domain.Template, error) {
	var t domain.Template
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) List() ([]domain.Template, error) {
	var templates []domain.Template
	err := r.db.Order("created_at desc").Find(&templates).Error
	return templates, err
}

func (r *TemplateRepo) Update(t *domain.Template) error {
	return r.db.Save(t).Error
}

func (r *UserRepo) Create(u *domain.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepo) GetByUsername(username string) (*domain.User, error) {
	var u domain.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(id uint) (*domain.User, error) {
	var u domain.User
	err := r.db.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) List() ([]domain.User, error) {
	var users []domain.User
	err := r.db.Order("created_at desc").Find(&users).Error
	return users, err
}

func (r *UserRepo) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *UserRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Count(&count).Error
	return count, err
}

func (r *ArrearsRecordRepo) Create(ar *domain.ArrearsRecord) error {
	return r.db.Create(ar).Error
}

func (r *ArrearsRecordRepo) ListByContractID(contractID uint) ([]domain.ArrearsRecord, error) {
	var records []domain.ArrearsRecord
	err := r.db.Where("contract_id = ?", contractID).Order("record_date desc").Find(&records).Error
	return records, err
}

func (r *ArrearsRecordRepo) ListByDateAndLevel(date time.Time, level int) ([]domain.ArrearsRecord, error) {
	var records []domain.ArrearsRecord
	dateStr := date.Format("2006-01-02")
	q := r.db.Where("record_date = ?", dateStr)
	if level > 0 {
		q = q.Where("level = ?", level)
	}
	err := q.Find(&records).Error
	return records, err
}
