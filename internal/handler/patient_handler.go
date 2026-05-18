package handler

import (
	"log/slog"

	"clinmitra/internal/models"
	"clinmitra/internal/service"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxSearchLength = 100
)

type PatientHandler struct {
	patientService *service.PatientService
}

// NewPatientHandler creates a PatientHandler backed by the given PatientService.
func NewPatientHandler(patientService *service.PatientService) *PatientHandler {
	return &PatientHandler{patientService: patientService}
}

// CreatePatient validates and creates a new patient record.
func (h *PatientHandler) CreatePatient(input service.CreatePatientInput) (*models.Patient, error) {
	slog.Info("creating patient", "name", input.Name, "phone", input.Phone)
	result, err := h.patientService.CreatePatient(input)
	if err != nil {
		slog.Warn("create patient failed", "name", input.Name, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("patient created", "id", result.ID, "name", result.Name)
	return result, nil
}

// UpdatePatient updates an existing patient record by ID.
func (h *PatientHandler) UpdatePatient(id string, input service.CreatePatientInput) (*models.Patient, error) {
	slog.Info("updating patient", "id", id, "name", input.Name)
	result, err := h.patientService.UpdatePatient(id, input)
	if err != nil {
		slog.Warn("update patient failed", "id", id, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("patient updated", "id", id)
	return result, nil
}

// GetPatient retrieves a single patient by ID.
func (h *PatientHandler) GetPatient(id string) (*models.Patient, error) {
	result, err := h.patientService.GetPatient(id)
	if err != nil {
		slog.Debug("get patient failed", "id", id, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// ListPatients returns a paginated, optionally filtered list of patients.
func (h *PatientHandler) ListPatients(page, pageSize int, search string) (*service.PatientListResponse, error) {
	page, pageSize = sanitizePagination(page, pageSize)
	search = sanitizeSearch(search)
	result, err := h.patientService.ListPatients(page, pageSize, search)
	if err != nil {
		slog.Error("list patients failed", "page", page, "search", search, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// DeletePatient soft-deletes a patient by ID (blocked if unpaid invoices exist).
func (h *PatientHandler) DeletePatient(id string) error {
	slog.Info("deleting patient", "id", id)
	err := h.patientService.DeletePatient(id)
	if err != nil {
		slog.Warn("delete patient failed", "id", id, "error", err.Error())
		return safeError(err)
	}
	slog.Info("patient deleted", "id", id)
	return nil
}

// GetPatientHistory returns the treatment history for a patient.
func (h *PatientHandler) GetPatientHistory(patientID string) ([]models.PatientTreatment, error) {
	result, err := h.patientService.GetPatientHistory(patientID)
	return result, safeError(err)
}

// CheckDuplicatePhone checks if a phone number is already registered.
func (h *PatientHandler) CheckDuplicatePhone(phone string) (*models.Patient, error) {
	result, err := h.patientService.CheckDuplicatePhone(phone)
	return result, safeError(err)
}

// GetPatientCount returns the total number of patients in the system.
func (h *PatientHandler) GetPatientCount() (int64, error) {
	result, err := h.patientService.GetPatientCount()
	return result, safeError(err)
}

// sanitizePagination enforces safe defaults and limits on pagination parameters.
func sanitizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// sanitizeSearch truncates search strings to prevent excessively long LIKE queries.
func sanitizeSearch(search string) string {
	if len(search) > maxSearchLength {
		return search[:maxSearchLength]
	}
	return search
}
