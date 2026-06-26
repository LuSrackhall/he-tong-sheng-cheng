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
	if err := db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		seedPass := os.Getenv("ADMIN_PASSWORD")
		if seedPass == "" {
			seedPass = "admin123"
			log.Println("WARNING: Using default admin password. Set ADMIN_PASSWORD env var for production.")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(seedPass), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		if err := db.Create(&domain.User{Username: "admin", Password: string(hash), Role: "admin"}).Error; err != nil {
			return nil, err
		}
	}

	return db, nil
}
