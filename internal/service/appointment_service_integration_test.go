package service

import (
	"errors"
	"testing"

	"clinmitra/internal/auth"
	"clinmitra/internal/models"
	"clinmitra/internal/utils"
)

// --- Mock Appointment Repository ---

type mockAppointmentRepoForService struct {
	appointments map[string]*models.Appointment
	conflicting  *models.Appointment
}

func newMockAppointmentRepoForService() *mockAppointmentRepoForService {
	return &mockAppointmentRepoForService{
		appointments: make(map[string]*models.Appointment),
	}
}

func (m *mockAppointmentRepoForService) Create(appt *models.Appointment) error {
	m.appointments[appt.ID] = appt
	return nil
}

func (m *mockAppointmentRepoForService) FindByID(id string) (*models.Appointment, error) {
	a, ok := m.appointments[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return a, nil
}

func (m *mockAppointmentRepoForService) Update(appt *models.Appointment) error {
	m.appointments[appt.ID] = appt
	return nil
}

func (m *mockAppointmentRepoForService) Delete(id string) error {
	delete(m.appointments, id)
	return nil
}

func (m *mockAppointmentRepoForService) ListByDate(date string) ([]models.Appointment, error) {
	var results []models.Appointment
	for _, a := range m.appointments {
		if a.AppointmentDate == date {
			results = append(results, *a)
		}
	}
	return results, nil
}

func (m *mockAppointmentRepoForService) ListByDateRange(start, end string) ([]models.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentRepoForService) ListByPatient(patientID string) ([]models.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentRepoForService) FindConflicting(date, startTime, endTime, excludeID string) (*models.Appointment, error) {
	return m.conflicting, nil
}

func (m *mockAppointmentRepoForService) CountByDate(date string) (int64, error) {
	var count int64
	for _, a := range m.appointments {
		if a.AppointmentDate == date && a.Status == models.AppointmentScheduled {
			count++
		}
	}
	return count, nil
}

// --- Mock Patient Repo for Appointment Tests ---

type mockPatientRepoForAppointment struct {
	patients map[string]*models.Patient
}

func newMockPatientRepoForAppointment() *mockPatientRepoForAppointment {
	return &mockPatientRepoForAppointment{patients: make(map[string]*models.Patient)}
}

func (m *mockPatientRepoForAppointment) Create(p *models.Patient) error { return nil }
func (m *mockPatientRepoForAppointment) FindByID(id string) (*models.Patient, error) {
	p, ok := m.patients[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return p, nil
}
func (m *mockPatientRepoForAppointment) Update(p *models.Patient) error { return nil }
func (m *mockPatientRepoForAppointment) Delete(id string) error         { return nil }
func (m *mockPatientRepoForAppointment) List(page, pageSize int, search string) ([]models.Patient, int64, error) {
	return nil, 0, nil
}
func (m *mockPatientRepoForAppointment) FindByPhone(phone string) (*models.Patient, error) {
	return nil, utils.ErrNotFound
}
func (m *mockPatientRepoForAppointment) Count() (int64, error) { return 0, nil }

// --- Mock Audit Repo for Tests ---

type mockAuditRepoForApptTests struct{}

func (m *mockAuditRepoForApptTests) Create(log *models.AuditLog) error { return nil }
func (m *mockAuditRepoForApptTests) ListByEntity(entityType, entityID string) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditRepoForApptTests) ListByUser(userID string, limit int) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditRepoForApptTests) ListRecent(limit int) ([]models.AuditLog, error) {
	return nil, nil
}

func newTestAuditService() *AuditService {
	return NewAuditService(&mockAuditRepoForApptTests{})
}

// --- Appointment Service Tests ---

func TestAppointmentService_CancelAppointment_NotFound(t *testing.T) {
	apptRepo := newMockAppointmentRepoForService()

	svc := &AppointmentService{
		appointmentRepo: apptRepo,
		patientRepo:     newMockPatientRepoForAppointment(),
		authService:     &AuthService{sessionManager: auth.NewSessionManager(8)},
		auditService:    newTestAuditService(),
	}

	err := svc.CancelAppointment("nonexistent", "patient request")
	if err == nil {
		t.Fatal("expected error for nonexistent appointment")
	}

	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T (%v)", err, err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got: %s", appErr.Code)
	}
}

func TestAppointmentService_CancelAppointment_AlreadyCompleted(t *testing.T) {
	apptRepo := newMockAppointmentRepoForService()
	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		Status:    models.AppointmentCompleted,
	}

	svc := &AppointmentService{
		appointmentRepo: apptRepo,
		patientRepo:     newMockPatientRepoForAppointment(),
		authService:     &AuthService{sessionManager: auth.NewSessionManager(8)},
		auditService:    newTestAuditService(),
	}

	err := svc.CancelAppointment("appt-1", "patient request")
	if err == nil {
		t.Fatal("expected error when cancelling completed appointment")
	}

	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T (%v)", err, err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestAppointmentService_CancelAppointment_Success(t *testing.T) {
	apptRepo := newMockAppointmentRepoForService()
	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		Status:    models.AppointmentScheduled,
		PatientID: "patient-1",
	}

	svc := &AppointmentService{
		appointmentRepo: apptRepo,
		patientRepo:     newMockPatientRepoForAppointment(),
		authService:     &AuthService{sessionManager: auth.NewSessionManager(8)},
		auditService:    newTestAuditService(),
	}

	err := svc.CancelAppointment("appt-1", "patient request")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify status was updated
	appt := apptRepo.appointments["appt-1"]
	if appt.Status != models.AppointmentCancelled {
		t.Errorf("expected status cancelled, got: %s", appt.Status)
	}
	if appt.CancelReason != "patient request" {
		t.Errorf("expected cancel reason 'patient request', got: %s", appt.CancelReason)
	}
}

func TestAppointmentService_CompleteAppointment_NotScheduled(t *testing.T) {
	apptRepo := newMockAppointmentRepoForService()
	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		Status:    models.AppointmentCancelled,
	}

	svc := &AppointmentService{
		appointmentRepo: apptRepo,
		patientRepo:     newMockPatientRepoForAppointment(),
		authService:     &AuthService{sessionManager: auth.NewSessionManager(8)},
		auditService:    newTestAuditService(),
	}

	err := svc.CompleteAppointment("appt-1")
	if err == nil {
		t.Fatal("expected error when completing non-scheduled appointment")
	}

	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T (%v)", err, err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestAppointmentService_CreateAppointment_PatientNotFound(t *testing.T) {
	apptRepo := newMockAppointmentRepoForService()
	patientRepo := newMockPatientRepoForAppointment()

	svc := &AppointmentService{
		appointmentRepo: apptRepo,
		patientRepo:     patientRepo,
		authService:     &AuthService{sessionManager: auth.NewSessionManager(8)},
		auditService:    newTestAuditService(),
	}

	input := CreateAppointmentInput{
		PatientID: "nonexistent",
		Date:      "2026-05-16",
		StartTime: "10:00",
		EndTime:   "10:30",
	}

	_, err := svc.CreateAppointment(input)
	if err == nil {
		t.Fatal("expected error for nonexistent patient")
	}

	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T (%v)", err, err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestAppointmentService_CreateAppointment_TimeConflict(t *testing.T) {
	apptRepo := newMockAppointmentRepoForService()
	apptRepo.conflicting = &models.Appointment{
		BaseModel: models.BaseModel{ID: "existing-appt"},
		StartTime: "10:00",
		EndTime:   "10:30",
	}

	patientRepo := newMockPatientRepoForAppointment()
	patientRepo.patients["patient-1"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "patient-1"},
		Name:      "Test Patient",
	}

	svc := &AppointmentService{
		appointmentRepo: apptRepo,
		patientRepo:     patientRepo,
		authService:     &AuthService{sessionManager: auth.NewSessionManager(8)},
		auditService:    newTestAuditService(),
	}

	input := CreateAppointmentInput{
		PatientID: "patient-1",
		Date:      "2026-05-16",
		StartTime: "10:15",
		EndTime:   "10:45",
	}

	_, err := svc.CreateAppointment(input)
	if err == nil {
		t.Fatal("expected conflict error")
	}

	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T (%v)", err, err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestAppointmentService_CreateAppointment_MissingFields(t *testing.T) {
	svc := &AppointmentService{
		appointmentRepo: newMockAppointmentRepoForService(),
		patientRepo:     newMockPatientRepoForAppointment(),
		authService:     &AuthService{sessionManager: auth.NewSessionManager(8)},
		auditService:    newTestAuditService(),
	}

	tests := []struct {
		name  string
		input CreateAppointmentInput
	}{
		{"missing patient", CreateAppointmentInput{Date: "2026-05-16", StartTime: "10:00", EndTime: "10:30"}},
		{"missing date", CreateAppointmentInput{PatientID: "p1", StartTime: "10:00", EndTime: "10:30"}},
		{"missing start time", CreateAppointmentInput{PatientID: "p1", Date: "2026-05-16", EndTime: "10:30"}},
		{"missing end time", CreateAppointmentInput{PatientID: "p1", Date: "2026-05-16", StartTime: "10:00"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateAppointment(tt.input)
			if err == nil {
				t.Fatal("expected validation error")
			}
			var appErr *utils.AppError
			if !errors.As(err, &appErr) {
				t.Fatalf("expected AppError, got: %T", err)
			}
			if appErr.Code != "VALIDATION_ERROR" {
				t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
			}
		})
	}
}
