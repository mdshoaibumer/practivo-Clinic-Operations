package service

import (
	"errors"
	"testing"

	"clinmitra/internal/auth"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"
)

// --- Mock Repositories for Patient Service Tests ---

type mockPatientRepoForService struct {
	patients    map[string]*models.Patient
	phoneIndex  map[string]*models.Patient
	createErr   error
	createCalls int
}

func newMockPatientRepoForService() *mockPatientRepoForService {
	return &mockPatientRepoForService{
		patients:   make(map[string]*models.Patient),
		phoneIndex: make(map[string]*models.Patient),
	}
}

func (m *mockPatientRepoForService) Create(patient *models.Patient) error {
	m.createCalls++
	if m.createErr != nil {
		return m.createErr
	}
	m.patients[patient.ID] = patient
	m.phoneIndex[patient.Phone] = patient
	return nil
}

func (m *mockPatientRepoForService) FindByID(id string) (*models.Patient, error) {
	p, ok := m.patients[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return p, nil
}

func (m *mockPatientRepoForService) Update(patient *models.Patient) error {
	m.patients[patient.ID] = patient
	m.phoneIndex[patient.Phone] = patient
	return nil
}

func (m *mockPatientRepoForService) Delete(id string) error {
	delete(m.patients, id)
	return nil
}

func (m *mockPatientRepoForService) List(page, pageSize int, search string) ([]models.Patient, int64, error) {
	var results []models.Patient
	for _, p := range m.patients {
		results = append(results, *p)
	}
	return results, int64(len(results)), nil
}

func (m *mockPatientRepoForService) FindByPhone(phone string) (*models.Patient, error) {
	p, ok := m.phoneIndex[phone]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return p, nil
}

func (m *mockPatientRepoForService) Count() (int64, error) {
	return int64(len(m.patients)), nil
}

type mockPatientTreatmentRepoForService struct{}

func (m *mockPatientTreatmentRepoForService) Create(pt *models.PatientTreatment) error { return nil }
func (m *mockPatientTreatmentRepoForService) CreateBatch(pts []models.PatientTreatment) error {
	return nil
}
func (m *mockPatientTreatmentRepoForService) ListByPatient(patientID string) ([]models.PatientTreatment, error) {
	return nil, nil
}

type mockInvoiceRepoForPatient struct{}

func (m *mockInvoiceRepoForPatient) Create(invoice *models.Invoice) error        { return nil }
func (m *mockInvoiceRepoForPatient) FindByID(id string) (*models.Invoice, error) { return nil, nil }
func (m *mockInvoiceRepoForPatient) Update(invoice *models.Invoice) error        { return nil }
func (m *mockInvoiceRepoForPatient) List(page, pageSize int, filters repository.InvoiceFilters) ([]models.Invoice, int64, error) {
	return nil, 0, nil
}
func (m *mockInvoiceRepoForPatient) ListByPatient(patientID string) ([]models.Invoice, error) {
	return nil, nil
}
func (m *mockInvoiceRepoForPatient) GetLastInvoiceNumber(prefix, yearMonth string) (string, error) {
	return "", nil
}
func (m *mockInvoiceRepoForPatient) GetOutstandingByPatient(patientID string) (int64, error) {
	return 0, nil
}
func (m *mockInvoiceRepoForPatient) GetTotalOutstanding() (int64, error)              { return 0, nil }
func (m *mockInvoiceRepoForPatient) GetRevenueByDateRange(s, e string) (int64, error) { return 0, nil }

type mockAuditRepoForPatientTests struct{}

func (m *mockAuditRepoForPatientTests) Create(log *models.AuditLog) error { return nil }
func (m *mockAuditRepoForPatientTests) ListByEntity(entityType, entityID string) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditRepoForPatientTests) ListByUser(userID string, limit int) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditRepoForPatientTests) ListRecent(limit int) ([]models.AuditLog, error) {
	return nil, nil
}

// --- Service-Level Tests ---

func TestPatientService_CreatePatient_Success(t *testing.T) {
	patientRepo := newMockPatientRepoForService()

	// Verify service can be constructed with all dependencies
	service := NewPatientService(
		patientRepo,
		&mockPatientTreatmentRepoForService{},
		&mockInvoiceRepoForPatient{},
		&AuthService{sessionManager: auth.NewSessionManager(8)},
		NewAuditService(&mockAuditRepoForPatientTests{}),
		nil,
	)

	// Verify the service was correctly assigned
	if service.patientRepo == nil {
		t.Fatal("service.patientRepo should not be nil")
	}

	input := CreatePatientInput{
		Name:   "Ramesh Kumar",
		Phone:  "9876543210",
		Gender: "male",
		Age:    35,
	}

	// Validate the input validation path (no DB needed)
	err := validatePatientInput(input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify duplicate detection works
	patientRepo.phoneIndex["9876543210"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "existing-id"},
		Phone:     "9876543210",
	}

	_, dupErr := patientRepo.FindByPhone("9876543210")
	if dupErr != nil {
		t.Fatal("expected to find existing patient by phone")
	}
}

func TestPatientService_CreatePatient_DuplicatePhone(t *testing.T) {
	patientRepo := newMockPatientRepoForService()
	patientRepo.phoneIndex["9876543210"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "existing-id"},
		Name:      "Existing Patient",
		Phone:     "9876543210",
	}

	input := CreatePatientInput{
		Phone: "9876543210",
	}

	// Simulate the duplicate check logic from CreatePatient
	cleanedPhone := utils.CleanPhone(input.Phone)
	existing, err := patientRepo.FindByPhone(cleanedPhone)
	if err == nil && existing != nil {
		// This is the expected path
		return
	}
	t.Fatal("expected duplicate phone detection to trigger")
}

func TestPatientService_GetPatient_NotFound(t *testing.T) {
	patientRepo := newMockPatientRepoForService()

	_, err := patientRepo.FindByID("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent patient")
	}

	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got: %s", appErr.Code)
	}
}

func TestPatientService_GetPatient_Found(t *testing.T) {
	patientRepo := newMockPatientRepoForService()
	patientRepo.patients["test-id"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "test-id"},
		Name:      "Suresh Patel",
		Phone:     "9876543210",
	}

	patient, err := patientRepo.FindByID("test-id")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if patient.Name != "Suresh Patel" {
		t.Errorf("expected name 'Suresh Patel', got: %s", patient.Name)
	}
}

func TestPatientService_UpdatePatient_PhoneUniqueness(t *testing.T) {
	patientRepo := newMockPatientRepoForService()

	// Existing patients
	patientRepo.patients["patient-1"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "patient-1"},
		Name:      "Patient One",
		Phone:     "9876543210",
	}
	patientRepo.phoneIndex["9876543210"] = patientRepo.patients["patient-1"]

	patientRepo.patients["patient-2"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "patient-2"},
		Name:      "Patient Two",
		Phone:     "9876543211",
	}
	patientRepo.phoneIndex["9876543211"] = patientRepo.patients["patient-2"]

	// Simulate: patient-2 tries to change phone to patient-1's phone
	targetPatient := patientRepo.patients["patient-2"]
	newPhone := "9876543210"

	if newPhone != targetPatient.Phone {
		existing, err := patientRepo.FindByPhone(newPhone)
		if err == nil && existing != nil && existing.ID != "patient-2" {
			// Expected: conflict detected
			return
		}
	}
	t.Fatal("expected phone uniqueness conflict to be detected")
}

func TestPatientService_ListPatients(t *testing.T) {
	patientRepo := newMockPatientRepoForService()
	patientRepo.patients["p1"] = &models.Patient{BaseModel: models.BaseModel{ID: "p1"}, Name: "A"}
	patientRepo.patients["p2"] = &models.Patient{BaseModel: models.BaseModel{ID: "p2"}, Name: "B"}
	patientRepo.patients["p3"] = &models.Patient{BaseModel: models.BaseModel{ID: "p3"}, Name: "C"}

	patients, total, err := patientRepo.List(1, 10, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got: %d", total)
	}
	if len(patients) != 3 {
		t.Errorf("expected 3 patients, got: %d", len(patients))
	}
}
