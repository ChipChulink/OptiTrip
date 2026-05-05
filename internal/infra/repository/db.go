package repository

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"optitrip/internal/core/domain"
	"optitrip/internal/infra/config"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	log.Println("Database connected successfully")
	return nil
}

func Migrate() error {
	err := DB.AutoMigrate(
		&domain.City{},
		&domain.Category{},
		&domain.Place{},
		&domain.PlaceCategory{},
		&domain.PlaceSchedule{},
		&domain.TripRequest{},
		&domain.TripPlan{},
		&domain.CatalogVersion{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Println("Database migrated successfully")
	return nil
}
