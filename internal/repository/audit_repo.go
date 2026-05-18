package repository

import (
	"clinmitra/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditRepo struct {
	db *gorm.DB
}

// NewAuditRepository creates a GORM-backed AuditRepository implementation.
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepo{db: db}
}

// Create persists a new AuditLog entry, auto-generating a UUID if the ID is empty.
func (r *auditRepo) Create(log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	return r.db.Create(log).Error
}

// ListByEntity returns all audit entries for a specific entity (type+ID),
// ordered newest first.
func (r *auditRepo) ListByEntity(entityType, entityID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, WrapError(err)
}

// ListByUser returns audit entries for a specific user, limited to the given
// count, ordered newest first.
func (r *auditRepo) ListByUser(userID string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, WrapError(err)
}

// ListRecent returns the most recent audit entries system-wide, up to the
// given limit.
func (r *auditRepo) ListRecent(limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, WrapError(err)
}
