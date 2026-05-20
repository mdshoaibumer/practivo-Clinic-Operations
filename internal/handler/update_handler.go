package handler

import (
	"fmt"
	"log/slog"

	"clinmitra/internal/service"
)

// UpdateHandler exposes update operations to the Wails frontend.
type UpdateHandler struct {
	updateService *service.UpdateService
}

// NewUpdateHandler creates a new UpdateHandler.
func NewUpdateHandler(updateService *service.UpdateService) *UpdateHandler {
	return &UpdateHandler{updateService: updateService}
}

// CheckForUpdate checks if a newer version is available on GitHub.
func (h *UpdateHandler) CheckForUpdate() (*service.UpdateInfo, error) {
	info, err := h.updateService.CheckForUpdate()
	if err != nil {
		slog.Error("update check failed", "error", err)
		return nil, safeError(err)
	}
	return info, nil
}

// DownloadAndInstallUpdate downloads and launches the installer.
func (h *UpdateHandler) DownloadAndInstallUpdate(downloadURL string) error {
	if downloadURL == "" {
		return safeError(fmt.Errorf("no download URL provided"))
	}

	if err := h.updateService.DownloadAndInstallUpdate(downloadURL); err != nil {
		slog.Error("update install failed", "error", err)
		return safeError(err)
	}

	return nil
}
