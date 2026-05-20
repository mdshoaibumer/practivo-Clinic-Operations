package service

import (
	"testing"

	"clinmitra/internal/auth"
	"clinmitra/internal/config"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupCoverageStack sets up a full service stack for coverage tests,
// including DashboardService which isn't in setupFullStack.
func setupCoverageStack(t *testing.T) (
	*AuthService,
	*PatientService,
	*AppointmentService,
	*InvoiceService,
	*SettingsService,
	*DashboardService,
	*AuditService,
) {
	t.Helper()
	dsn := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)
	t.Cleanup(func() { sqlDB.Close() })

	db.AutoMigrate(
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

	cfg := &config.Config{
		AppName:          "Test",
		Version:          "0.0.1",
		DataDir:          t.TempDir(),
		DBPath:           ":memory:",
		BackupDir:        t.TempDir(),
		LogDir:           "",
		MaxLoginAttempts: 5,
		LockoutMinutes:   15,
		SessionHours:     8,
		BcryptCost:       4,
	}

	auditService := NewAuditService(auditRepo)
	authService := NewAuthService(userRepo, sessionManager, loginLimiter, auditService, cfg)
	settingsService := NewSettingsService(clinicRepo, treatmentRepo, authService, auditService, cfg)
	patientService := NewPatientService(patientRepo, patientTreatmentRepo, invoiceRepo, authService, auditService, db)
	appointmentService := NewAppointmentService(appointmentRepo, patientRepo, authService, auditService, db)
	invoiceService := NewInvoiceService(invoiceRepo, invoiceItemRepo, paymentRepo, patientRepo, treatmentRepo, clinicRepo, patientTreatmentRepo, authService, auditService, db)
	dashboardService := NewDashboardService(invoiceRepo, paymentRepo, appointmentRepo, patientRepo)

	return authService, patientService, appointmentService, invoiceService, settingsService, dashboardService, auditService
}

func setupAndLogin(t *testing.T) (
	*AuthService,
	*PatientService,
	*AppointmentService,
	*InvoiceService,
	*SettingsService,
	*DashboardService,
	*AuditService,
) {
	t.Helper()
	authService, patientService, appointmentService, invoiceService, settingsService, dashboardService, auditService := setupCoverageStack(t)

	settingsService.CompleteSetup(SetupInput{
		ClinicName:    "Test Clinic",
		DoctorName:    "Dr. Test",
		Phone:         "9876543210",
		AdminUsername: "admin",
		AdminPassword: "admin123",
		AdminFullName: "Admin User",
	})
	authService.Login("admin", "admin123")

	return authService, patientService, appointmentService, invoiceService, settingsService, dashboardService, auditService
}

// --- DashboardService Tests ---

func TestDashboardService_GetDashboardStats(t *testing.T) {
	_, patientService, appointmentService, invoiceService, settingsService, dashboardService, _ := setupAndLogin(t)

	// Create a treatment
	treatment, _ := settingsService.CreateTreatment("Filling", "FIL", "restorative", "", 50000)

	// Create a patient
	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	// Create appointment for today
	appointmentService.CreateAppointment(CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-20", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	// Create invoice
	inv, _ := invoiceService.CreateInvoice(CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []InvoiceItemInput{{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000}},
	})

	// Record payment
	invoiceService.RecordPayment(RecordPaymentInput{
		InvoiceID: inv.ID, Amount: 30000, Method: "cash", PaymentDate: "2026-05-20",
	})

	stats, err := dashboardService.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats error: %v", err)
	}
	if stats.TotalPatients != 1 {
		t.Fatalf("expected 1 patient, got %d", stats.TotalPatients)
	}
}

func TestDashboardService_GetDailyReport(t *testing.T) {
	_, patientService, _, invoiceService, settingsService, dashboardService, _ := setupAndLogin(t)

	treatment, _ := settingsService.CreateTreatment("Filling", "FIL", "restorative", "", 50000)
	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	inv, _ := invoiceService.CreateInvoice(CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []InvoiceItemInput{{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000}},
	})
	invoiceService.RecordPayment(RecordPaymentInput{
		InvoiceID: inv.ID, Amount: 50000, Method: "cash", PaymentDate: "2026-05-20",
	})

	report, err := dashboardService.GetDailyReport("2026-05-20")
	if err != nil {
		t.Fatalf("GetDailyReport error: %v", err)
	}
	if report.Date != "2026-05-20" {
		t.Fatalf("expected date 2026-05-20, got %s", report.Date)
	}
	if report.TotalCollection != 50000 {
		t.Fatalf("expected 50000, got %d", report.TotalCollection)
	}
}

func TestDashboardService_GetMonthlyReport(t *testing.T) {
	_, patientService, _, invoiceService, settingsService, dashboardService, _ := setupAndLogin(t)

	treatment, _ := settingsService.CreateTreatment("Filling", "FIL", "restorative", "", 50000)
	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	invoiceService.CreateInvoice(CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []InvoiceItemInput{{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000}},
	})

	report, err := dashboardService.GetMonthlyReport(2026, 5)
	if err != nil {
		t.Fatalf("GetMonthlyReport error: %v", err)
	}
	if report.Year != 2026 || report.Month != 5 {
		t.Fatalf("wrong year/month: %d/%d", report.Year, report.Month)
	}
}

// --- Settings Service Coverage ---

func TestSettingsService_GetClinicSettings(t *testing.T) {
	_, _, _, _, settingsService, _, _ := setupAndLogin(t)

	settings, err := settingsService.GetClinicSettings()
	if err != nil {
		t.Fatalf("GetClinicSettings error: %v", err)
	}
	if settings.ClinicName != "Test Clinic" {
		t.Fatalf("expected Test Clinic, got %s", settings.ClinicName)
	}
}

func TestSettingsService_UpdateClinicSettings(t *testing.T) {
	_, _, _, _, settingsService, _, _ := setupAndLogin(t)

	settings, _ := settingsService.GetClinicSettings()
	settings.ClinicName = "Updated Clinic"
	settings.DoctorName = "Dr. Updated"

	err := settingsService.UpdateClinicSettings(settings)
	if err != nil {
		t.Fatalf("UpdateClinicSettings error: %v", err)
	}

	updated, _ := settingsService.GetClinicSettings()
	if updated.ClinicName != "Updated Clinic" {
		t.Fatalf("expected Updated Clinic, got %s", updated.ClinicName)
	}
}

func TestSettingsService_ListAllTreatments(t *testing.T) {
	_, _, _, _, settingsService, _, _ := setupAndLogin(t)

	settingsService.CreateTreatment("A", "A", "cat", "", 1000)
	settingsService.CreateTreatment("B", "B", "cat", "", 2000)

	treatments, err := settingsService.ListAllTreatments()
	if err != nil {
		t.Fatalf("ListAllTreatments error: %v", err)
	}
	if len(treatments) < 2 {
		t.Fatalf("expected at least 2, got %d", len(treatments))
	}
}

func TestSettingsService_UpdateTreatment(t *testing.T) {
	_, _, _, _, settingsService, _, _ := setupAndLogin(t)

	treatment, _ := settingsService.CreateTreatment("Filling", "FIL", "restorative", "", 50000)

	err := settingsService.UpdateTreatment(treatment.ID, "Root Canal", "RC", "endodontic", "Root canal procedure", 100000)
	if err != nil {
		t.Fatalf("UpdateTreatment error: %v", err)
	}
}

func TestSettingsService_DeleteTreatment(t *testing.T) {
	_, _, _, _, settingsService, _, _ := setupAndLogin(t)

	treatment, _ := settingsService.CreateTreatment("Temp", "TMP", "cat", "", 1000)

	err := settingsService.DeleteTreatment(treatment.ID)
	if err != nil {
		t.Fatalf("DeleteTreatment error: %v", err)
	}
}

func TestSettingsService_SaveAndRemoveLogo(t *testing.T) {
	_, _, _, _, settingsService, _, _ := setupAndLogin(t)

	err := settingsService.SaveLogo("data:image/png;base64,iVBORw0KGgo=")
	if err != nil {
		t.Fatalf("SaveLogo error: %v", err)
	}

	settings, _ := settingsService.GetClinicSettings()
	if settings.LogoBase64 == "" {
		t.Fatal("expected logo to be saved")
	}

	err = settingsService.RemoveLogo()
	if err != nil {
		t.Fatalf("RemoveLogo error: %v", err)
	}

	settings, _ = settingsService.GetClinicSettings()
	if settings.LogoBase64 != "" {
		t.Fatal("expected logo to be removed")
	}
}

// --- Patient Service Coverage ---

func TestPatientService_GetPatient(t *testing.T) {
	_, patientService, _, _, _, _, _ := setupAndLogin(t)

	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test Patient", Phone: "9876543211", Gender: "male", Age: 30,
	})

	fetched, err := patientService.GetPatient(patient.ID)
	if err != nil {
		t.Fatalf("GetPatient error: %v", err)
	}
	if fetched.ID != patient.ID {
		t.Fatal("ID mismatch")
	}
}

func TestPatientService_ListPatients_Pagination(t *testing.T) {
	_, patientService, _, _, _, _, _ := setupAndLogin(t)

	patientService.CreatePatient(CreatePatientInput{
		Name: "Alice Smith", Phone: "9876543211", Gender: "female", Age: 25,
	})
	patientService.CreatePatient(CreatePatientInput{
		Name: "Bob Jones", Phone: "9876543212", Gender: "male", Age: 35,
	})

	result, err := patientService.ListPatients(1, 10, "")
	if err != nil {
		t.Fatalf("ListPatients error: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2, got %d", result.Total)
	}

	// Search
	result, err = patientService.ListPatients(1, 10, "Alice")
	if err != nil {
		t.Fatalf("ListPatients search error: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1, got %d", result.Total)
	}

	// Edge cases for pagination
	result, _ = patientService.ListPatients(0, 0, "")
	if result.Page != 1 || result.PageSize != 20 {
		t.Fatalf("expected defaults page=1 pageSize=20, got %d/%d", result.Page, result.PageSize)
	}

	result, _ = patientService.ListPatients(1, 200, "")
	if result.PageSize != 100 {
		t.Fatalf("expected max pageSize=100, got %d", result.PageSize)
	}
}

func TestPatientService_DeletePatient(t *testing.T) {
	_, patientService, _, _, _, _, _ := setupAndLogin(t)

	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Delete Me", Phone: "9876543211", Gender: "male", Age: 40,
	})

	err := patientService.DeletePatient(patient.ID)
	if err != nil {
		t.Fatalf("DeletePatient error: %v", err)
	}
}

func TestPatientService_CheckDuplicatePhone(t *testing.T) {
	_, patientService, _, _, _, _, _ := setupAndLogin(t)

	patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	found, err := patientService.CheckDuplicatePhone("9876543211")
	if err != nil {
		t.Fatalf("CheckDuplicatePhone error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find duplicate")
	}
}

func TestPatientService_GetPatientCount(t *testing.T) {
	_, patientService, _, _, _, _, _ := setupAndLogin(t)

	patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	count, err := patientService.GetPatientCount()
	if err != nil {
		t.Fatalf("GetPatientCount error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
}

func TestPatientService_GetPatientHistory(t *testing.T) {
	_, patientService, _, invoiceService, settingsService, _, _ := setupAndLogin(t)

	treatment, _ := settingsService.CreateTreatment("Filling", "FIL", "restorative", "", 50000)
	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	invoiceService.CreateInvoice(CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []InvoiceItemInput{{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000}},
	})

	history, err := patientService.GetPatientHistory(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientHistory error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1, got %d", len(history))
	}
}

// --- Appointment Service Coverage ---

func TestAppointmentService_UpdateAppointment(t *testing.T) {
	_, patientService, appointmentService, _, _, _, _ := setupAndLogin(t)

	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	appt, _ := appointmentService.CreateAppointment(CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-20", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	updated, err := appointmentService.UpdateAppointment(appt.ID, CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-21", StartTime: "10:00", EndTime: "10:30", Duration: 30, Purpose: "Filling",
	})
	if err != nil {
		t.Fatalf("UpdateAppointment error: %v", err)
	}
	if updated.Purpose != "Filling" {
		t.Fatalf("expected Filling, got %s", updated.Purpose)
	}
}

func TestAppointmentService_GetPatientAppointments(t *testing.T) {
	_, patientService, appointmentService, _, _, _, _ := setupAndLogin(t)

	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	appointmentService.CreateAppointment(CreateAppointmentInput{
		PatientID: patient.ID, Date: "2026-12-20", StartTime: "09:00", EndTime: "09:30", Duration: 30, Purpose: "Checkup",
	})

	appts, err := appointmentService.GetPatientAppointments(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientAppointments error: %v", err)
	}
	if len(appts) != 1 {
		t.Fatalf("expected 1, got %d", len(appts))
	}
}

// --- Invoice Service Coverage ---

func TestInvoiceService_GetPatientInvoices(t *testing.T) {
	_, patientService, _, invoiceService, settingsService, _, _ := setupAndLogin(t)

	treatment, _ := settingsService.CreateTreatment("Filling", "FIL", "restorative", "", 50000)
	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	invoiceService.CreateInvoice(CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []InvoiceItemInput{{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000}},
	})

	invoices, err := invoiceService.GetPatientInvoices(patient.ID)
	if err != nil {
		t.Fatalf("GetPatientInvoices error: %v", err)
	}
	if len(invoices) != 1 {
		t.Fatalf("expected 1, got %d", len(invoices))
	}
}

func TestInvoiceService_ListInvoicesFilters(t *testing.T) {
	_, patientService, _, invoiceService, settingsService, _, _ := setupAndLogin(t)

	treatment, _ := settingsService.CreateTreatment("Filling", "FIL", "restorative", "", 50000)
	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	invoiceService.CreateInvoice(CreateInvoiceInput{
		PatientID: patient.ID,
		Items:     []InvoiceItemInput{{TreatmentID: treatment.ID, Description: "Filling", Quantity: 1, UnitPrice: 50000}},
	})

	// Filter by status
	result, err := invoiceService.ListInvoices(1, 10, "issued", "", "", "", "")
	if err != nil {
		t.Fatalf("ListInvoices error: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1, got %d", result.Total)
	}

	// Filter by patient
	result, _ = invoiceService.ListInvoices(1, 10, "", "", "", patient.ID, "")
	if result.Total != 1 {
		t.Fatalf("expected 1 by patient, got %d", result.Total)
	}
}

// --- Audit Service Coverage ---

func TestAuditService_GetEntityHistory(t *testing.T) {
	_, patientService, _, _, _, _, auditService := setupAndLogin(t)

	patient, _ := patientService.CreatePatient(CreatePatientInput{
		Name: "Test", Phone: "9876543211", Gender: "male", Age: 30,
	})

	// Wait a moment for async audit log
	history, err := auditService.GetEntityHistory("patient", patient.ID)
	if err != nil {
		t.Fatalf("GetEntityHistory error: %v", err)
	}
	// May be empty since audit is async, just check no error
	_ = history
}

func TestAuditService_GetRecentActivity(t *testing.T) {
	_, _, _, _, _, _, auditService := setupAndLogin(t)

	activity, err := auditService.GetRecentActivity(10)
	if err != nil {
		t.Fatalf("GetRecentActivity error: %v", err)
	}
	_ = activity
}

// --- Auth Service: GetCurrentUser after logout ---

func TestAuthService_GetCurrentUser_AfterLogout(t *testing.T) {
	authService, _, _, _, _, _, _ := setupAndLogin(t)

	authService.Logout()

	resp, err := authService.GetCurrentUser()
	if err != nil {
		t.Fatalf("GetCurrentUser error: %v", err)
	}
	if resp.LoggedIn {
		t.Fatal("expected LoggedIn=false after logout")
	}
}
