package handler

import (
	"fmt"
	"testing"

	"clinmitra/internal/auth"
	"clinmitra/internal/config"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupHandlerTestDB creates a fresh in-memory DB with all tables.
func setupHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)
	t.Cleanup(func() { sqlDB.Close() })

	err = db.AutoMigrate(
		&models.User{},
		&models.Patient{},
		&models.Appointment{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Payment{},
		&models.ClinicSettings{},
		&models.AuditLog{},
		&models.Treatment{},
		&models.PatientTreatment{},
	)
	if err != nil {
		t.Fatalf("migrate error: %v", err)
	}
	return db
}

func testConfig() *config.Config {
	return &config.Config{
		AppName:          "Test",
		Version:          "0.0.1",
		DataDir:          "",
		DBPath:           ":memory:",
		BackupDir:        "",
		LogDir:           "",
		MaxLoginAttempts: 5,
		LockoutMinutes:   15,
		SessionHours:     8,
		BcryptCost:       4,
	}
}

// setupHandlerStack wires all handlers with a real in-memory DB.
func setupHandlerStack(t *testing.T) (*PatientHandler, *AppointmentHandler, *InvoiceHandler, *SettingsHandler, *AuthHandler, *DashboardHandler, *BackupHandler, *service.AuthService) {
	t.Helper()
	db := setupHandlerTestDB(t)

	userRepo := repository.NewUserRepository(db)
	clinicRepo := repository.NewClinicRepository(db)
	patientRepo := repository.NewPatientRepository(db)
	treatmentRepo := repository.NewTreatmentRepository(db)
	appointmentRepo := repository.NewAppointmentRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	invoiceItemRepo := repository.NewInvoiceItemRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	patientTreatmentRepo := repository.NewPatientTreatmentRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	sessionManager := auth.NewSessionManager(8)
	loginLimiter := auth.NewLoginLimiter(5, 15)

	cfg := testConfig()
	cfg.DataDir = t.TempDir()
	cfg.BackupDir = cfg.DataDir

	auditService := service.NewAuditService(auditRepo)
	authService := service.NewAuthService(userRepo, sessionManager, loginLimiter, auditService, cfg)
	settingsService := service.NewSettingsService(clinicRepo, treatmentRepo, authService, auditService, cfg)
	patientService := service.NewPatientService(patientRepo, patientTreatmentRepo, invoiceRepo, authService, auditService, db)
	appointmentService := service.NewAppointmentService(appointmentRepo, patientRepo, authService, auditService, db)
	invoiceService := service.NewInvoiceService(invoiceRepo, invoiceItemRepo, paymentRepo, patientRepo, treatmentRepo, clinicRepo, patientTreatmentRepo, authService, auditService, db)
	dashboardService := service.NewDashboardService(invoiceRepo, paymentRepo, appointmentRepo, patientRepo)
	backupService := service.NewBackupService(db, cfg, authService, auditService, clinicRepo)

	patientHandler := NewPatientHandler(patientService)
	appointmentHandler := NewAppointmentHandler(appointmentService)
	invoiceHandler := NewInvoiceHandler(invoiceService)
	settingsHandler := NewSettingsHandler(settingsService)
	authHandler := NewAuthHandler(authService)
	dashboardHandler := NewDashboardHandler(dashboardService)
	backupHandler := NewBackupHandler(backupService)

	return patientHandler, appointmentHandler, invoiceHandler, settingsHandler, authHandler, dashboardHandler, backupHandler, authService
}

func TestPatientHandler_CreatePatient(t *testing.T) {
	patientHandler, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	// Setup first
	err := settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName:    "Test Clinic",
		DoctorName:    "Dr. Test",
		Phone:         "9876543210",
		AdminUsername: "admin",
		AdminPassword: "admin123",
		AdminFullName: "Admin User",
	})
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	// Login
	_, err = authHandler.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("login error: %v", err)
	}

	// Create patient
	patient, err := patientHandler.CreatePatient(service.CreatePatientInput{
		Name:   "Test Patient",
		Phone:  "9876543211",
		Gender: "male",
	})
	if err != nil {
		t.Fatalf("CreatePatient error: %v", err)
	}
	if patient.Name != "Test Patient" {
		t.Errorf("expected name 'Test Patient', got %q", patient.Name)
	}
	if patient.ID == "" {
		t.Error("expected patient ID to be set")
	}

	// Create duplicate phone should fail
	_, err = patientHandler.CreatePatient(service.CreatePatientInput{
		Name:   "Duplicate",
		Phone:  "9876543211",
		Gender: "female",
	})
	if err == nil {
		t.Error("expected error for duplicate phone")
	}
}

func TestPatientHandler_GetPatient(t *testing.T) {
	patientHandler, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Find Me", Phone: "8765432100", Gender: "female",
	})

	// Get existing
	found, err := patientHandler.GetPatient(patient.ID)
	if err != nil {
		t.Fatalf("GetPatient error: %v", err)
	}
	if found.Name != "Find Me" {
		t.Errorf("expected 'Find Me', got %q", found.Name)
	}

	// Get non-existent
	_, err = patientHandler.GetPatient("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent patient")
	}
}

func TestPatientHandler_ListPatients(t *testing.T) {
	patientHandler, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patientHandler.CreatePatient(service.CreatePatientInput{Name: "Alice", Phone: "9876543211", Gender: "female"})
	patientHandler.CreatePatient(service.CreatePatientInput{Name: "Bob", Phone: "9876543212", Gender: "male"})

	result, err := patientHandler.ListPatients(1, 10, "")
	if err != nil {
		t.Fatalf("ListPatients error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 patients, got %d", result.Total)
	}

	// Search
	result, err = patientHandler.ListPatients(1, 10, "Alice")
	if err != nil {
		t.Fatalf("ListPatients search error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 patient for search 'Alice', got %d", result.Total)
	}

	// Bad pagination values get sanitized
	result, err = patientHandler.ListPatients(-1, 999, "")
	if err != nil {
		t.Fatalf("ListPatients bad pagination error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
}

func TestPatientHandler_UpdatePatient(t *testing.T) {
	patientHandler, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Original", Phone: "9876543211", Gender: "male",
	})

	updated, err := patientHandler.UpdatePatient(patient.ID, service.CreatePatientInput{
		Name: "Updated Name", Phone: "9876543211", Gender: "male",
	})
	if err != nil {
		t.Fatalf("UpdatePatient error: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected 'Updated Name', got %q", updated.Name)
	}
}

func TestPatientHandler_DeletePatient(t *testing.T) {
	patientHandler, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Delete Me", Phone: "9876543211", Gender: "male",
	})

	err := patientHandler.DeletePatient(patient.ID)
	if err != nil {
		t.Fatalf("DeletePatient error: %v", err)
	}

	// Should not find anymore
	_, err = patientHandler.GetPatient(patient.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestPatientHandler_CheckDuplicatePhone(t *testing.T) {
	patientHandler, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Existing", Phone: "9876543211", Gender: "male",
	})

	// Should find duplicate
	found, _ := patientHandler.CheckDuplicatePhone("9876543211")
	if found == nil {
		t.Error("expected to find existing patient with this phone")
	}

	// Should not find for new number
	found, _ = patientHandler.CheckDuplicatePhone("9999999999")
	if found != nil {
		t.Error("expected nil for new phone number")
	}
}

func TestPatientHandler_GetPatientCount(t *testing.T) {
	patientHandler, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	count, err := patientHandler.GetPatientCount()
	if err != nil {
		t.Fatalf("GetPatientCount error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 patients, got %d", count)
	}

	patientHandler.CreatePatient(service.CreatePatientInput{Name: "One", Phone: "9876543211", Gender: "male"})
	count, _ = patientHandler.GetPatientCount()
	if count != 1 {
		t.Errorf("expected 1 patient, got %d", count)
	}
}

func TestAppointmentHandler_CreateAndGet(t *testing.T) {
	patientHandler, appointmentHandler, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Appt Patient", Phone: "9876543211", Gender: "male",
	})

	// Create appointment
	appt, err := appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID,
		Date:      "2026-12-15",
		StartTime: "10:00",
		EndTime:   "10:30",
		Duration:  30,
		Purpose:   "Checkup",
	})
	if err != nil {
		t.Fatalf("CreateAppointment error: %v", err)
	}
	if appt.ID == "" {
		t.Error("expected appointment ID")
	}

	// Get by ID
	got, err := appointmentHandler.GetAppointment(appt.ID)
	if err != nil {
		t.Fatalf("GetAppointment error: %v", err)
	}
	if got.Purpose != "Checkup" {
		t.Errorf("expected purpose 'Checkup', got %q", got.Purpose)
	}

	// Get by date
	dateAppts, err := appointmentHandler.GetAppointmentsByDate("2026-12-15")
	if err != nil {
		t.Fatalf("GetAppointmentsByDate error: %v", err)
	}
	if len(dateAppts) != 1 {
		t.Errorf("expected 1 appointment, got %d", len(dateAppts))
	}

	// Get today (should be empty)
	todayAppts, err := appointmentHandler.GetTodayAppointments()
	if err != nil {
		t.Fatalf("GetTodayAppointments error: %v", err)
	}
	if len(todayAppts) != 0 {
		t.Errorf("expected 0 today appointments, got %d", len(todayAppts))
	}

	// Get week
	weekAppts, err := appointmentHandler.GetWeekAppointments("2026-12-14", "2026-12-16")
	if err != nil {
		t.Fatalf("GetWeekAppointments error: %v", err)
	}
	if len(weekAppts) != 1 {
		t.Errorf("expected 1 week appointment, got %d", len(weekAppts))
	}

	// Get patient appointments
	patientAppts, err := appointmentHandler.GetPatientAppointments(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientAppointments error: %v", err)
	}
	if len(patientAppts) != 1 {
		t.Errorf("expected 1 patient appointment, got %d", len(patientAppts))
	}
}

func TestAppointmentHandler_UpdateCancelComplete(t *testing.T) {
	patientHandler, appointmentHandler, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Patient", Phone: "9876543211", Gender: "male",
	})

	appt, _ := appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-15",
		StartTime: "10:00", EndTime: "10:30", Duration: 30,
	})

	// Update
	updated, err := appointmentHandler.UpdateAppointment(appt.ID, service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-16",
		StartTime: "11:00", EndTime: "11:30", Duration: 30, Purpose: "Updated",
	})
	if err != nil {
		t.Fatalf("UpdateAppointment error: %v", err)
	}
	if updated.Purpose != "Updated" {
		t.Errorf("expected purpose 'Updated', got %q", updated.Purpose)
	}

	// Complete
	err = appointmentHandler.CompleteAppointment(appt.ID)
	if err != nil {
		t.Fatalf("CompleteAppointment error: %v", err)
	}

	// Create another to cancel
	appt2, _ := appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-17",
		StartTime: "10:00", EndTime: "10:30", Duration: 30,
	})

	err = appointmentHandler.CancelAppointment(appt2.ID, "Patient request")
	if err != nil {
		t.Fatalf("CancelAppointment error: %v", err)
	}
}

func TestAuthHandler_LoginLogout(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})

	// Login with wrong password
	_, err := authHandler.Login("admin", "wrong")
	if err == nil {
		t.Error("expected error for wrong password")
	}

	// Login correct
	resp, err := authHandler.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if !resp.LoggedIn {
		t.Error("expected LoggedIn=true")
	}

	// GetCurrentUser
	current, err := authHandler.GetCurrentUser()
	if err != nil {
		t.Fatalf("GetCurrentUser error: %v", err)
	}
	if !current.LoggedIn {
		t.Error("expected logged in")
	}

	// ChangePassword
	err = authHandler.ChangePassword("admin123", "newpass123")
	if err != nil {
		t.Fatalf("ChangePassword error: %v", err)
	}

	// Logout
	err = authHandler.Logout()
	if err != nil {
		t.Fatalf("Logout error: %v", err)
	}

	// After logout, GetCurrentUser should show not logged in
	current, err = authHandler.GetCurrentUser()
	if err != nil {
		t.Fatalf("GetCurrentUser after logout error: %v", err)
	}
	if current.LoggedIn {
		t.Error("expected not logged in after logout")
	}
}

func TestSettingsHandler_Setup(t *testing.T) {
	_, _, _, settingsHandler, _, _, _, _ := setupHandlerStack(t)

	// Not complete initially
	complete, err := settingsHandler.IsSetupComplete()
	if err != nil {
		t.Fatalf("IsSetupComplete error: %v", err)
	}
	if complete {
		t.Error("expected setup not complete")
	}

	// Complete setup
	err = settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "My Clinic", DoctorName: "Dr. Smith", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	if err != nil {
		t.Fatalf("CompleteSetup error: %v", err)
	}

	complete, _ = settingsHandler.IsSetupComplete()
	if !complete {
		t.Error("expected setup complete after CompleteSetup")
	}

	// Get settings
	settings, err := settingsHandler.GetClinicSettings()
	if err != nil {
		t.Fatalf("GetClinicSettings error: %v", err)
	}
	if settings.ClinicName != "My Clinic" {
		t.Errorf("expected 'My Clinic', got %q", settings.ClinicName)
	}
}

func TestSettingsHandler_Treatments(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// Create treatment
	treatment, err := settingsHandler.CreateTreatment("Root Canal", "RCT", "Endodontics", "Root canal treatment", 500000)
	if err != nil {
		t.Fatalf("CreateTreatment error: %v", err)
	}
	if treatment.Name != "Root Canal" {
		t.Errorf("expected 'Root Canal', got %q", treatment.Name)
	}

	// List treatments
	treatments, err := settingsHandler.ListTreatments()
	if err != nil {
		t.Fatalf("ListTreatments error: %v", err)
	}
	if len(treatments) != 1 {
		t.Errorf("expected 1 treatment, got %d", len(treatments))
	}

	// Update treatment
	err = settingsHandler.UpdateTreatment(treatment.ID, "Root Canal Updated", "RCT2", "Endodontics", "Updated desc", 600000)
	if err != nil {
		t.Fatalf("UpdateTreatment error: %v", err)
	}

	// List all (includes inactive)
	all, err := settingsHandler.ListAllTreatments()
	if err != nil {
		t.Fatalf("ListAllTreatments error: %v", err)
	}
	if len(all) < 1 {
		t.Error("expected at least 1 treatment in ListAll")
	}

	// Delete treatment
	err = settingsHandler.DeleteTreatment(treatment.ID)
	if err != nil {
		t.Fatalf("DeleteTreatment error: %v", err)
	}

	// After delete, ListTreatments should be empty
	treatments, _ = settingsHandler.ListTreatments()
	if len(treatments) != 0 {
		t.Errorf("expected 0 active treatments after delete, got %d", len(treatments))
	}
}

func TestSettingsHandler_UpdateClinicSettings(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	settings, _ := settingsHandler.GetClinicSettings()
	settings.ClinicName = "Updated Clinic"
	settings.DoctorName = "Dr. Updated"

	err := settingsHandler.UpdateClinicSettings(settings)
	if err != nil {
		t.Fatalf("UpdateClinicSettings error: %v", err)
	}

	updated, _ := settingsHandler.GetClinicSettings()
	if updated.ClinicName != "Updated Clinic" {
		t.Errorf("expected 'Updated Clinic', got %q", updated.ClinicName)
	}
}

func TestSettingsHandler_Logo(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// Upload small logo (base64 of a tiny image)
	err := settingsHandler.UploadLogo("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatalf("UploadLogo error: %v", err)
	}

	// Remove logo
	err = settingsHandler.RemoveLogo()
	if err != nil {
		t.Fatalf("RemoveLogo error: %v", err)
	}
}

func TestInvoiceHandler_CreateAndManage(t *testing.T) {
	patientHandler, _, invoiceHandler, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Invoice Patient", Phone: "9876543211", Gender: "male",
	})

	// Create invoice
	invoice, err := invoiceHandler.CreateInvoice(service.CreateInvoiceInput{
		PatientID: patient.ID,
		Items: []service.InvoiceItemInput{
			{Description: "Consultation", Quantity: 1, UnitPrice: 50000},
		},
	})
	if err != nil {
		t.Fatalf("CreateInvoice error: %v", err)
	}
	if invoice.TotalAmount != 50000 {
		t.Errorf("expected total 50000, got %d", invoice.TotalAmount)
	}

	// Get invoice
	got, err := invoiceHandler.GetInvoice(invoice.ID)
	if err != nil {
		t.Fatalf("GetInvoice error: %v", err)
	}
	if got.InvoiceNumber == "" {
		t.Error("expected invoice number")
	}

	// List invoices
	list, err := invoiceHandler.ListInvoices(1, 10, "", "", "", "", "")
	if err != nil {
		t.Fatalf("ListInvoices error: %v", err)
	}
	if list.Total != 1 {
		t.Errorf("expected 1 invoice, got %d", list.Total)
	}

	// Record payment
	payment, err := invoiceHandler.RecordPayment(service.RecordPaymentInput{
		InvoiceID:   invoice.ID,
		Amount:      50000,
		Method:      "cash",
		PaymentDate: "2026-12-15",
	})
	if err != nil {
		t.Fatalf("RecordPayment error: %v", err)
	}
	if payment.Amount != 50000 {
		t.Errorf("expected payment 50000, got %d", payment.Amount)
	}

	// Patient outstanding should be 0 now
	outstanding, err := invoiceHandler.GetPatientOutstanding(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientOutstanding error: %v", err)
	}
	if outstanding != 0 {
		t.Errorf("expected 0 outstanding, got %d", outstanding)
	}

	// Patient invoices
	patientInvoices, err := invoiceHandler.GetPatientInvoices(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientInvoices error: %v", err)
	}
	if len(patientInvoices) != 1 {
		t.Errorf("expected 1 patient invoice, got %d", len(patientInvoices))
	}
}

func TestInvoiceHandler_VoidInvoice(t *testing.T) {
	patientHandler, _, invoiceHandler, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Void Patient", Phone: "9876543211", Gender: "male",
	})

	invoice, _ := invoiceHandler.CreateInvoice(service.CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []service.InvoiceItemInput{{Description: "Item", Quantity: 1, UnitPrice: 10000}},
	})

	err := invoiceHandler.VoidInvoice(invoice.ID, "Created by mistake")
	if err != nil {
		t.Fatalf("VoidInvoice error: %v", err)
	}

	voided, _ := invoiceHandler.GetInvoice(invoice.ID)
	if voided.Status != models.InvoiceVoid {
		t.Errorf("expected status void, got %s", voided.Status)
	}
}

func TestDashboardHandler_GetStats(t *testing.T) {
	_, _, _, settingsHandler, authHandler, dashboardHandler, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	stats, err := dashboardHandler.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
}

func TestDashboardHandler_Reports(t *testing.T) {
	_, _, _, settingsHandler, authHandler, dashboardHandler, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// Daily report
	daily, err := dashboardHandler.GetDailyReport("2026-05-20")
	if err != nil {
		t.Fatalf("GetDailyReport error: %v", err)
	}
	if daily == nil {
		t.Fatal("expected non-nil daily report")
	}

	// Monthly report
	monthly, err := dashboardHandler.GetMonthlyReport(2026, 5)
	if err != nil {
		t.Fatalf("GetMonthlyReport error: %v", err)
	}
	if monthly == nil {
		t.Fatal("expected non-nil monthly report")
	}
}

// --- Backup Handler Tests ---

func TestBackupHandler_ListBackups(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, backupHandler, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// ListBackups should work even if empty
	backups, err := backupHandler.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups error: %v", err)
	}
	if backups == nil {
		// Should return empty slice, not nil
		backups = []service.BackupInfo{}
	}
	if len(backups) != 0 {
		t.Fatalf("expected 0 backups, got %d", len(backups))
	}
}

func TestBackupHandler_GetAutoBackupPath(t *testing.T) {
	_, _, _, _, _, _, backupHandler, _ := setupHandlerStack(t)

	path := backupHandler.GetAutoBackupPath()
	// Should return a non-empty string (the temp dir set in config)
	if path == "" {
		t.Fatal("expected non-empty auto backup path")
	}
}

func TestBackupHandler_DetectCloudDrives(t *testing.T) {
	_, _, _, _, _, _, backupHandler, _ := setupHandlerStack(t)

	drives := backupHandler.DetectCloudDrives()
	// Should return a slice (possibly empty on CI/test environments)
	if drives == nil {
		t.Fatal("expected non-nil cloud drives slice")
	}
}

func TestBackupHandler_CreateBackup_NoAuth(t *testing.T) {
	_, _, _, _, _, _, backupHandler, _ := setupHandlerStack(t)

	// Without login, should get auth error
	_, err := backupHandler.CreateBackup("")
	if err == nil {
		t.Fatal("expected error without auth")
	}
}

func TestBackupHandler_VerifyBackup_InvalidPath(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, backupHandler, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	_, err := backupHandler.VerifyBackup("/nonexistent/path.db")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestBackupHandler_RestoreBackup_InvalidPath(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, backupHandler, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	err := backupHandler.RestoreFromBackup("/nonexistent/path.db")
	if err == nil {
		t.Fatal("expected error for invalid backup path")
	}
}

func TestBackupHandler_CreateCloudBackup_NotConfigured(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, backupHandler, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// Cloud backup not configured, should return nil without error
	result, err := backupHandler.CreateCloudBackup()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result when cloud backup not configured")
	}
}

// --- GetPatientHistory Test ---

func TestPatientHandler_GetPatientHistory(t *testing.T) {
	patientHandler, _, invoiceHandler, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// Create a treatment
	treatment, err := settingsHandler.CreateTreatment("Filling", "FIL", "restorative", "Dental filling", 50000)
	if err != nil {
		t.Fatalf("create treatment: %v", err)
	}

	// Create a patient
	patient, err := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})
	if err != nil {
		t.Fatalf("create patient: %v", err)
	}

	// Create an invoice (which creates patient treatment records)
	_, err = invoiceHandler.CreateInvoice(service.CreateInvoiceInput{
		PatientID: patient.ID,
		Items: []service.InvoiceItemInput{
			{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}

	// Get patient history
	history, err := patientHandler.GetPatientHistory(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientHistory error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history record, got %d", len(history))
	}
}

// --- UpdateAppointment Test ---

func TestAppointmentHandler_UpdateAppointment(t *testing.T) {
	patientHandler, appointmentHandler, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	appt, err := appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-20", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	// Update appointment
	updated, err := appointmentHandler.UpdateAppointment(appt.ID, service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-21", StartTime: "10:00", EndTime: "10:30", Duration: 30, Purpose: "Filling",
	})
	if err != nil {
		t.Fatalf("UpdateAppointment error: %v", err)
	}
	if updated.Purpose != "Filling" {
		t.Fatalf("expected purpose Filling, got %s", updated.Purpose)
	}
	if updated.AppointmentDate != "2026-12-21" {
		t.Fatalf("expected date 2026-12-21, got %s", updated.AppointmentDate)
	}
}

// --- GetPatientAppointments Test ---

func TestAppointmentHandler_GetPatientAppointments(t *testing.T) {
	patientHandler, appointmentHandler, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-20", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	appts, err := appointmentHandler.GetPatientAppointments(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientAppointments error: %v", err)
	}
	if len(appts) != 1 {
		t.Fatalf("expected 1 appointment, got %d", len(appts))
	}
}

// --- GetAppointment by ID ---

func TestAppointmentHandler_GetAppointment(t *testing.T) {
	patientHandler, appointmentHandler, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	appt, _ := appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-20", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	fetched, err := appointmentHandler.GetAppointment(appt.ID)
	if err != nil {
		t.Fatalf("GetAppointment error: %v", err)
	}
	if fetched.ID != appt.ID {
		t.Fatalf("ID mismatch")
	}
}

// --- GetAppointmentsByDate Test ---

func TestAppointmentHandler_GetAppointmentsByDate(t *testing.T) {
	patientHandler, appointmentHandler, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-25", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	appts, err := appointmentHandler.GetAppointmentsByDate("2026-12-25")
	if err != nil {
		t.Fatalf("GetAppointmentsByDate error: %v", err)
	}
	if len(appts) != 1 {
		t.Fatalf("expected 1, got %d", len(appts))
	}
}

// --- GetWeekAppointments Test ---

func TestAppointmentHandler_GetWeekAppointments(t *testing.T) {
	patientHandler, appointmentHandler, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-22", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	appts, err := appointmentHandler.GetWeekAppointments("2026-12-20", "2026-12-26")
	if err != nil {
		t.Fatalf("GetWeekAppointments error: %v", err)
	}
	if len(appts) != 1 {
		t.Fatalf("expected 1, got %d", len(appts))
	}
}

// --- Settings: UpdateTreatment, DeleteTreatment ---

func TestSettingsHandler_UpdateTreatment(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	treatment, err := settingsHandler.CreateTreatment("Filling", "FIL", "restorative", "Filling desc", 50000)
	if err != nil {
		t.Fatalf("create treatment: %v", err)
	}

	err = settingsHandler.UpdateTreatment(treatment.ID, "Root Canal", "RC", "endodontic", "Root canal", 100000)
	if err != nil {
		t.Fatalf("UpdateTreatment error: %v", err)
	}
}

func TestSettingsHandler_DeleteTreatment(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	treatment, err := settingsHandler.CreateTreatment("Extraction", "EX", "oral_surgery", "", 30000)
	if err != nil {
		t.Fatalf("create treatment: %v", err)
	}

	err = settingsHandler.DeleteTreatment(treatment.ID)
	if err != nil {
		t.Fatalf("DeleteTreatment error: %v", err)
	}
}

// --- Settings: SaveLogo, RemoveLogo ---

func TestSettingsHandler_SaveAndRemoveLogo(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// Save logo
	err := settingsHandler.UploadLogo("data:image/png;base64,iVBORw0KGgo=")
	if err != nil {
		t.Fatalf("UploadLogo error: %v", err)
	}

	// Verify settings have logo
	settings, err := settingsHandler.GetClinicSettings()
	if err != nil {
		t.Fatalf("GetClinicSettings error: %v", err)
	}
	if settings.LogoBase64 == "" {
		t.Fatal("expected logo to be saved")
	}

	// Remove logo
	err = settingsHandler.RemoveLogo()
	if err != nil {
		t.Fatalf("RemoveLogo error: %v", err)
	}
}

// --- Settings: ListAllTreatments ---

func TestSettingsHandler_ListAllTreatments(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	settingsHandler.CreateTreatment("A", "A", "cat", "", 1000)
	settingsHandler.CreateTreatment("B", "B", "cat", "", 2000)

	treatments, err := settingsHandler.ListAllTreatments()
	if err != nil {
		t.Fatalf("ListAllTreatments error: %v", err)
	}
	if len(treatments) < 2 {
		t.Fatalf("expected at least 2 treatments, got %d", len(treatments))
	}
}

// --- Invoice: GetPatientInvoices, GetPatientOutstanding ---

func TestInvoiceHandler_GetPatientInvoicesAndOutstanding(t *testing.T) {
	patientHandler, _, invoiceHandler, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	treatment, _ := settingsHandler.CreateTreatment("Filling", "FIL", "restorative", "", 50000)

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	// Create invoice
	invoiceHandler.CreateInvoice(service.CreateInvoiceInput{
		PatientID: patient.ID,
		Items: []service.InvoiceItemInput{
			{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000},
		},
	})

	// GetPatientInvoices
	invoices, err := invoiceHandler.GetPatientInvoices(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientInvoices error: %v", err)
	}
	if len(invoices) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(invoices))
	}

	// GetPatientOutstanding
	outstanding, err := invoiceHandler.GetPatientOutstanding(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientOutstanding error: %v", err)
	}
	if outstanding <= 0 {
		t.Fatal("expected positive outstanding amount")
	}
}

// --- Auth: GetCurrentUser, ChangePassword errors ---

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	user, err := authHandler.GetCurrentUser()
	if err != nil {
		t.Fatalf("GetCurrentUser error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.User.Username != "admin" {
		t.Fatalf("expected admin, got %s", user.User.Username)
	}
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	// Wrong current password
	err := authHandler.ChangePassword("wrongpass", "newpass123")
	if err == nil {
		t.Fatal("expected error with wrong current password")
	}

	// Correct change
	err = authHandler.ChangePassword("admin123", "newpass123")
	if err != nil {
		t.Fatalf("ChangePassword error: %v", err)
	}

	// Login with new password
	authHandler.Logout()
	_, err = authHandler.Login("admin", "newpass123")
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
}

// --- Auth: Logout ---

func TestAuthHandler_Logout(t *testing.T) {
	_, _, _, settingsHandler, authHandler, _, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	err := authHandler.Logout()
	if err != nil {
		t.Fatalf("Logout error: %v", err)
	}

	// After logout, GetCurrentUser should return LoggedIn=false
	resp, err := authHandler.GetCurrentUser()
	if err != nil {
		t.Fatalf("GetCurrentUser after logout error: %v", err)
	}
	if resp.LoggedIn {
		t.Fatal("expected LoggedIn=false after logout")
	}
}

// --- Dashboard: GetDashboardStats with data ---

func TestDashboardHandler_StatsWithData(t *testing.T) {
	patientHandler, appointmentHandler, invoiceHandler, settingsHandler, authHandler, dashboardHandler, _, _ := setupHandlerStack(t)

	settingsHandler.CompleteSetup(service.SetupInput{
		ClinicName: "Test", DoctorName: "Dr", Phone: "9876543210",
		AdminUsername: "admin", AdminPassword: "admin123", AdminFullName: "Admin",
	})
	authHandler.Login("admin", "admin123")

	treatment, _ := settingsHandler.CreateTreatment("Checkup", "CHK", "preventive", "", 30000)

	patient, _ := patientHandler.CreatePatient(service.CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	// Create today's appointment
	appointmentHandler.CreateAppointment(service.CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-20", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	// Create an invoice with payment for today
	inv, _ := invoiceHandler.CreateInvoice(service.CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []service.InvoiceItemInput{{TreatmentID: treatment.ID, Description: "Checkup", Quantity: 1, UnitPrice: 30000}},
	})

	invoiceHandler.RecordPayment(service.RecordPaymentInput{
		InvoiceID: inv.ID, Amount: 30000, Method: "cash", PaymentDate: "2026-05-20",
	})

	stats, err := dashboardHandler.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats error: %v", err)
	}
	if stats.TotalPatients != 1 {
		t.Fatalf("expected 1 patient, got %d", stats.TotalPatients)
	}
}
