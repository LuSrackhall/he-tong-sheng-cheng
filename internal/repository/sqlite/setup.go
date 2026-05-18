package sqlite

import (
	"asset-leasing-system/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Setup(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&domain.Asset{},
		&domain.Tenant{},
		&domain.Contract{},
		&domain.Payment{},
		&domain.Receipt{},
		&domain.ReceiptBook{},
		&domain.Template{},
		&domain.User{},
		&domain.ArrearsRecord{},
	); err != nil {
		return nil, err
	}

	// Seed default admin user
	var count int64
	db.Model(&domain.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Create(&domain.User{Username: "admin", Password: string(hash), Role: "admin"})
	}

	return db, nil
}

type AssetRepo struct{ db *gorm.DB }
type TenantRepo struct{ db *gorm.DB }
type ContractRepo struct{ db *gorm.DB }
type PaymentRepo struct{ db *gorm.DB }
type ReceiptRepo struct{ db *gorm.DB }
type ReceiptBookRepo struct{ db *gorm.DB }
type TemplateRepo struct{ db *gorm.DB }
type UserRepo struct{ db *gorm.DB }
type ArrearsRecordRepo struct{ db *gorm.DB }

func NewAssetRepo(db *gorm.DB) *AssetRepo          { return &AssetRepo{db} }
func NewTenantRepo(db *gorm.DB) *TenantRepo          { return &TenantRepo{db} }
func NewContractRepo(db *gorm.DB) *ContractRepo      { return &ContractRepo{db} }
func NewPaymentRepo(db *gorm.DB) *PaymentRepo        { return &PaymentRepo{db} }
func NewReceiptRepo(db *gorm.DB) *ReceiptRepo        { return &ReceiptRepo{db} }
func NewReceiptBookRepo(db *gorm.DB) *ReceiptBookRepo { return &ReceiptBookRepo{db} }
func NewTemplateRepo(db *gorm.DB) *TemplateRepo       { return &TemplateRepo{db} }
func NewUserRepo(db *gorm.DB) *UserRepo              { return &UserRepo{db} }
func NewArrearsRecordRepo(db *gorm.DB) *ArrearsRecordRepo { return &ArrearsRecordRepo{db} }

func (r *AssetRepo) Create(a *domain.Asset) error {
	return r.db.Create(a).Error
}

func (r *AssetRepo) GetByID(id uint) (*domain.Asset, error) {
	var a domain.Asset
	err := r.db.First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AssetRepo) List(search string, assetType string, offset, limit int) ([]domain.Asset, int64, error) {
	var assets []domain.Asset
	var total int64
	q := r.db.Model(&domain.Asset{})
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
	return r.db.Save(a).Error
}
