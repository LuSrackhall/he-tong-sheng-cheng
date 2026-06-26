package di

import (
	"asset-leasing-system/internal/config"
	"asset-leasing-system/internal/repository/common"
	"asset-leasing-system/internal/repository/postgres"
	"asset-leasing-system/internal/repository/sqlite"
	"log"

	"gorm.io/gorm"
)

type Dependencies struct {
	DB              *gorm.DB
	AssetRepo       *common.AssetRepo
	TenantRepo      *common.TenantRepo
	ContractRepo    *common.ContractRepo
	PaymentRepo     *common.PaymentRepo
	ReceiptRepo     *common.ReceiptRepo
	ReceiptBookRepo *common.ReceiptBookRepo
	TemplateRepo    *common.TemplateRepo
	UserRepo        *common.UserRepo
	ArrearsRepo     *common.ArrearsRecordRepo
}

func Initialize(cfg *config.Config) *Dependencies {
	var db *gorm.DB
	var err error

	switch cfg.Mode {
	case "postgres":
		db, err = postgres.Setup(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	default:
		db, err = sqlite.Setup(cfg.DBName + ".db")
	}
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	return &Dependencies{
		DB:              db,
		AssetRepo:       common.NewAssetRepo(db),
		TenantRepo:      common.NewTenantRepo(db),
		ContractRepo:    common.NewContractRepo(db),
		PaymentRepo:     common.NewPaymentRepo(db),
		ReceiptRepo:     common.NewReceiptRepo(db),
		ReceiptBookRepo: common.NewReceiptBookRepo(db),
		TemplateRepo:    common.NewTemplateRepo(db),
		UserRepo:        common.NewUserRepo(db),
		ArrearsRepo:     common.NewArrearsRecordRepo(db),
	}
}
