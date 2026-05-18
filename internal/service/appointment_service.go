package service

import (
	"log/slog"

	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateAppointmentInput holds the validated input fields for creating or
// updating an appointment.
type CreateAppointmentInput struct {
	PatientID string `json:"patientId"`
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Duration  int    `json:"duration"`
	Purpose   string `json:"purpose"`
	Notes     string `json:"notes"`
}

// AppointmentService implements business logic for scheduling, updating,
// cancelling, and completing appointments.
type AppointmentService struct {
	appointmentRepo repository.AppointmentRepository
	patientRepo     repository.PatientRepository
	authService     *AuthService
	auditService    *AuditService
	db              *gorm.DB
}

// NewAppointmentService creates an AppointmentService with all required dependencies.
func NewAppointmentService(
	appointmentRepo repository.AppointmentRepository,
	patientRepo repository.PatientRepository,
	authService *AuthService,
	auditService *AuditService,
	db *gorm.DB,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
		authService:     authService,
		auditService:    auditService,
		db:              db,
	}
}

// CreateAppointment validates input, checks for time-slot conflicts, and
// persists a new appointment in a database transaction. Logs the action to
// the audit trail.
func (s *AppointmentService) CreateAppointment(input CreateAppointmentInput) (*models.Appointment, error) {
	if err := utils.ValidateRequired("Patient", input.PatientID); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("Date", input.Date); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("Start time", input.StartTime); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("End time", input.EndTime); err != nil {
		return nil, err
	}

	// Verify patient exists
	if _, err := s.patientRepo.FindByID(input.PatientID); err != nil {
		return nil, utils.ValidationError("Patient not found")
	}

	// Check for conflicts
	conflict, err := s.appointmentRepo.FindConflicting(input.Date, input.StartTime, input.EndTime, "")
	if err != nil {
		return nil, err
	}
	if conflict != nil {
		return nil, utils.ValidationError("Time slot conflicts with an existing appointment")
	}

	duration := input.Duration
	if duration == 0 {
		duration = 30
	}

	appointment := &models.Appointment{
		BaseModel:       models.BaseModel{ID: uuid.New().String()},
		PatientID:       input.PatientID,
		AppointmentDate: input.Date,
		StartTime:       input.StartTime,
		EndTime:         input.EndTime,
		Duration:        duration,
		Status:          models.AppointmentScheduled,
		Purpose:         input.Purpose,
		Notes:           input.Notes,
		CreatedBy:       s.authService.GetCurrentUserID(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(appointment).Error
	})
	if err != nil {
		slog.Error("appointment creation transaction failed",
			"patientId", input.PatientID,
			"date", input.Date,
			"startTime", input.StartTime,
			"error", err.Error(),
		)
		return nil, err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditCreate, "appointment", appointment.ID, nil, appointment)
	return appointment, nil
}

// UpdateAppointment modifies a scheduled appointment's details after validating
// required fields and checking for time-slot conflicts (excluding itself).
// Only scheduled appointments can be edited.
func (s *AppointmentService) UpdateAppointment(id string, input CreateAppointmentInput) (*models.Appointment, error) {
	appointment, err := s.appointmentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if appointment.Status != models.AppointmentScheduled {
		return nil, utils.ValidationError("Can only edit scheduled appointments")
	}

	// Validate required fields
	if err := utils.ValidateRequired("Patient", input.PatientID); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("Date", input.Date); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("Start time", input.StartTime); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("End time", input.EndTime); err != nil {
		return nil, err
	}

	// Check conflicts excluding current appointment
	conflict, err := s.appointmentRepo.FindConflicting(input.Date, input.StartTime, input.EndTime, id)
	if err != nil {
		return nil, err
	}
	if conflict != nil {
		return nil, utils.ValidationError("Time slot conflicts with an existing appointment")
	}

	old := *appointment
	appointment.PatientID = input.PatientID
	appointment.AppointmentDate = input.Date
	appointment.StartTime = input.StartTime
	appointment.EndTime = input.EndTime
	appointment.Duration = input.Duration
	appointment.Purpose = input.Purpose
	appointment.Notes = input.Notes

	if err := s.appointmentRepo.Update(appointment); err != nil {
		return nil, err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "appointment", id, old, appointment)
	return appointment, nil
}

// CancelAppointment marks an appointment as cancelled with the given reason.
// Completed appointments cannot be cancelled.
func (s *AppointmentService) CancelAppointment(id, reason string) error {
	appointment, err := s.appointmentRepo.FindByID(id)
	if err != nil {
		return err
	}

	if appointment.Status == models.AppointmentCompleted {
		return utils.ValidationError("Cannot cancel a completed appointment")
	}

	appointment.Status = models.AppointmentCancelled
	appointment.CancelReason = reason

	if err := s.appointmentRepo.Update(appointment); err != nil {
		return err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "appointment", id, nil, map[string]string{
		"action": "cancelled",
		"reason": reason,
	})
	return nil
}

// CompleteAppointment transitions a scheduled appointment to completed status.
func (s *AppointmentService) CompleteAppointment(id string) error {
	appointment, err := s.appointmentRepo.FindByID(id)
	if err != nil {
		return err
	}

	if appointment.Status != models.AppointmentScheduled {
		return utils.ValidationError("Only scheduled appointments can be completed")
	}

	appointment.Status = models.AppointmentCompleted

	if err := s.appointmentRepo.Update(appointment); err != nil {
		return err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "appointment", id, nil, map[string]string{
		"action": "completed",
	})
	return nil
}

// GetTodayAppointments returns all appointments for today's date.
func (s *AppointmentService) GetTodayAppointments() ([]models.Appointment, error) {
	return s.appointmentRepo.ListByDate(utils.TodayDate())
}

// GetAppointmentsByDate returns all appointments for the specified date.
func (s *AppointmentService) GetAppointmentsByDate(date string) ([]models.Appointment, error) {
	return s.appointmentRepo.ListByDate(date)
}

// GetWeekAppointments returns all appointments within a date range (inclusive).
func (s *AppointmentService) GetWeekAppointments(startDate, endDate string) ([]models.Appointment, error) {
	return s.appointmentRepo.ListByDateRange(startDate, endDate)
}

// GetPatientAppointments returns all appointments for a specific patient.
func (s *AppointmentService) GetPatientAppointments(patientID string) ([]models.Appointment, error) {
	return s.appointmentRepo.ListByPatient(patientID)
}

// GetAppointment retrieves a single appointment by ID with its patient relation.
func (s *AppointmentService) GetAppointment(id string) (*models.Appointment, error) {
	return s.appointmentRepo.FindByID(id)
}
