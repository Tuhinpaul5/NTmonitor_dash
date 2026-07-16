package database

import (
	"gorm.io/driver/postgres"
	"NTMonitor/models"
	"gorm.io/gorm"

	"log"
	"fmt"
	"strings"
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

type EnumDef struct {
	Name   string
	Values []string
}

func createEnumTypes(db *gorm.DB) {
	enums := []EnumDef{
		{"user_status_enum", enumValues(models.UserStatusActive, models.UserStatusInactive, models.UserStatusSuspended, models.UserStatusPending)},
		{"user_type_enum", enumValues(models.UserTypeAdmin, models.UserTypeModerator, models.UserTypeUser, models.UserTypeGuest)},
		{"otp_type_enum", enumValues(models.OtpTypeRegister, models.OtpTypePasswordReset)},
		{"node_status_enum", enumValues(models.NodeStatusActive, models.NodeStatusInactive, models.NodeStatusIdle)},
	}

	for _, e := range enums {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = ?)", e.Name).Scan(&exists)
		if exists {
			continue
		}
		quoted := make([]string, len(e.Values))
		for i, v := range e.Values {
			quoted[i] = "'" + v + "'"
		}
		query := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", e.Name, strings.Join(quoted, ", "))
		if err := db.Exec(query).Error; err != nil {
			log.Printf("failed to create enum %s: %v", e.Name, err)
		}
	}
}

// enumValues converts any stringer-ish typed constants (NodeStatus, UserStatus, etc.)
// into a []string for building the ENUM DDL.
func enumValues[T ~string](vals ...T) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = string(v)
	}
	return out
}