package handler

import (
	"log/slog"

	"clinmitra/internal/service"
)

type BackupHandler struct {
	backupService *service.BackupService
}

// NewBackupHandler creates a BackupHandler backed by the given service.
func NewBackupHandler(backupService *service.BackupService) *BackupHandler {
	return &BackupHandler{backupService: backupService}
}

// CreateBackup triggers a database backup to the specified directory
// (or the default backup directory if empty).
func (h *BackupHandler) CreateBackup(destinationDir string) (*service.BackupInfo, error) {
	slog.Info("creating backup", "destination", destinationDir)
	result, err := h.backupService.CreateBackup(destinationDir)
	if err != nil {
		slog.Error("backup failed", "destination", destinationDir, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("backup completed", "filePath", result.FilePath, "size", result.Size)
	return result, nil
}

// RestoreFromBackup replaces the current database with a verified backup file.
func (h *BackupHandler) RestoreFromBackup(backupPath string) error {
	slog.Info("restoring from backup", "path", backupPath)
	err := h.backupService.RestoreFromBackup(backupPath)
	if err != nil {
		slog.Error("restore failed", "path", backupPath, "error", err.Error())
		return safeError(err)
	}
	slog.Info("restore completed", "path", backupPath)
	return nil
}

// VerifyBackup checks a backup file's integrity via SQLite PRAGMA integrity_check.
func (h *BackupHandler) VerifyBackup(filePath string) (bool, error) {
	slog.Info("verifying backup", "path", filePath)
	result, err := h.backupService.VerifyBackup(filePath)
	if err != nil {
		slog.Warn("backup verification failed", "path", filePath, "error", err.Error())
		return false, safeError(err)
	}
	slog.Info("backup verified", "path", filePath, "valid", result)
	return result, nil
}

// ListBackups returns all available backup files sorted by date (newest first).
func (h *BackupHandler) ListBackups() ([]service.BackupInfo, error) {
	result, err := h.backupService.ListBackups()
	if err != nil {
		slog.Error("list backups failed", "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// GetAutoBackupPath returns the configured automatic backup directory path.
func (h *BackupHandler) GetAutoBackupPath() string {
	return h.backupService.GetAutoBackupPath()
}

// DetectCloudDrives scans the system for Google Drive, OneDrive, and Dropbox
// sync folders and returns available options for cloud backup configuration.
func (h *BackupHandler) DetectCloudDrives() []service.CloudDriveInfo {
	return h.backupService.DetectCloudDrives()
}

// CreateCloudBackup triggers a backup to the configured cloud sync folder.
func (h *BackupHandler) CreateCloudBackup() (*service.BackupInfo, error) {
	slog.Info("creating cloud backup")
	result, err := h.backupService.CreateCloudBackup()
	if err != nil {
		slog.Error("cloud backup failed", "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("cloud backup completed", "filePath", result.FilePath)
	return result, nil
}
