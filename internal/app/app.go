package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"clinmitra/internal/auth"
	"clinmitra/internal/config"
	"clinmitra/internal/db"
	"clinmitra/internal/handler"
	"clinmitra/internal/repository"
	"clinmitra/internal/service"

	"gorm.io/gorm"
)

type Application struct {
	ctx context.Context
	cfg *config.Config
	db  *gorm.DB

	// Handlers (Wails-bound)
	AuthHandler        *handler.AuthHandler
	SettingsHandler    *handler.SettingsHandler
	PatientHandler     *handler.PatientHandler
	AppointmentHandler *handler.AppointmentHandler
	InvoiceHandler     *handler.InvoiceHandler
	DashboardHandler   *handler.DashboardHandler
	BackupHandler      *handler.BackupHandler
	UpdateHandler      *handler.UpdateHandler

	// Services (for shutdown hooks)
	backupService *service.BackupService
	logFile       *os.File // retained for graceful close on shutdown
}

// NewApplication creates and wires together all application components:
// config, logger, database, repositories, services, and handlers.
// Returns a fully initialized Application ready for Wails binding.
func NewApplication() (*Application, error) {
	// Initialize config
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize structured logging with rotation
	logPath := filepath.Join(cfg.LogDir, "clinmitra.log")
	if err := rotateLogIfNeeded(logPath, 10*1024*1024); err != nil { // 10MB max
		// Non-fatal: proceed without rotation
		slog.Warn("log rotation failed", "error", err)
	}

	logFile, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0600,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed default data
	if err := db.SeedTreatments(database); err != nil {
		return nil, fmt.Errorf("failed to seed treatments: %w", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	clinicRepo := repository.NewClinicRepository(database)
	patientRepo := repository.NewPatientRepository(database)
	treatmentRepo := repository.NewTreatmentRepository(database)
	appointmentRepo := repository.NewAppointmentRepository(database)
	invoiceRepo := repository.NewInvoiceRepository(database)
	invoiceItemRepo := repository.NewInvoiceItemRepository(database)
	paymentRepo := repository.NewPaymentRepository(database)
	patientTreatmentRepo := repository.NewPatientTreatmentRepository(database)
	auditRepo := repository.NewAuditRepository(database)

	// Initialize auth components
	sessionManager := auth.NewSessionManager(cfg.SessionHours)
	loginLimiter := auth.NewLoginLimiter(cfg.MaxLoginAttempts, cfg.LockoutMinutes)

	// Initialize services
	auditService := service.NewAuditService(auditRepo)
	authService := service.NewAuthService(userRepo, sessionManager, loginLimiter, auditService, cfg)
	settingsService := service.NewSettingsService(clinicRepo, treatmentRepo, authService, auditService, cfg)
	patientService := service.NewPatientService(patientRepo, patientTreatmentRepo, invoiceRepo, authService, auditService, database)
	appointmentService := service.NewAppointmentService(appointmentRepo, patientRepo, authService, auditService, database)
	invoiceService := service.NewInvoiceService(invoiceRepo, invoiceItemRepo, paymentRepo, patientRepo, treatmentRepo, clinicRepo, patientTreatmentRepo, authService, auditService, database)
	dashboardService := service.NewDashboardService(invoiceRepo, paymentRepo, appointmentRepo, patientRepo)
	backupService := service.NewBackupService(database, cfg, authService, auditService, clinicRepo)
	updateService := service.NewUpdateService(cfg)

	// Initialize handlers
	app := &Application{
		cfg:                cfg,
		db:                 database,
		logFile:            logFile,
		AuthHandler:        handler.NewAuthHandler(authService),
		SettingsHandler:    handler.NewSettingsHandler(settingsService),
		PatientHandler:     handler.NewPatientHandler(patientService),
		AppointmentHandler: handler.NewAppointmentHandler(appointmentService),
		InvoiceHandler:     handler.NewInvoiceHandler(invoiceService),
		DashboardHandler:   handler.NewDashboardHandler(dashboardService),
		BackupHandler:      handler.NewBackupHandler(backupService),
		UpdateHandler:      handler.NewUpdateHandler(updateService),
		backupService:      backupService,
	}

	return app, nil
}

// Startup is called by Wails when the application window is ready.
// It stores the context for runtime access and logs the startup event.
func (a *Application) Startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("application started", "version", a.cfg.Version, "app", a.cfg.AppName)
}

// Shutdown is called by Wails when the application is closing.
// It performs a best-effort auto-backup and gracefully closes the database
// (including a WAL checkpoint to flush pending writes).
func (a *Application) Shutdown(ctx context.Context) {
	slog.Info("application shutting down")

	// Auto-backup on shutdown (best effort — includes cloud sync if configured)
	if _, err := a.backupService.CreateBackupWithCloudSync(); err != nil {
		slog.Error("auto-backup on shutdown failed", "error", err)
	}

	// Close database (includes WAL checkpoint)
	if err := db.CloseDatabase(a.db); err != nil {
		slog.Error("error closing database", "error", err)
	}

	// Close log file
	if a.logFile != nil {
		a.logFile.Close()
	}
}

// GetBindings returns the list of handler structs to be bound to the
// Wails frontend via JavaScript/TypeScript bindings.
func (a *Application) GetBindings() []interface{} {
	return []interface{}{
		a.AuthHandler,
		a.SettingsHandler,
		a.PatientHandler,
		a.AppointmentHandler,
		a.InvoiceHandler,
		a.DashboardHandler,
		a.BackupHandler,
		a.UpdateHandler,
	}
}

// rotateLogIfNeeded rotates the log file if it exceeds maxSize bytes.
// Keeps one previous log file as .1 backup.
// Time complexity: O(1) — single stat + rename
// Space complexity: O(1)
func rotateLogIfNeeded(logPath string, maxSize int64) error {
	info, err := os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No log file yet
		}
		return err
	}

	if info.Size() < maxSize {
		return nil
	}

	// Rotate: rename current to .1, removing old .1 if exists
	rotatedPath := logPath + ".1"
	os.Remove(rotatedPath)
	return os.Rename(logPath, rotatedPath)
}
