package repository

import (
	"NTMonitor/models"

	"gorm.io/gorm"
)

type OtpRepository struct {
	DB *gorm.DB
}

func NewOtpRepository(db *gorm.DB) *OtpRepository {
	return &OtpRepository{DB: db}
}

func (r *OtpRepository) Create(otp *models.OTP) error {
	return r.DB.Create(otp).Error
}

func (r *OtpRepository) Get(otpID string) (*models.OTP, error) {
	var otp models.OTP
	err := r.DB.First(&otp, "id = ?", otpID).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}