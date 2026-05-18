package repository

import (
	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type treatmentRepo struct {
	db *gorm.DB
}

// NewTreatmentRepository creates a GORM-backed TreatmentRepository implementation.
func NewTreatmentRepository(db *gorm.DB) TreatmentRepository {
	return &treatmentRepo{db: db}
}

// Create persists a new Treatment record to the database.
func (r *treatmentRepo) Create(treatment *models.Treatment) error {
	return r.db.Create(treatment).Error
}

// FindByID retrieves a treatment by primary key.
func (r *treatmentRepo) FindByID(id string) (*models.Treatment, error) {
	var treatment models.Treatment
	err := r.db.Where("id = ?", id).First(&treatment).Error
	if err != nil {
		return nil, WrapError(err)
	}
	return &treatment, nil
}

// Update saves changed fields on an existing treatment record.
func (r *treatmentRepo) Update(treatment *models.Treatment) error {
	return r.db.Model(treatment).Updates(treatment).Error
}

// Delete soft-deletes a treatment by setting is_active=false rather than
// removing the row (preserves referential integrity for historical invoices).
func (r *treatmentRepo) Delete(id string) error {
	return r.db.Model(&models.Treatment{}).Where("id = ?", id).Update("is_active", false).Error
}

// ListActive returns all treatments where is_active=true, sorted by category and name.
func (r *treatmentRepo) ListActive() ([]models.Treatment, error) {
	var treatments []models.Treatment
	err := r.db.Where("is_active = ?", true).Order("category, name").Find(&treatments).Error
	return treatments, WrapError(err)
}

// ListAll returns all treatments (including inactive), sorted by category and name.
func (r *treatmentRepo) ListAll() ([]models.Treatment, error) {
	var treatments []models.Treatment
	err := r.db.Order("category, name").Find(&treatments).Error
	return treatments, WrapError(err)
}
