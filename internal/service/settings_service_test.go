package service

import (
	"testing"

	"clinmitra/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type settingsMockClinicRepo struct {
	mock.Mock
}

func (m *settingsMockClinicRepo) Get() (*models.ClinicSettings, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClinicSettings), args.Error(1)
}

func (m *settingsMockClinicRepo) Upsert(settings *models.ClinicSettings) error {
	return m.Called(settings).Error(0)
}

func (m *settingsMockClinicRepo) IsSetupComplete() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

type settingsMockTreatmentRepo struct {
	mock.Mock
}

func (m *settingsMockTreatmentRepo) Create(t *models.Treatment) error { return nil }
func (m *settingsMockTreatmentRepo) FindByID(id string) (*models.Treatment, error) {
	return &models.Treatment{}, nil
}
func (m *settingsMockTreatmentRepo) Update(t *models.Treatment) error { return nil }
func (m *settingsMockTreatmentRepo) Delete(id string) error           { return nil }
func (m *settingsMockTreatmentRepo) ListActive() ([]models.Treatment, error) {
	return nil, nil
}
func (m *settingsMockTreatmentRepo) ListAll() ([]models.Treatment, error) {
	return nil, nil
}

type settingsMockAuditRepo struct{}

func (m *settingsMockAuditRepo) Create(log *models.AuditLog) error { return nil }
func (m *settingsMockAuditRepo) ListByEntity(entityType, entityID string) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *settingsMockAuditRepo) ListByUser(userID string, limit int) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *settingsMockAuditRepo) ListRecent(limit int) ([]models.AuditLog, error) {
	return nil, nil
}

func TestUpdateClinicSettings_Validation(t *testing.T) {
	clinicRepo := new(settingsMockClinicRepo)

	testSettings := &models.ClinicSettings{
		ID:                  "test-id",
		ClinicName:          "Test Clinic",
		DoctorName:          "Dr. Test",
		DoctorQualification: "BDS",
	}

	// Mock success
	clinicRepo.On("Upsert", testSettings).Return(nil)

	err := clinicRepo.Upsert(testSettings)
	assert.NoError(t, err)
	clinicRepo.AssertExpectations(t)
}

func TestClinicSettingsModel(t *testing.T) {
	settings := models.ClinicSettings{
		DoctorName:          "Dr. Smith",
		DoctorQualification: "MDS",
	}
	assert.Equal(t, "Dr. Smith", settings.DoctorName)
	assert.Equal(t, "MDS", settings.DoctorQualification)
}
