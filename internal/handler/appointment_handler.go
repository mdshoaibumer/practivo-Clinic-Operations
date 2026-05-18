package handler

import (
	"log/slog"

	"clinmitra/internal/models"
	"clinmitra/internal/service"
)

type AppointmentHandler struct {
	appointmentService *service.AppointmentService
}

// NewAppointmentHandler creates an AppointmentHandler backed by the given service.
func NewAppointmentHandler(appointmentService *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{appointmentService: appointmentService}
}

// CreateAppointment schedules a new appointment after conflict checking.
func (h *AppointmentHandler) CreateAppointment(input service.CreateAppointmentInput) (*models.Appointment, error) {
	slog.Info("creating appointment", "patientId", input.PatientID, "date", input.Date, "start", input.StartTime, "end", input.EndTime)
	result, err := h.appointmentService.CreateAppointment(input)
	if err != nil {
		slog.Warn("create appointment failed", "patientId", input.PatientID, "date", input.Date, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("appointment created", "id", result.ID, "patientId", input.PatientID, "date", input.Date)
	return result, nil
}

// UpdateAppointment modifies an existing scheduled appointment.
func (h *AppointmentHandler) UpdateAppointment(id string, input service.CreateAppointmentInput) (*models.Appointment, error) {
	slog.Info("updating appointment", "id", id, "date", input.Date, "start", input.StartTime)
	result, err := h.appointmentService.UpdateAppointment(id, input)
	if err != nil {
		slog.Warn("update appointment failed", "id", id, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("appointment updated", "id", id)
	return result, nil
}

// CancelAppointment marks an appointment as cancelled with a reason.
func (h *AppointmentHandler) CancelAppointment(id, reason string) error {
	slog.Info("cancelling appointment", "id", id, "reason", reason)
	err := h.appointmentService.CancelAppointment(id, reason)
	if err != nil {
		slog.Warn("cancel appointment failed", "id", id, "error", err.Error())
		return safeError(err)
	}
	slog.Info("appointment cancelled", "id", id)
	return nil
}

// CompleteAppointment transitions a scheduled appointment to completed.
func (h *AppointmentHandler) CompleteAppointment(id string) error {
	slog.Info("completing appointment", "id", id)
	err := h.appointmentService.CompleteAppointment(id)
	if err != nil {
		slog.Warn("complete appointment failed", "id", id, "error", err.Error())
		return safeError(err)
	}
	slog.Info("appointment completed", "id", id)
	return nil
}

// GetTodayAppointments returns all appointments for today.
func (h *AppointmentHandler) GetTodayAppointments() ([]models.Appointment, error) {
	result, err := h.appointmentService.GetTodayAppointments()
	if err != nil {
		slog.Error("get today appointments failed", "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// GetAppointmentsByDate returns all appointments for a specific date.
func (h *AppointmentHandler) GetAppointmentsByDate(date string) ([]models.Appointment, error) {
	result, err := h.appointmentService.GetAppointmentsByDate(date)
	if err != nil {
		slog.Error("get appointments by date failed", "date", date, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// GetWeekAppointments returns appointments within a date range.
func (h *AppointmentHandler) GetWeekAppointments(startDate, endDate string) ([]models.Appointment, error) {
	result, err := h.appointmentService.GetWeekAppointments(startDate, endDate)
	if err != nil {
		slog.Error("get week appointments failed", "startDate", startDate, "endDate", endDate, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// GetPatientAppointments returns all appointments for a specific patient.
func (h *AppointmentHandler) GetPatientAppointments(patientID string) ([]models.Appointment, error) {
	result, err := h.appointmentService.GetPatientAppointments(patientID)
	if err != nil {
		slog.Error("get patient appointments failed", "patientId", patientID, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// GetAppointment retrieves a single appointment by ID.
func (h *AppointmentHandler) GetAppointment(id string) (*models.Appointment, error) {
	result, err := h.appointmentService.GetAppointment(id)
	if err != nil {
		slog.Debug("get appointment failed", "id", id, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}
