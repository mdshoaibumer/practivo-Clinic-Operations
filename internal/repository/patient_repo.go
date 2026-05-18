package repository

import (
	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type patientRepo struct {
	db *gorm.DB
}

// NewPatientRepository creates a GORM-backed PatientRepository implementation.
func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepo{db: db}
}

// Create persists a new Patient record to the database.
func (r *patientRepo) Create(patient *models.Patient) error {
	return r.db.Create(patient).Error
}

// FindByID retrieves a patient by primary key. Returns utils.ErrNotFound
// if no matching record exists.
func (r *patientRepo) FindByID(id string) (*models.Patient, error) {
	var patient models.Patient
	err := r.db.Where("id = ?", id).First(&patient).Error
	if err != nil {
		return nil, WrapError(err)
	}
	return &patient, nil
}

// Update saves changed fields on an existing patient record.
func (r *patientRepo) Update(patient *models.Patient) error {
	return r.db.Model(patient).Updates(patient).Error
}

// Delete soft-deletes a patient by ID (sets deleted_at timestamp).
func (r *patientRepo) Delete(id string) error {
	return r.db.Delete(&models.Patient{}, "id = ?", id).Error
}

// List returns a paginated list of patients, optionally filtered by a search
// term that matches against name or phone using LIKE.
func (r *patientRepo) List(page, pageSize int, search string) ([]models.Patient, int64, error) {
	var patients []models.Patient
	var total int64

	query := r.db.Model(&models.Patient{})

	if search != "" {
		query = query.Where("name LIKE ? OR phone LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, WrapError(err)
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&patients).Error
	if err != nil {
		return nil, 0, WrapError(err)
	}

	return patients, total, nil
}

// FindByPhone looks up a patient by exact phone number match.
// Returns utils.ErrNotFound if no match.
func (r *patientRepo) FindByPhone(phone string) (*models.Patient, error) {
	var patient models.Patient
	err := r.db.Where("phone = ?", phone).First(&patient).Error
	if err != nil {
		return nil, WrapError(err)
	}
	return &patient, nil
}

// Count returns the total number of non-deleted patients.
func (r *patientRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Patient{}).Count(&count).Error
	return count, err
}
