package service

import (
	"log/slog"

	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreatePatientInput struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Gender         string `json:"gender"`
	Age            int    `json:"age"`
	DateOfBirth    string `json:"dateOfBirth"`
	Address        string `json:"address"`
	City           string `json:"city"`
	BloodGroup     string `json:"bloodGroup"`
	MedicalHistory string `json:"medicalHistory"`
	Allergies      string `json:"allergies"`
	Notes          string `json:"notes"`
}

type PatientListResponse struct {
	Patients []models.Patient `json:"patients"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type PatientService struct {
	patientRepo          repository.PatientRepository
	patientTreatmentRepo repository.PatientTreatmentRepository
	invoiceRepo          repository.InvoiceRepository
	authService          *AuthService
	auditService         *AuditService
	db                   *gorm.DB
}

// NewPatientService creates a PatientService with all required dependencies.
func NewPatientService(
	patientRepo repository.PatientRepository,
	patientTreatmentRepo repository.PatientTreatmentRepository,
	invoiceRepo repository.InvoiceRepository,
	authService *AuthService,
	auditService *AuditService,
	db *gorm.DB,
) *PatientService {
	return &PatientService{
		patientRepo:          patientRepo,
		patientTreatmentRepo: patientTreatmentRepo,
		invoiceRepo:          invoiceRepo,
		authService:          authService,
		auditService:         auditService,
		db:                   db,
	}
}

// CreatePatient validates input (name, phone, age) and persists a new
// patient record inside a transaction. Logs the action to the audit trail.
func (s *PatientService) CreatePatient(input CreatePatientInput) (*models.Patient, error) {
	if err := utils.ValidateRequired("Name", input.Name); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("Phone", input.Phone); err != nil {
		return nil, err
	}
	if err := utils.ValidateMinLength("Name", input.Name, 2); err != nil {
		return nil, err
	}

	cleanedPhone := utils.CleanPhone(input.Phone)
	if err := utils.ValidatePhone(cleanedPhone); err != nil {
		return nil, err
	}

	// Check for duplicate phone number
	existing, err := s.patientRepo.FindByPhone(cleanedPhone)
	if err == nil && existing != nil {
		return nil, utils.ValidationError("A patient with this phone number already exists")
	}

	if input.Age != 0 {
		if err := utils.ValidateAge(input.Age); err != nil {
			return nil, err
		}
	}

	patient := &models.Patient{
		BaseModel:      models.BaseModel{ID: uuid.New().String()},
		Name:           input.Name,
		Phone:          cleanedPhone,
		Email:          input.Email,
		Gender:         models.Gender(input.Gender),
		Age:            input.Age,
		DateOfBirth:    input.DateOfBirth,
		Address:        input.Address,
		City:           input.City,
		BloodGroup:     input.BloodGroup,
		MedicalHistory: input.MedicalHistory,
		Allergies:      input.Allergies,
		Notes:          input.Notes,
		CreatedBy:      s.authService.GetCurrentUserID(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(patient).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		slog.Error("patient creation transaction failed", "name", input.Name, "phone", cleanedPhone, "error", err.Error())
		return nil, err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditCreate, "patient", patient.ID, nil, patient)
	return patient, nil
}

// UpdatePatient validates input and updates an existing patient record.
// Records the old and new values in the audit trail.
func (s *PatientService) UpdatePatient(id string, input CreatePatientInput) (*models.Patient, error) {
	patient, err := s.patientRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if err := utils.ValidateRequired("Name", input.Name); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("Phone", input.Phone); err != nil {
		return nil, err
	}

	cleanedPhone := utils.CleanPhone(input.Phone)
	if err := utils.ValidatePhone(cleanedPhone); err != nil {
		return nil, err
	}

	// Check for duplicate phone number (exclude current patient)
	if cleanedPhone != patient.Phone {
		existing, err := s.patientRepo.FindByPhone(cleanedPhone)
		if err == nil && existing != nil && existing.ID != id {
			return nil, utils.ValidationError("A patient with this phone number already exists")
		}
	}

	old := *patient
	patient.Name = input.Name
	patient.Phone = cleanedPhone
	patient.Email = input.Email
	patient.Gender = models.Gender(input.Gender)
	patient.Age = input.Age
	patient.DateOfBirth = input.DateOfBirth
	patient.Address = input.Address
	patient.City = input.City
	patient.BloodGroup = input.BloodGroup
	patient.MedicalHistory = input.MedicalHistory
	patient.Allergies = input.Allergies
	patient.Notes = input.Notes

	err = s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(patient).Updates(map[string]interface{}{
			"name":            patient.Name,
			"phone":           patient.Phone,
			"email":           patient.Email,
			"gender":          patient.Gender,
			"age":             patient.Age,
			"date_of_birth":   patient.DateOfBirth,
			"address":         patient.Address,
			"city":            patient.City,
			"blood_group":     patient.BloodGroup,
			"medical_history": patient.MedicalHistory,
			"allergies":       patient.Allergies,
			"notes":           patient.Notes,
		}).Error
	})
	if err != nil {
		slog.Error("patient update transaction failed", "id", id, "error", err.Error())
		return nil, err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "patient", id, old, patient)
	return patient, nil
}

// GetPatient retrieves a patient by ID. Returns ErrNotFound if not found.
func (s *PatientService) GetPatient(id string) (*models.Patient, error) {
	patient, err := s.patientRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return patient, nil
}

// ListPatients returns a paginated list of patients, optionally filtered
// by a search string (matches name or phone).
func (s *PatientService) ListPatients(page, pageSize int, search string) (*PatientListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	patients, total, err := s.patientRepo.List(page, pageSize, search)
	if err != nil {
		return nil, err
	}

	return &PatientListResponse{
		Patients: patients,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// DeletePatient soft-deletes a patient by ID. Prevents deletion if the
// patient has unpaid invoices. Requires admin role.
func (s *PatientService) DeletePatient(id string) error {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return err
	}

	// Check for unpaid invoices
	outstanding, err := s.invoiceRepo.GetOutstandingByPatient(id)
	if err != nil {
		return err
	}
	if outstanding > 0 {
		return utils.ValidationError("Cannot delete patient with unpaid invoices")
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditDelete, "patient", id, nil, nil)
	return s.patientRepo.Delete(id)
}

// GetPatientHistory returns all treatment records for the given patient.
func (s *PatientService) GetPatientHistory(patientID string) ([]models.PatientTreatment, error) {
	return s.patientTreatmentRepo.ListByPatient(patientID)
}

// CheckDuplicatePhone looks up a patient by phone number to detect duplicates.
func (s *PatientService) CheckDuplicatePhone(phone string) (*models.Patient, error) {
	cleanedPhone := utils.CleanPhone(phone)
	return s.patientRepo.FindByPhone(cleanedPhone)
}

// GetPatientCount returns the total number of patients in the database.
func (s *PatientService) GetPatientCount() (int64, error) {
	return s.patientRepo.Count()
}
