package repository

import (
	"errors"

	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type appointmentRepo struct {
	db *gorm.DB
}

// NewAppointmentRepository creates a GORM-backed AppointmentRepository implementation.
func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepo{db: db}
}

// Create persists a new Appointment record to the database.
func (r *appointmentRepo) Create(appointment *models.Appointment) error {
	return r.db.Create(appointment).Error
}

// FindByID retrieves an appointment by ID with its Patient relation preloaded.
func (r *appointmentRepo) FindByID(id string) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("Patient").Where("id = ?", id).First(&appointment).Error
	if err != nil {
		return nil, WrapError(err)
	}
	return &appointment, nil
}

// Update saves changed fields on an existing appointment using an explicit
// column map to ensure zero-value fields (like empty cancel_reason) are persisted.
func (r *appointmentRepo) Update(appointment *models.Appointment) error {
	return r.db.Model(appointment).Updates(map[string]interface{}{
		"patient_id":       appointment.PatientID,
		"appointment_date": appointment.AppointmentDate,
		"start_time":       appointment.StartTime,
		"end_time":         appointment.EndTime,
		"duration":         appointment.Duration,
		"status":           appointment.Status,
		"purpose":          appointment.Purpose,
		"notes":            appointment.Notes,
		"cancel_reason":    appointment.CancelReason,
	}).Error
}

// Delete soft-deletes an appointment by ID.
func (r *appointmentRepo) Delete(id string) error {
	return r.db.Delete(&models.Appointment{}, "id = ?", id).Error
}

// ListByDate returns all appointments for a given date, ordered by start time,
// with the Patient relation preloaded.
func (r *appointmentRepo) ListByDate(date string) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Preload("Patient").
		Where("appointment_date = ?", date).
		Order("start_time ASC").
		Find(&appointments).Error
	return appointments, WrapError(err)
}

// ListByDateRange returns all appointments between startDate and endDate
// (inclusive), ordered by date then time, with Patient preloaded.
func (r *appointmentRepo) ListByDateRange(startDate, endDate string) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Preload("Patient").
		Where("appointment_date >= ? AND appointment_date <= ?", startDate, endDate).
		Order("appointment_date ASC, start_time ASC").
		Find(&appointments).Error
	return appointments, WrapError(err)
}

// ListByPatient returns all appointments for a patient, newest first.
func (r *appointmentRepo) ListByPatient(patientID string) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("patient_id = ?", patientID).
		Order("appointment_date DESC, start_time DESC").
		Find(&appointments).Error
	return appointments, WrapError(err)
}

// FindConflicting detects time-slot overlaps for scheduled appointments on a
// given date. If excludeID is non-empty, that appointment is excluded from the
// check (used when editing an existing appointment). Returns nil if no conflict.
func (r *appointmentRepo) FindConflicting(date, startTime, endTime, excludeID string) (*models.Appointment, error) {
	var appointment models.Appointment
	query := r.db.Where("appointment_date = ? AND status = ? AND ((start_time < ? AND end_time > ?) OR (start_time >= ? AND start_time < ?))",
		date, models.AppointmentScheduled, endTime, startTime, startTime, endTime)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	err := query.First(&appointment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, WrapError(err)
	}
	return &appointment, nil
}

// CountByDate returns the number of scheduled appointments for a given date.
func (r *appointmentRepo) CountByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Appointment{}).
		Where("appointment_date = ? AND status = ?", date, models.AppointmentScheduled).
		Count(&count).Error
	return count, err
}
