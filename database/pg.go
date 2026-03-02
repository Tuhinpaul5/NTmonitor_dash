package database

import (
	"NTMonitor/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	// Auto-migrate the schemas
	db.AutoMigrate(&models.User{}, &models.UserData{})

	return db
}