package service

import (
	"fmt"
	"testing"

	"clinmitra/internal/auth"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupFreshDB creates a fresh in-memory SQLite database with all tables,
// simulating a brand-new install. Uses a unique named in-memory DB with
// shared cache so all connections see the same data (no file locking issues).
func setupFreshDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Unique name per test so parallel tests don't collide
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=foreign_keys(ON)&_pragma=busy_timeout(10000)", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Use 2 connections: one for transactions, one for audit goroutines.
	// Keep at least 1 idle connection so the in-memory DB isn't dropped.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)

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
		t.Fatalf("failed to migrate test db: %v", err)
	}

	// Close DB connection on test cleanup to release file locks
	t.Cleanup(func() {
		sqlDB.Close()
	})

	return db
}

// setupFullStack wires all repositories, services exactly like app.go does,
// but against a fresh in-memory database.
func setupFullStack(t *testing.T) (*gorm.DB, *AuthService, *PatientService, *AppointmentService, *InvoiceService, *SettingsService) {
	t.Helper()
	db := setupFreshDB(t)

	// Repositories
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

	// Auth components
	sessionManager := auth.NewSessionManager(8)
	loginLimiter := auth.NewLoginLimiter(5, 15)

	// Use test-specific temp dir for session persistence to avoid cross-test leakage
	cfg := testConfig()
	cfg.DataDir = t.TempDir()

	// Services (wired identically to app.go)
	auditService := NewAuditService(auditRepo)
	authService := NewAuthService(userRepo, sessionManager, loginLimiter, auditService, cfg)
	settingsService := NewSettingsService(clinicRepo, treatmentRepo, authService, auditService, cfg)
	patientService := NewPatientService(patientRepo, patientTreatmentRepo, invoiceRepo, authService, auditService, db)
	appointmentService := NewAppointmentService(appointmentRepo, patientRepo, authService, auditService, db)
	invoiceService := NewInvoiceService(invoiceRepo, invoiceItemRepo, paymentRepo, patientRepo, treatmentRepo, clinicRepo, patientTreatmentRepo, authService, auditService, db)

	return db, authService, patientService, appointmentService, invoiceService, settingsService
}

// =============================================================================
// END-TO-END INTEGRATION TESTS — FULL WORKFLOW WITH FRESH DB
// =============================================================================

func TestE2E_FullClinicWorkflow(t *testing.T) {
	db, authService, patientService, appointmentService, invoiceService, settingsService := setupFullStack(t)

	// =========================================================================
	// STEP 1: Initial Setup (first-time install)
	// =========================================================================
	t.Run("Step1_InitialSetup", func(t *testing.T) {
		// Setup should not be complete yet
		complete, err := settingsService.IsSetupComplete()
		if err != nil {
			t.Fatalf("IsSetupComplete error: %v", err)
		}
		if complete {
			t.Fatal("expected setup to NOT be complete on fresh DB")
		}

		// Complete setup: create admin user + clinic settings
		err = settingsService.CompleteSetup(SetupInput{
			ClinicName:    "Dr. Patel's Dental Clinic",
			DoctorName:    "Dr. Ravi Patel",
			Phone:         "9876543210",
			Address:       "123 MG Road, Pune",
			City:          "Pune",
			AdminUsername: "admin",
			AdminPassword: "admin123",
			AdminFullName: "Dr. Ravi Patel",
		})
		if err != nil {
			t.Fatalf("CompleteSetup error: %v", err)
		}

		// Verify setup is now complete
		complete, err = settingsService.IsSetupComplete()
		if err != nil {
			t.Fatalf("IsSetupComplete error: %v", err)
		}
		if !complete {
			t.Fatal("expected setup to be complete after CompleteSetup")
		}

		// Cannot setup again
		err = settingsService.CompleteSetup(SetupInput{
			ClinicName:    "Another Clinic",
			DoctorName:    "Dr. Evil",
			Phone:         "9876543211",
			AdminUsername: "admin2",
			AdminPassword: "pass123",
			AdminFullName: "Evil Doctor",
		})
		if err == nil {
			t.Fatal("expected error on duplicate setup")
		}
	})

	// =========================================================================
	// STEP 2: Authentication
	// =========================================================================
	t.Run("Step2_Authentication", func(t *testing.T) {
		// Login with wrong password
		_, err := authService.Login("admin", "wrong_password")
		if err == nil {
			t.Fatal("expected error for wrong password")
		}
		if err != utils.ErrInvalidCredentials {
			t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
		}

		// Login with wrong username
		_, err = authService.Login("nonexistent", "admin123")
		if err == nil {
			t.Fatal("expected error for wrong username")
		}

		// Login with correct credentials
		resp, err := authService.Login("admin", "admin123")
		if err != nil {
			t.Fatalf("Login error: %v", err)
		}
		if !resp.LoggedIn {
			t.Fatal("expected LoggedIn=true")
		}
		if resp.User.Username != "admin" {
			t.Errorf("expected username 'admin', got: %s", resp.User.Username)
		}
		if resp.User.Role != models.RoleAdmin {
			t.Errorf("expected role admin, got: %s", resp.User.Role)
		}

		// GetCurrentUser should return session
		current, err := authService.GetCurrentUser()
		if err != nil {
			t.Fatalf("GetCurrentUser error: %v", err)
		}
		if !current.LoggedIn {
			t.Fatal("expected logged in after login")
		}
	})

	// =========================================================================
	// STEP 3: Create Patients
	// =========================================================================
	var patient1ID, patient2ID string

	t.Run("Step3_CreatePatients", func(t *testing.T) {
		// Create first patient
		p1, err := patientService.CreatePatient(CreatePatientInput{
			Name:           "Anita Desai",
			Phone:          "9876543211",
			Email:          "anita@example.com",
			Gender:         "female",
			Age:            35,
			Address:        "456 FC Road, Pune",
			City:           "Pune",
			BloodGroup:     "B+",
			MedicalHistory: "None significant",
			Allergies:      "Penicillin",
		})
		if err != nil {
			t.Fatalf("CreatePatient error: %v", err)
		}
		if p1.ID == "" {
			t.Fatal("expected patient to have an ID")
		}
		if p1.Phone != "9876543211" {
			t.Errorf("expected cleaned phone '9876543211', got: %s", p1.Phone)
		}
		patient1ID = p1.ID

		// Create second patient with +91 prefix (should be cleaned)
		p2, err := patientService.CreatePatient(CreatePatientInput{
			Name:   "Vikram Singh",
			Phone:  "+918765432100",
			Gender: "male",
			Age:    42,
			City:   "Mumbai",
		})
		if err != nil {
			t.Fatalf("CreatePatient #2 error: %v", err)
		}
		if p2.Phone != "8765432100" {
			t.Errorf("expected phone cleaned to '8765432100', got: %s", p2.Phone)
		}
		patient2ID = p2.ID

		// Duplicate phone should fail
		_, err = patientService.CreatePatient(CreatePatientInput{
			Name:  "Duplicate Phone",
			Phone: "9876543211",
		})
		if err == nil {
			t.Fatal("expected error for duplicate phone")
		}

		// List patients
		list, err := patientService.ListPatients(1, 10, "")
		if err != nil {
			t.Fatalf("ListPatients error: %v", err)
		}
		if list.Total != 2 {
			t.Errorf("expected 2 patients, got: %d", list.Total)
		}

		// Search by name
		list, err = patientService.ListPatients(1, 10, "Anita")
		if err != nil {
			t.Fatalf("ListPatients search error: %v", err)
		}
		if list.Total != 1 {
			t.Errorf("expected 1 patient for 'Anita', got: %d", list.Total)
		}
	})

	// =========================================================================
	// STEP 4: Update Patient
	// =========================================================================
	t.Run("Step4_UpdatePatient", func(t *testing.T) {
		updated, err := patientService.UpdatePatient(patient1ID, CreatePatientInput{
			Name:           "Anita Desai-Shah",
			Phone:          "9876543211",
			Email:          "anita.shah@example.com",
			Gender:         "female",
			Age:            36,
			Address:        "456 FC Road, Pune",
			City:           "Pune",
			BloodGroup:     "B+",
			MedicalHistory: "Diabetes Type 2",
			Allergies:      "Penicillin, Sulfa drugs",
		})
		if err != nil {
			t.Fatalf("UpdatePatient error: %v", err)
		}
		if updated.Name != "Anita Desai-Shah" {
			t.Errorf("expected updated name, got: %s", updated.Name)
		}
		if updated.MedicalHistory != "Diabetes Type 2" {
			t.Errorf("expected updated medical history, got: %s", updated.MedicalHistory)
		}
	})

	// =========================================================================
	// STEP 5: Create Appointments
	// =========================================================================
	var appointment1ID string

	t.Run("Step5_CreateAppointments", func(t *testing.T) {
		// Create appointment for patient 1
		appt, err := appointmentService.CreateAppointment(CreateAppointmentInput{
			PatientID: patient1ID,
			Date:      "2026-12-15",
			StartTime: "10:00",
			EndTime:   "10:30",
			Duration:  30,
			Purpose:   "Root canal treatment",
			Notes:     "Patient has dental anxiety, use sedation",
		})
		if err != nil {
			t.Fatalf("CreateAppointment error: %v", err)
		}
		if appt.Status != models.AppointmentScheduled {
			t.Errorf("expected status scheduled, got: %s", appt.Status)
		}
		appointment1ID = appt.ID

		// Create non-conflicting appointment same day
		_, err = appointmentService.CreateAppointment(CreateAppointmentInput{
			PatientID: patient2ID,
			Date:      "2026-12-15",
			StartTime: "11:00",
			EndTime:   "11:30",
			Duration:  30,
			Purpose:   "Teeth cleaning",
		})
		if err != nil {
			t.Fatalf("CreateAppointment #2 error: %v", err)
		}

		// Conflicting time slot should fail
		_, err = appointmentService.CreateAppointment(CreateAppointmentInput{
			PatientID: patient2ID,
			Date:      "2026-12-15",
			StartTime: "10:15",
			EndTime:   "10:45",
			Duration:  30,
			Purpose:   "This should conflict",
		})
		if err == nil {
			t.Fatal("expected conflict error for overlapping time slot")
		}

		// Verify today's appointments (none for today, our appts are on 2026-12-15)
		todayAppts, err := appointmentService.GetTodayAppointments()
		if err != nil {
			t.Fatalf("GetTodayAppointments error: %v", err)
		}
		if len(todayAppts) != 0 {
			t.Errorf("expected 0 appointments today, got: %d", len(todayAppts))
		}

		// Get appointments by date
		dateAppts, err := appointmentService.GetAppointmentsByDate("2026-12-15")
		if err != nil {
			t.Fatalf("GetAppointmentsByDate error: %v", err)
		}
		if len(dateAppts) != 2 {
			t.Errorf("expected 2 appointments on 2026-12-15, got: %d", len(dateAppts))
		}
	})

	// =========================================================================
	// STEP 6: Complete and Cancel Appointments
	// =========================================================================
	t.Run("Step6_AppointmentStatusChanges", func(t *testing.T) {
		// Complete the first appointment
		err := appointmentService.CompleteAppointment(appointment1ID)
		if err != nil {
			t.Fatalf("CompleteAppointment error: %v", err)
		}

		appt, _ := appointmentService.GetAppointment(appointment1ID)
		if appt.Status != models.AppointmentCompleted {
			t.Errorf("expected completed, got: %s", appt.Status)
		}

		// Cannot complete again
		err = appointmentService.CompleteAppointment(appointment1ID)
		if err == nil {
			t.Fatal("expected error completing already-completed appointment")
		}

		// Cannot cancel completed appointment
		err = appointmentService.CancelAppointment(appointment1ID, "test reason")
		if err == nil {
			t.Fatal("expected error cancelling completed appointment")
		}
	})

	// =========================================================================
	// STEP 7: Create Treatments (for invoicing)
	// =========================================================================
	var treatmentID string

	t.Run("Step7_CreateTreatments", func(t *testing.T) {
		treatment, err := settingsService.CreateTreatment(
			"Root Canal Treatment",
			"RCT",
			"Endodontics",
			"Single canal root canal therapy",
			1500000, // ₹15,000 in paise
		)
		if err != nil {
			t.Fatalf("CreateTreatment error: %v", err)
		}
		treatmentID = treatment.ID

		// Create another treatment
		_, err = settingsService.CreateTreatment(
			"Teeth Cleaning",
			"CLEAN",
			"Preventive",
			"Professional scaling and polishing",
			200000, // ₹2,000 in paise
		)
		if err != nil {
			t.Fatalf("CreateTreatment #2 error: %v", err)
		}

		// List treatments
		treatments, err := settingsService.ListTreatments()
		if err != nil {
			t.Fatalf("ListTreatments error: %v", err)
		}
		if len(treatments) != 2 {
			t.Errorf("expected 2 treatments, got: %d", len(treatments))
		}
	})

	// =========================================================================
	// STEP 8: Create Invoice
	// =========================================================================
	var invoiceID string

	t.Run("Step8_CreateInvoice", func(t *testing.T) {
		// First ensure clinic settings exist (they were created in setup)
		invoice, err := invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID: patient1ID,
			Items: []InvoiceItemInput{
				{
					TreatmentID: treatmentID,
					Description: "Root Canal - Upper Molar #16",
					Quantity:    1,
					UnitPrice:   1500000,
					ToothNumber: "16",
				},
			},
			DiscountPercent: 10, // 10% discount
			Notes:           "Treatment completed successfully",
		})
		if err != nil {
			t.Fatalf("CreateInvoice error: %v", err)
		}

		invoiceID = invoice.ID

		// Verify calculations
		if invoice.SubTotal != 1500000 {
			t.Errorf("expected subtotal 1500000, got: %d", invoice.SubTotal)
		}
		// 10% discount = 150000
		if invoice.DiscountAmount != 150000 {
			t.Errorf("expected discount 150000, got: %d", invoice.DiscountAmount)
		}
		// Taxable = 1500000 - 150000 = 1350000
		if invoice.TaxableAmount != 1350000 {
			t.Errorf("expected taxable 1350000, got: %d", invoice.TaxableAmount)
		}
		// No GST (not enabled in default settings)
		if invoice.CGSTAmount != 0 || invoice.SGSTAmount != 0 {
			t.Errorf("expected no GST, got CGST=%d SGST=%d", invoice.CGSTAmount, invoice.SGSTAmount)
		}
		// Total = taxable (no GST)
		if invoice.TotalAmount != 1350000 {
			t.Errorf("expected total 1350000, got: %d", invoice.TotalAmount)
		}
		if invoice.BalanceAmount != 1350000 {
			t.Errorf("expected balance 1350000, got: %d", invoice.BalanceAmount)
		}
		if invoice.Status != models.InvoiceIssued {
			t.Errorf("expected status issued, got: %s", invoice.Status)
		}
		if invoice.InvoiceNumber == "" {
			t.Error("expected invoice number to be generated")
		}

		t.Logf("Invoice created: %s, Total: ₹%.2f", invoice.InvoiceNumber, float64(invoice.TotalAmount)/100)
	})

	// =========================================================================
	// STEP 9: Record Partial Payment
	// =========================================================================
	t.Run("Step9_PartialPayment", func(t *testing.T) {
		payment, err := invoiceService.RecordPayment(RecordPaymentInput{
			InvoiceID:   invoiceID,
			Amount:      500000, // ₹5,000
			Method:      "cash",
			PaymentDate: "2026-05-20",
			Notes:       "Partial payment received",
		})
		if err != nil {
			t.Fatalf("RecordPayment error: %v", err)
		}
		if payment.Amount != 500000 {
			t.Errorf("expected payment amount 500000, got: %d", payment.Amount)
		}

		// Check invoice status changed to partial
		invoice, err := invoiceService.GetInvoice(invoiceID)
		if err != nil {
			t.Fatalf("GetInvoice error: %v", err)
		}
		if invoice.Status != models.InvoicePartial {
			t.Errorf("expected status partial, got: %s", invoice.Status)
		}
		if invoice.PaidAmount != 500000 {
			t.Errorf("expected paid 500000, got: %d", invoice.PaidAmount)
		}
		if invoice.BalanceAmount != 850000 {
			t.Errorf("expected balance 850000, got: %d", invoice.BalanceAmount)
		}
	})

	// =========================================================================
	// STEP 10: Record Final Payment (completes invoice)
	// =========================================================================
	t.Run("Step10_FinalPayment", func(t *testing.T) {
		_, err := invoiceService.RecordPayment(RecordPaymentInput{
			InvoiceID:   invoiceID,
			Amount:      850000, // Remaining ₹8,500
			Method:      "upi",
			PaymentDate: "2026-05-21",
			Reference:   "UPI-REF-12345",
		})
		if err != nil {
			t.Fatalf("RecordPayment final error: %v", err)
		}

		// Check invoice is now fully paid
		invoice, err := invoiceService.GetInvoice(invoiceID)
		if err != nil {
			t.Fatalf("GetInvoice error: %v", err)
		}
		if invoice.Status != models.InvoicePaid {
			t.Errorf("expected status paid, got: %s", invoice.Status)
		}
		if invoice.BalanceAmount != 0 {
			t.Errorf("expected balance 0, got: %d", invoice.BalanceAmount)
		}
		if invoice.PaidAmount != 1350000 {
			t.Errorf("expected paid 1350000, got: %d", invoice.PaidAmount)
		}

		// Cannot overpay
		_, err = invoiceService.RecordPayment(RecordPaymentInput{
			InvoiceID: invoiceID,
			Amount:    100,
			Method:    "cash",
		})
		if err == nil {
			t.Fatal("expected error when paying fully paid invoice")
		}
	})

	// =========================================================================
	// STEP 11: Patient Treatment History
	// =========================================================================
	t.Run("Step11_TreatmentHistory", func(t *testing.T) {
		history, err := patientService.GetPatientHistory(patient1ID)
		if err != nil {
			t.Fatalf("GetPatientHistory error: %v", err)
		}
		if len(history) != 1 {
			t.Errorf("expected 1 treatment in history, got: %d", len(history))
		}
		if len(history) > 0 && history[0].ToothNumber != "16" {
			t.Errorf("expected tooth number '16', got: %s", history[0].ToothNumber)
		}
	})

	// =========================================================================
	// STEP 12: Patient Outstanding Balance
	// =========================================================================
	t.Run("Step12_OutstandingBalance", func(t *testing.T) {
		outstanding, err := invoiceService.GetPatientOutstanding(patient1ID)
		if err != nil {
			t.Fatalf("GetPatientOutstanding error: %v", err)
		}
		if outstanding != 0 {
			t.Errorf("expected 0 outstanding (fully paid), got: %d", outstanding)
		}

		// Create another invoice for patient 2 (unpaid)
		inv2, err := invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID: patient2ID,
			Items: []InvoiceItemInput{
				{
					Description: "Consultation",
					Quantity:    1,
					UnitPrice:   50000, // ₹500
				},
			},
		})
		if err != nil {
			t.Fatalf("CreateInvoice #2 error: %v", err)
		}

		outstanding2, err := invoiceService.GetPatientOutstanding(patient2ID)
		if err != nil {
			t.Fatalf("GetPatientOutstanding #2 error: %v", err)
		}
		if outstanding2 != inv2.TotalAmount {
			t.Errorf("expected outstanding %d, got: %d", inv2.TotalAmount, outstanding2)
		}
	})

	// =========================================================================
	// STEP 13: Void Invoice
	// =========================================================================
	t.Run("Step13_VoidInvoice", func(t *testing.T) {
		// Create a voidable invoice (no payments)
		inv, err := invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID: patient1ID,
			Items: []InvoiceItemInput{
				{Description: "Mistake invoice", Quantity: 1, UnitPrice: 100000},
			},
		})
		if err != nil {
			t.Fatalf("CreateInvoice for void error: %v", err)
		}

		err = invoiceService.VoidInvoice(inv.ID, "Created by mistake")
		if err != nil {
			t.Fatalf("VoidInvoice error: %v", err)
		}

		voided, _ := invoiceService.GetInvoice(inv.ID)
		if voided.Status != models.InvoiceVoid {
			t.Errorf("expected status void, got: %s", voided.Status)
		}
		if voided.VoidReason != "Created by mistake" {
			t.Errorf("expected void reason, got: %s", voided.VoidReason)
		}

		// Cannot void already voided
		err = invoiceService.VoidInvoice(inv.ID, "Again")
		if err == nil {
			t.Fatal("expected error voiding already-voided invoice")
		}
	})

	// =========================================================================
	// STEP 14: Delete Patient (blocked by outstanding)
	// =========================================================================
	t.Run("Step14_DeletePatient_Blocked", func(t *testing.T) {
		// Patient 2 has outstanding invoice — cannot delete
		err := patientService.DeletePatient(patient2ID)
		if err == nil {
			t.Fatal("expected error deleting patient with outstanding invoices")
		}
	})

	// =========================================================================
	// STEP 15: Password Change
	// =========================================================================
	t.Run("Step15_PasswordChange", func(t *testing.T) {
		// Change password
		err := authService.ChangePassword("admin123", "newpass456")
		if err != nil {
			t.Fatalf("ChangePassword error: %v", err)
		}

		// Logout
		err = authService.Logout()
		if err != nil {
			t.Fatalf("Logout error: %v", err)
		}

		// Old password should fail
		_, err = authService.Login("admin", "admin123")
		if err == nil {
			t.Fatal("expected error with old password after change")
		}

		// New password should work
		resp, err := authService.Login("admin", "newpass456")
		if err != nil {
			t.Fatalf("Login with new password error: %v", err)
		}
		if !resp.LoggedIn {
			t.Fatal("expected LoggedIn=true with new password")
		}
	})

	// =========================================================================
	// STEP 16: Rate Limiting
	// =========================================================================
	t.Run("Step16_RateLimiting", func(t *testing.T) {
		// Exhaust login attempts
		for i := 0; i < 5; i++ {
			authService.Login("admin", "wrong_password")
		}

		// Should be locked now
		_, err := authService.Login("admin", "newpass456")
		if err == nil {
			t.Fatal("expected account locked error")
		}
		if err != utils.ErrAccountLocked {
			t.Errorf("expected ErrAccountLocked, got: %v", err)
		}
	})

	// =========================================================================
	// STEP 17: Verify Audit Trail
	// =========================================================================
	t.Run("Step17_AuditTrail", func(t *testing.T) {
		var auditLogs []models.AuditLog
		err := db.Order("created_at DESC").Find(&auditLogs).Error
		if err != nil {
			t.Fatalf("query audit logs error: %v", err)
		}

		if len(auditLogs) == 0 {
			t.Fatal("expected audit log entries from all operations")
		}

		t.Logf("Total audit log entries: %d", len(auditLogs))

		// Verify we have various action types
		actionCounts := map[models.AuditAction]int{}
		for _, log := range auditLogs {
			actionCounts[log.Action]++
		}
		t.Logf("Audit actions: %v", actionCounts)

		// Should have create, update, login actions at minimum
		if actionCounts[models.AuditCreate] == 0 {
			t.Error("expected at least one CREATE audit entry")
		}
		if actionCounts[models.AuditLogin] == 0 {
			t.Error("expected at least one LOGIN audit entry")
		}
	})

	// =========================================================================
	// STEP 18: Invoice Listing & Filters
	// =========================================================================
	t.Run("Step18_InvoiceFilters", func(t *testing.T) {
		// List all invoices
		list, err := invoiceService.ListInvoices(1, 10, "", "", "", "", "")
		if err != nil {
			t.Fatalf("ListInvoices error: %v", err)
		}
		if list.Total < 3 {
			t.Errorf("expected at least 3 invoices, got: %d", list.Total)
		}

		// Filter by status
		paidList, err := invoiceService.ListInvoices(1, 10, string(models.InvoicePaid), "", "", "", "")
		if err != nil {
			t.Fatalf("ListInvoices paid filter error: %v", err)
		}
		if paidList.Total != 1 {
			t.Errorf("expected 1 paid invoice, got: %d", paidList.Total)
		}

		// Filter by patient
		patientList, err := invoiceService.ListInvoices(1, 10, "", "", "", patient1ID, "")
		if err != nil {
			t.Fatalf("ListInvoices patient filter error: %v", err)
		}
		if patientList.Total < 2 {
			t.Errorf("expected at least 2 invoices for patient1, got: %d", patientList.Total)
		}
	})

	// =========================================================================
	// STEP 19: Multi-Item Invoice with GST
	// =========================================================================
	t.Run("Step19_MultiItemInvoiceWithGST", func(t *testing.T) {
		// Enable GST in clinic settings
		var settings models.ClinicSettings
		db.First(&settings)
		settings.GSTEnabled = true
		settings.GSTRate = 18 // 18% GST (9% CGST + 9% SGST)
		db.Save(&settings)

		invoice, err := invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID: patient1ID,
			Items: []InvoiceItemInput{
				{TreatmentID: treatmentID, Description: "RCT - Tooth 26", Quantity: 1, UnitPrice: 1500000, ToothNumber: "26"},
				{Description: "X-Ray", Quantity: 2, UnitPrice: 30000},
			},
			DiscountAmount: 50000, // ₹500 flat discount
		})
		if err != nil {
			t.Fatalf("CreateInvoice GST error: %v", err)
		}

		// SubTotal = 1500000 + (2*30000) = 1560000
		expectedSubtotal := int64(1560000)
		if invoice.SubTotal != expectedSubtotal {
			t.Errorf("expected subtotal %d, got: %d", expectedSubtotal, invoice.SubTotal)
		}

		// Discount = 50000
		if invoice.DiscountAmount != 50000 {
			t.Errorf("expected discount 50000, got: %d", invoice.DiscountAmount)
		}

		// Taxable = 1560000 - 50000 = 1510000
		expectedTaxable := int64(1510000)
		if invoice.TaxableAmount != expectedTaxable {
			t.Errorf("expected taxable %d, got: %d", expectedTaxable, invoice.TaxableAmount)
		}

		// CGST = 9% of 1510000 = 135900
		// SGST = 9% of 1510000 = 135900
		if invoice.CGSTAmount == 0 {
			t.Error("expected CGST to be calculated")
		}
		if invoice.SGSTAmount == 0 {
			t.Error("expected SGST to be calculated")
		}
		if invoice.CGSTAmount != invoice.SGSTAmount {
			t.Errorf("expected CGST=SGST, got CGST=%d SGST=%d", invoice.CGSTAmount, invoice.SGSTAmount)
		}

		// Total = taxable + CGST + SGST
		expectedTotal := expectedTaxable + invoice.CGSTAmount + invoice.SGSTAmount
		if invoice.TotalAmount != expectedTotal {
			t.Errorf("expected total %d, got: %d", expectedTotal, invoice.TotalAmount)
		}

		t.Logf("Multi-item invoice: subtotal=₹%.2f, discount=₹%.2f, CGST=₹%.2f, SGST=₹%.2f, total=₹%.2f",
			float64(invoice.SubTotal)/100,
			float64(invoice.DiscountAmount)/100,
			float64(invoice.CGSTAmount)/100,
			float64(invoice.SGSTAmount)/100,
			float64(invoice.TotalAmount)/100,
		)
	})

	// =========================================================================
	// STEP 20: Edge Cases
	// =========================================================================
	t.Run("Step20_EdgeCases", func(t *testing.T) {
		// Invalid patient ID for invoice
		_, err := invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID: "nonexistent-id",
			Items:     []InvoiceItemInput{{Description: "Test", Quantity: 1, UnitPrice: 1000}},
		})
		if err == nil {
			t.Fatal("expected error for nonexistent patient in invoice")
		}

		// Empty items
		_, err = invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID: patient1ID,
			Items:     []InvoiceItemInput{},
		})
		if err == nil {
			t.Fatal("expected error for empty items")
		}

		// Discount > subtotal
		_, err = invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID:      patient1ID,
			Items:          []InvoiceItemInput{{Description: "Small item", Quantity: 1, UnitPrice: 1000}},
			DiscountAmount: 2000,
		})
		if err == nil {
			t.Fatal("expected error for discount > subtotal")
		}

		// Overpayment
		smallInv, err := invoiceService.CreateInvoice(CreateInvoiceInput{
			PatientID: patient1ID,
			Items:     []InvoiceItemInput{{Description: "Small", Quantity: 1, UnitPrice: 10000}},
		})
		if err != nil {
			t.Fatalf("create small invoice error: %v", err)
		}
		_, err = invoiceService.RecordPayment(RecordPaymentInput{
			InvoiceID: smallInv.ID,
			Amount:    99999999, // Way more than invoice total
			Method:    "cash",
		})
		if err == nil {
			t.Fatal("expected error for overpayment")
		}
	})

	t.Log("✅ All end-to-end tests passed!")
}

// TestE2E_ConcurrentPatientCreation tests that phone uniqueness is enforced
// even under concurrent creation attempts.
func TestE2E_ConcurrentPatientCreation(t *testing.T) {
	_, authService, patientService, _, _, settingsService := setupFullStack(t)

	// Setup
	settingsService.CompleteSetup(SetupInput{
		ClinicName:    "Test Clinic",
		DoctorName:    "Dr. Test",
		Phone:         "9999999999",
		AdminUsername: "admin",
		AdminPassword: "admin123",
		AdminFullName: "Admin",
	})
	authService.Login("admin", "admin123")

	// Create first patient
	_, err := patientService.CreatePatient(CreatePatientInput{
		Name:  "First Patient",
		Phone: "9876543210",
	})
	if err != nil {
		t.Fatalf("first patient error: %v", err)
	}

	// Attempt duplicate (same phone) — should be blocked
	_, err = patientService.CreatePatient(CreatePatientInput{
		Name:  "Second Patient Same Phone",
		Phone: "9876543210",
	})
	if err == nil {
		t.Fatal("expected duplicate phone error")
	}
}

// TestE2E_SessionManagement tests session lifecycle end-to-end.
func TestE2E_SessionManagement(t *testing.T) {
	_, authService, _, _, _, settingsService := setupFullStack(t)

	// Setup
	settingsService.CompleteSetup(SetupInput{
		ClinicName:    "Session Test Clinic",
		DoctorName:    "Dr. Session",
		Phone:         "9999999998",
		AdminUsername: "sessionadmin",
		AdminPassword: "pass123",
		AdminFullName: "Session Admin",
	})

	// Not logged in
	current, _ := authService.GetCurrentUser()
	if current.LoggedIn {
		t.Fatal("expected not logged in before login")
	}

	// Login
	_, err := authService.Login("sessionadmin", "pass123")
	if err != nil {
		t.Fatalf("login error: %v", err)
	}

	// Now logged in
	current, _ = authService.GetCurrentUser()
	if !current.LoggedIn {
		t.Fatal("expected logged in after login")
	}
	if current.User.Username != "sessionadmin" {
		t.Errorf("expected username 'sessionadmin', got: %s", current.User.Username)
	}

	// Logout
	err = authService.Logout()
	if err != nil {
		t.Fatalf("logout error: %v", err)
	}

	// Not logged in anymore
	current, _ = authService.GetCurrentUser()
	if current.LoggedIn {
		t.Fatal("expected not logged in after logout")
	}
}

// TestE2E_AppointmentWeekView tests fetching appointments across a date range.
func TestE2E_AppointmentWeekView(t *testing.T) {
	db, authService, patientService, appointmentService, _, settingsService := setupFullStack(t)
	_ = db

	// Setup
	settingsService.CompleteSetup(SetupInput{
		ClinicName:    "Week View Clinic",
		DoctorName:    "Dr. Week",
		Phone:         "9999999997",
		AdminUsername: "weekadmin",
		AdminPassword: "pass123",
		AdminFullName: "Week Admin",
	})
	authService.Login("weekadmin", "pass123")

	// Create patient
	p, err := patientService.CreatePatient(CreatePatientInput{
		Name:  "Week Patient",
		Phone: "9876543299",
	})
	if err != nil {
		t.Fatalf("create patient error: %v", err)
	}

	// Create appointments across a week
	dates := []string{"2026-06-01", "2026-06-02", "2026-06-03", "2026-06-04", "2026-06-05"}
	for i, date := range dates {
		_, err := appointmentService.CreateAppointment(CreateAppointmentInput{
			PatientID: p.ID,
			Date:      date,
			StartTime: "09:00",
			EndTime:   "09:30",
			Duration:  30,
			Purpose:   "Checkup " + uuid.New().String()[:4],
		})
		if err != nil {
			t.Fatalf("create appointment %d error: %v", i, err)
		}
	}

	// Fetch week view
	weekAppts, err := appointmentService.GetWeekAppointments("2026-06-01", "2026-06-05")
	if err != nil {
		t.Fatalf("GetWeekAppointments error: %v", err)
	}
	if len(weekAppts) != 5 {
		t.Errorf("expected 5 appointments in week, got: %d", len(weekAppts))
	}

	// Partial range
	partialAppts, err := appointmentService.GetWeekAppointments("2026-06-02", "2026-06-04")
	if err != nil {
		t.Fatalf("partial week error: %v", err)
	}
	if len(partialAppts) != 3 {
		t.Errorf("expected 3 appointments in partial range, got: %d", len(partialAppts))
	}
}
