package postgres

import (
	"asset-leasing-system/internal/domain"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Setup(host, port, user, pass, dbname string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
		seedPass := os.Getenv("ADMIN_PASSWORD")
		if seedPass == "" {
			seedPass = "admin123"
			log.Println("WARNING: Using default admin password. Set ADMIN_PASSWORD env var for production.")
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(seedPass), bcrypt.DefaultCost)
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
