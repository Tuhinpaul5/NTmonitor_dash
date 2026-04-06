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

	// Create enum types before migration
	createEnumTypes(db)

	// Auto-migrate the schemas
	db.AutoMigrate(
		&models.User{},
		&models.UserData{},
		&models.OTP{},
		&models.UserSession{},
	)

	return db
}

func createEnumTypes(db *gorm.DB) {
	// Check and create user_status_enum
	var statusExists bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status_enum')").Scan(&statusExists)
	if !statusExists {
		db.Exec("CREATE TYPE user_status_enum AS ENUM ('active', 'inactive', 'suspended', 'pending')")
	}

	// Check and create user_type_enum
	var typeExists bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type_enum')").Scan(&typeExists)
	if !typeExists {
		db.Exec("CREATE TYPE user_type_enum AS ENUM ('admin', 'moderator', 'user', 'guest')")
	}

	// Check and create otp_type_enum
	var otpTypeExists bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'otp_type_enum')").Scan(&otpTypeExists)
	if !otpTypeExists {
		db.Exec("CREATE TYPE otp_type_enum AS ENUM ('register', 'password_reset')")
	}
}
