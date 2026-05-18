package handler

import (
	"fmt"
	"log/slog"

	"clinmitra/internal/models"
	"clinmitra/internal/service"
)

// MaxLogoSize is the maximum allowed size for a logo in bytes (512KB).
const MaxLogoSize = 512 * 1024

type SettingsHandler struct {
	settingsService *service.SettingsService
}

// NewSettingsHandler creates a SettingsHandler backed by the given service.
func NewSettingsHandler(settingsService *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

// IsSetupComplete returns whether the initial clinic setup wizard has been completed.
func (h *SettingsHandler) IsSetupComplete() (bool, error) {
	result, err := h.settingsService.IsSetupComplete()
	return result, safeError(err)
}

// CompleteSetup runs the first-time setup wizard: creates the admin user
// and saves clinic settings. Can only be called once.
func (h *SettingsHandler) CompleteSetup(input service.SetupInput) error {
	slog.Info("completing initial setup")
	err := h.settingsService.CompleteSetup(input)
	if err != nil {
		slog.Warn("setup failed", "error", err.Error())
		return safeError(err)
	}
	slog.Info("initial setup completed successfully")
	return nil
}

// GetClinicSettings returns the current clinic configuration.
func (h *SettingsHandler) GetClinicSettings() (*models.ClinicSettings, error) {
	result, err := h.settingsService.GetClinicSettings()
	return result, safeError(err)
}

// UpdateClinicSettings persists updated clinic settings.
func (h *SettingsHandler) UpdateClinicSettings(settings *models.ClinicSettings) error {
	return safeError(h.settingsService.UpdateClinicSettings(settings))
}

// ListTreatments returns all active treatments available for invoicing.
func (h *SettingsHandler) ListTreatments() ([]models.Treatment, error) {
	result, err := h.settingsService.ListTreatments()
	return result, safeError(err)
}

// ListAllTreatments returns all treatments including inactive ones.
func (h *SettingsHandler) ListAllTreatments() ([]models.Treatment, error) {
	result, err := h.settingsService.ListAllTreatments()
	return result, safeError(err)
}

// CreateTreatment adds a new dental treatment/procedure to the system.
func (h *SettingsHandler) CreateTreatment(name, code, category, description string, defaultPrice int64) (*models.Treatment, error) {
	slog.Info("creating treatment", "name", name, "code", code, "category", category)
	result, err := h.settingsService.CreateTreatment(name, code, category, description, defaultPrice)
	if err != nil {
		slog.Warn("create treatment failed", "name", name, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("treatment created", "id", result.ID, "name", name)
	return result, nil
}

// UpdateTreatment modifies an existing treatment's details.
func (h *SettingsHandler) UpdateTreatment(id, name, code, category, description string, defaultPrice int64) error {
	slog.Info("updating treatment", "id", id, "name", name)
	err := h.settingsService.UpdateTreatment(id, name, code, category, description, defaultPrice)
	if err != nil {
		slog.Warn("update treatment failed", "id", id, "error", err.Error())
		return safeError(err)
	}
	slog.Info("treatment updated", "id", id)
	return nil
}

// DeleteTreatment soft-deletes a treatment (marks as inactive).
func (h *SettingsHandler) DeleteTreatment(id string) error {
	slog.Info("deleting treatment", "id", id)
	err := h.settingsService.DeleteTreatment(id)
	if err != nil {
		slog.Warn("delete treatment failed", "id", id, "error", err.Error())
		return safeError(err)
	}
	slog.Info("treatment deleted", "id", id)
	return nil
}

// UploadLogo saves a base64-encoded logo image to clinic settings.
// Accepts the base64 data string (with or without data URI prefix).
// Maximum size: 512KB of raw data.
func (h *SettingsHandler) UploadLogo(base64Data string) error {
	if len(base64Data) == 0 {
		return safeError(fmt.Errorf("no logo data provided"))
	}
	// Base64 is ~4/3 of original size; 512KB raw ≈ 700KB base64
	if len(base64Data) > MaxLogoSize*2 {
		return safeError(fmt.Errorf("logo file too large (max 512KB)"))
	}
	return safeError(h.settingsService.SaveLogo(base64Data))
}

// RemoveLogo clears the clinic logo from settings.
func (h *SettingsHandler) RemoveLogo() error {
	return safeError(h.settingsService.RemoveLogo())
}
