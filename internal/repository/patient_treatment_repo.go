package repository

import (
	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type patientTreatmentRepo struct {
	db *gorm.DB
}

// NewPatientTreatmentRepository creates a GORM-backed PatientTreatmentRepository implementation.
func NewPatientTreatmentRepository(db *gorm.DB) PatientTreatmentRepository {
	return &patientTreatmentRepo{db: db}
}

// Create persists a single PatientTreatment record.
func (r *patientTreatmentRepo) Create(pt *models.PatientTreatment) error {
	return r.db.Create(pt).Error
}

// CreateBatch inserts multiple PatientTreatment records in batches of 10.
func (r *patientTreatmentRepo) CreateBatch(pts []models.PatientTreatment) error {
	return r.db.CreateInBatches(pts, 10).Error
}

// ListByPatient returns all treatment history for a patient with the Treatment
// relation preloaded, ordered by treatment date (newest first).
func (r *patientTreatmentRepo) ListByPatient(patientID string) ([]models.PatientTreatment, error) {
	var treatments []models.PatientTreatment
	err := r.db.Preload("Treatment").
		Where("patient_id = ?", patientID).
		Order("treatment_date DESC").
		Find(&treatments).Error
	return treatments, WrapError(err)
}
