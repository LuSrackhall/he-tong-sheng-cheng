package common

import (
	"asset-leasing-system/internal/domain"
	"time"

	"gorm.io/gorm"
)

// 所有 repo struct 定义

type AssetRepo struct{ DB *gorm.DB }
type TenantRepo struct{ DB *gorm.DB }
type ContractRepo struct{ DB *gorm.DB }
type PaymentRepo struct{ DB *gorm.DB }
type ReceiptRepo struct{ DB *gorm.DB }
type ReceiptBookRepo struct{ DB *gorm.DB }
type TemplateRepo struct{ DB *gorm.DB }
type UserRepo struct{ DB *gorm.DB }
type ArrearsRecordRepo struct{ DB *gorm.DB }

// 所有 repo 构造函数

func NewAssetRepo(db *gorm.DB) *AssetRepo            { return &AssetRepo{DB: db} }
func NewTenantRepo(db *gorm.DB) *TenantRepo           { return &TenantRepo{DB: db} }
func NewContractRepo(db *gorm.DB) *ContractRepo       { return &ContractRepo{DB: db} }
func NewPaymentRepo(db *gorm.DB) *PaymentRepo         { return &PaymentRepo{DB: db} }
func NewReceiptRepo(db *gorm.DB) *ReceiptRepo         { return &ReceiptRepo{DB: db} }
func NewReceiptBookRepo(db *gorm.DB) *ReceiptBookRepo { return &ReceiptBookRepo{DB: db} }
func NewTemplateRepo(db *gorm.DB) *TemplateRepo       { return &TemplateRepo{DB: db} }
func NewUserRepo(db *gorm.DB) *UserRepo               { return &UserRepo{DB: db} }
func NewArrearsRecordRepo(db *gorm.DB) *ArrearsRecordRepo {
	return &ArrearsRecordRepo{DB: db}
}

// 接口断言
var _ domain.AssetRepo = (*AssetRepo)(nil)
var _ domain.TenantRepo = (*TenantRepo)(nil)
var _ domain.ContractRepo = (*ContractRepo)(nil)
var _ domain.PaymentRepo = (*PaymentRepo)(nil)
var _ domain.ReceiptRepo = (*ReceiptRepo)(nil)
var _ domain.ReceiptBookRepo = (*ReceiptBookRepo)(nil)
var _ domain.TemplateRepo = (*TemplateRepo)(nil)
var _ domain.UserRepo = (*UserRepo)(nil)
var _ domain.ArrearsRecordRepo = (*ArrearsRecordRepo)(nil)

// ── AssetRepo ──

func (r *AssetRepo) Create(a *domain.Asset) error {
	return r.DB.Create(a).Error
}

func (r *AssetRepo) GetByID(id uint) (*domain.Asset, error) {
	var a domain.Asset
	err := r.DB.First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AssetRepo) List(search string, assetType string, offset, limit int) ([]domain.Asset, int64, error) {
	var assets []domain.Asset
	var total int64
	q := r.DB.Model(&domain.Asset{})
	if search != "" {
		q = q.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if assetType != "" {
		q = q.Where("asset_type = ?", assetType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Order("created_at desc").Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

func (r *AssetRepo) Update(a *domain.Asset) error {
	return r.DB.Save(a).Error
}

// ── TenantRepo ──

func (r *TenantRepo) Create(t *domain.Tenant) error {
	return r.DB.Create(t).Error
}

func (r *TenantRepo) GetByID(id uint) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.DB.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) List(search string, offset, limit int) ([]domain.Tenant, int64, error) {
	var tenants []domain.Tenant
	var total int64
	q := r.DB.Model(&domain.Tenant{})
	if search != "" {
		q = q.Where("name LIKE ? OR phone LIKE ? OR id_card LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Order("created_at desc").Find(&tenants).Error; err != nil {
		return nil, 0, err
	}
	return tenants, total, nil
}

func (r *TenantRepo) Update(t *domain.Tenant) error {
	return r.DB.Save(t).Error
}

// ── ContractRepo ──

func (r *ContractRepo) Create(c *domain.Contract) error {
	return r.DB.Create(c).Error
}

func (r *ContractRepo) GetByID(id uint) (*domain.Contract, error) {
	var c domain.Contract
	err := r.DB.Preload("Asset").Preload("Tenant").First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContractRepo) List(search string, status string, offset, limit int) ([]domain.Contract, int64, error) {
	var contracts []domain.Contract
	var total int64
	q := r.DB.Model(&domain.Contract{}).Preload("Asset").Preload("Tenant")
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
	return r.DB.Save(c).Error
}

func (r *ContractRepo) ListByAssetID(assetID uint) ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.DB.Where("asset_id = ?", assetID).Order("created_at desc").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListByTenantID(tenantID uint) ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.DB.Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListActive() ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.DB.Where("status IN ?", []string{"active", "arrears"}).Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListUnpaid() ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.DB.Where("status != ?", "paidup").Preload("Asset").Preload("Tenant").Find(&contracts).Error
	return contracts, err
}

// ── PaymentRepo ──

func (r *PaymentRepo) Create(p *domain.Payment) error {
	return r.DB.Create(p).Error
}

func (r *PaymentRepo) GetByID(id uint) (*domain.Payment, error) {
	var p domain.Payment
	err := r.DB.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) ListByContractID(contractID uint) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.DB.Where("contract_id = ?", contractID).Order("paid_at desc").Find(&payments).Error
	return payments, err
}

// ── ReceiptRepo ──

func (r *ReceiptRepo) Create(rc *domain.Receipt) error {
	return r.DB.Create(rc).Error
}

func (r *ReceiptRepo) GetByID(id uint) (*domain.Receipt, error) {
	var rc domain.Receipt
	err := r.DB.First(&rc, id).Error
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReceiptRepo) GetByPaymentID(paymentID uint) (*domain.Receipt, error) {
	var rc domain.Receipt
	err := r.DB.Where("payment_id = ?", paymentID).First(&rc).Error
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReceiptRepo) ListByReceiptBookID(bookID uint) ([]domain.Receipt, error) {
	var receipts []domain.Receipt
	err := r.DB.Where("receipt_book_id = ?", bookID).Order("sequence_num desc").Find(&receipts).Error
	return receipts, err
}

func (r *ReceiptRepo) List(offset, limit int) ([]domain.Receipt, int64, error) {
	var receipts []domain.Receipt
	var total int64
	if err := r.DB.Model(&domain.Receipt{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.DB.Order("id desc").Offset(offset).Limit(limit).Find(&receipts).Error
	return receipts, total, err
}

// ── ReceiptBookRepo ──

func (r *ReceiptBookRepo) Create(rb *domain.ReceiptBook) error {
	return r.DB.Create(rb).Error
}

func (r *ReceiptBookRepo) GetByID(id uint) (*domain.ReceiptBook, error) {
	var rb domain.ReceiptBook
	err := r.DB.First(&rb, id).Error
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

func (r *ReceiptBookRepo) List() ([]domain.ReceiptBook, error) {
	var books []domain.ReceiptBook
	err := r.DB.Order("created_at desc").Find(&books).Error
	return books, err
}

func (r *ReceiptBookRepo) Update(rb *domain.ReceiptBook) error {
	return r.DB.Save(rb).Error
}

func (r *ReceiptBookRepo) GetActive() (*domain.ReceiptBook, error) {
	var rb domain.ReceiptBook
	err := r.DB.Where("status = ?", "active").First(&rb).Error
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

func (r *ReceiptBookRepo) AllocateSequence(bookID uint) (int, error) {
	// 原子递增 CurrentNum，避免并发冲突
	result := r.DB.Model(&domain.ReceiptBook{}).
		Where("id = ? AND current_num < start_num + total_pages AND status = ?", bookID, "active").
		Update("current_num", gorm.Expr("current_num + 1"))
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, nil // 收据本已用完
	}
	var rb domain.ReceiptBook
	if err := r.DB.First(&rb, bookID).Error; err != nil {
		return 0, err
	}
	return rb.CurrentNum, nil
}

// ── TemplateRepo ──

func (r *TemplateRepo) Create(t *domain.Template) error {
	return r.DB.Create(t).Error
}

func (r *TemplateRepo) GetByID(id uint) (*domain.Template, error) {
	var t domain.Template
	err := r.DB.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) List() ([]domain.Template, error) {
	var templates []domain.Template
	err := r.DB.Order("created_at desc").Find(&templates).Error
	return templates, err
}

func (r *TemplateRepo) Update(t *domain.Template) error {
	return r.DB.Save(t).Error
}

func (r *TemplateRepo) Delete(id uint) error {
	return r.DB.Delete(&domain.Template{}, id).Error
}

func (r *TemplateRepo) IsUsedByContract(id uint) (bool, error) {
	var count int64
	err := r.DB.Model(&domain.Contract{}).Where("template_id = ?", id).Count(&count).Error
	return count > 0, err
}

// ── UserRepo ──

func (r *UserRepo) Create(u *domain.User) error {
	return r.DB.Create(u).Error
}

func (r *UserRepo) GetByUsername(username string) (*domain.User, error) {
	var u domain.User
	err := r.DB.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(id uint) (*domain.User, error) {
	var u domain.User
	err := r.DB.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) List() ([]domain.User, error) {
	var users []domain.User
	err := r.DB.Order("created_at desc").Find(&users).Error
	return users, err
}

func (r *UserRepo) Delete(id uint) error {
	return r.DB.Delete(&domain.User{}, id).Error
}

func (r *UserRepo) Count() (int64, error) {
	var count int64
	err := r.DB.Model(&domain.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepo) Update(u *domain.User) error {
	return r.DB.Save(u).Error
}

// ── ArrearsRecordRepo ──

func (r *ArrearsRecordRepo) Create(ar *domain.ArrearsRecord) error {
	return r.DB.Create(ar).Error
}

func (r *ArrearsRecordRepo) ListByContractID(contractID uint) ([]domain.ArrearsRecord, error) {
	var records []domain.ArrearsRecord
	err := r.DB.Where("contract_id = ?", contractID).Order("record_date desc").Find(&records).Error
	return records, err
}

func (r *ArrearsRecordRepo) ListByDateAndLevel(date time.Time, level int) ([]domain.ArrearsRecord, error) {
	var records []domain.ArrearsRecord
	dateStr := date.Format("2006-01-02")
	q := r.DB.Where("record_date = ?", dateStr)
	if level > 0 {
		q = q.Where("level = ?", level)
	}
	err := q.Find(&records).Error
	return records, err
}
