package repository

import (
	"NTMonitor/models"
	"time"

	"gorm.io/gorm"
)

type SessionRepository struct {
	DB *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

func (r *SessionRepository) Create(session *models.UserSession) error {
	return r.DB.Create(session).Error
}

func (r *SessionRepository) FindByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.DB.Where("session_token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	return &session, err
}

func (r *SessionRepository) DeleteByToken(token string) error {
	return r.DB.Where("session_token = ?", token).Delete(&models.UserSession{}).Error
}

func (r *SessionRepository) DeleteExpired() error {
	return r.DB.Where("expires_at < ?", time.Now()).Delete(&models.UserSession{}).Error
}

func (r *SessionRepository) DeleteByUserID(userID string) error {
	return r.DB.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}
