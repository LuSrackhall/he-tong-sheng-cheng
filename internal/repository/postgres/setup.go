package postgres

import (
	"asset-leasing-system/internal/domain"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Setup(host, port, user, pass, dbname, sslmode, adminPassword string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, dbname, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 配置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

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
