package sqlite

import (
	"asset-leasing-system/internal/domain"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Setup(dbPath string, adminPassword string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 启用 WAL 模式和优化 PRAGMA，提升并发性能
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA busy_timeout=5000").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA synchronous=NORMAL").Error; err != nil {
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
	if err := db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash admin password: %v", err)
		}
		if err := db.Create(&domain.User{Username: "admin", Password: string(hash), Role: "admin"}).Error; err != nil {
			return nil, err
		}
	}

	return db, nil
}
