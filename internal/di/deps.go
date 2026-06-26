package di

import (
	"asset-leasing-system/internal/config"
	"asset-leasing-system/internal/domain"
	"asset-leasing-system/internal/repository/postgres"
	"asset-leasing-system/internal/repository/sqlite"
	"log"

	"gorm.io/gorm"
)

type Dependencies struct {
	DB              *gorm.DB
	AssetRepo       domain.AssetRepo
	TenantRepo      domain.TenantRepo
	ContractRepo    domain.ContractRepo
	PaymentRepo     domain.PaymentRepo
	ReceiptRepo     domain.ReceiptRepo
	ReceiptBookRepo domain.ReceiptBookRepo
	TemplateRepo    domain.TemplateRepo
	UserRepo        domain.UserRepo
	ArrearsRepo     domain.ArrearsRecordRepo
}

func Initialize(cfg *config.Config) *Dependencies {
	var db *gorm.DB
	var err error

	switch cfg.Mode {
	case "postgres":
		db, err = postgres.Setup(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode, cfg.AdminPassword)
	default:
		db, err = sqlite.Setup(cfg.DBName+".db", cfg.AdminPassword)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return wire(cfg, db)
}

func wire(cfg *config.Config, db *gorm.DB) *Dependencies {
	switch cfg.Mode {
	case "postgres":
		return &Dependencies{
			DB:              db,
			AssetRepo:       postgres.NewAssetRepo(db),
			TenantRepo:      postgres.NewTenantRepo(db),
			ContractRepo:    postgres.NewContractRepo(db),
			PaymentRepo:     postgres.NewPaymentRepo(db),
			ReceiptRepo:     postgres.NewReceiptRepo(db),
			ReceiptBookRepo: postgres.NewReceiptBookRepo(db),
			TemplateRepo:    postgres.NewTemplateRepo(db),
			UserRepo:        postgres.NewUserRepo(db),
			ArrearsRepo:     postgres.NewArrearsRecordRepo(db),
		}
	default:
		return &Dependencies{
			DB:              db,
			AssetRepo:       sqlite.NewAssetRepo(db),
			TenantRepo:      sqlite.NewTenantRepo(db),
			ContractRepo:    sqlite.NewContractRepo(db),
			PaymentRepo:     sqlite.NewPaymentRepo(db),
			ReceiptRepo:     sqlite.NewReceiptRepo(db),
			ReceiptBookRepo: sqlite.NewReceiptBookRepo(db),
			TemplateRepo:    sqlite.NewTemplateRepo(db),
			UserRepo:        sqlite.NewUserRepo(db),
			ArrearsRepo:     sqlite.NewArrearsRecordRepo(db),
		}
	}
}
