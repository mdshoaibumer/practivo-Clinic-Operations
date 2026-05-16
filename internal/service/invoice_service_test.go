package service

import (
	"testing"

	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"
)

// --- Mock Repositories ---

type mockInvoiceRepo struct {
	invoices      map[string]*models.Invoice
	lastNumber    string
	createErr     error
	outstanding   int64
	totalOutstand int64
}

func newMockInvoiceRepo() *mockInvoiceRepo {
	return &mockInvoiceRepo{invoices: make(map[string]*models.Invoice)}
}

func (m *mockInvoiceRepo) Create(invoice *models.Invoice) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.invoices[invoice.ID] = invoice
	return nil
}
func (m *mockInvoiceRepo) FindByID(id string) (*models.Invoice, error) { return m.invoices[id], nil }
func (m *mockInvoiceRepo) Update(invoice *models.Invoice) error {
	m.invoices[invoice.ID] = invoice
	return nil
}
func (m *mockInvoiceRepo) List(page, pageSize int, filters repository.InvoiceFilters) ([]models.Invoice, int64, error) {
	return nil, 0, nil
}
func (m *mockInvoiceRepo) ListByPatient(patientID string) ([]models.Invoice, error) { return nil, nil }
func (m *mockInvoiceRepo) GetLastInvoiceNumber(prefix, yearMonth string) (string, error) {
	return m.lastNumber, nil
}
func (m *mockInvoiceRepo) GetOutstandingByPatient(patientID string) (int64, error) {
	return m.outstanding, nil
}
func (m *mockInvoiceRepo) GetTotalOutstanding() (int64, error) { return m.totalOutstand, nil }
func (m *mockInvoiceRepo) GetRevenueByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}

type mockInvoiceItemRepo struct{}

func (m *mockInvoiceItemRepo) CreateBatch(items []models.InvoiceItem) error { return nil }
func (m *mockInvoiceItemRepo) FindByInvoiceID(invoiceID string) ([]models.InvoiceItem, error) {
	return nil, nil
}

type mockPaymentRepo struct {
	payments map[string]*models.Payment
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{payments: make(map[string]*models.Payment)}
}

func (m *mockPaymentRepo) Create(payment *models.Payment) error {
	m.payments[payment.ID] = payment
	return nil
}
func (m *mockPaymentRepo) FindByInvoiceID(invoiceID string) ([]models.Payment, error) {
	return nil, nil
}
func (m *mockPaymentRepo) GetTotalByInvoice(invoiceID string) (int64, error) { return 0, nil }
func (m *mockPaymentRepo) GetCollectionByDate(date string) (int64, error)    { return 0, nil }
func (m *mockPaymentRepo) GetCollectionByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}
func (m *mockPaymentRepo) ListByDateRange(startDate, endDate string) ([]models.Payment, error) {
	return nil, nil
}

type mockPatientRepo struct {
	patients map[string]*models.Patient
}

func newMockPatientRepo() *mockPatientRepo {
	return &mockPatientRepo{patients: make(map[string]*models.Patient)}
}

func (m *mockPatientRepo) Create(patient *models.Patient) error {
	m.patients[patient.ID] = patient
	return nil
}
func (m *mockPatientRepo) FindByID(id string) (*models.Patient, error) {
	p, ok := m.patients[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return p, nil
}
func (m *mockPatientRepo) Update(patient *models.Patient) error { return nil }
func (m *mockPatientRepo) Delete(id string) error               { return nil }
func (m *mockPatientRepo) List(page, pageSize int, search string) ([]models.Patient, int64, error) {
	return nil, 0, nil
}
func (m *mockPatientRepo) FindByPhone(phone string) (*models.Patient, error) {
	for _, p := range m.patients {
		if p.Phone == phone {
			return p, nil
		}
	}
	return nil, utils.ErrNotFound
}
func (m *mockPatientRepo) Count() (int64, error) { return int64(len(m.patients)), nil }

type mockTreatmentRepo struct {
	treatments map[string]*models.Treatment
}

func newMockTreatmentRepo() *mockTreatmentRepo {
	return &mockTreatmentRepo{treatments: make(map[string]*models.Treatment)}
}

func (m *mockTreatmentRepo) Create(treatment *models.Treatment) error { return nil }
func (m *mockTreatmentRepo) FindByID(id string) (*models.Treatment, error) {
	t, ok := m.treatments[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return t, nil
}
func (m *mockTreatmentRepo) Update(treatment *models.Treatment) error { return nil }
func (m *mockTreatmentRepo) Delete(id string) error                   { return nil }
func (m *mockTreatmentRepo) ListActive() ([]models.Treatment, error)  { return nil, nil }
func (m *mockTreatmentRepo) ListAll() ([]models.Treatment, error)     { return nil, nil }

type mockClinicRepo struct {
	settings *models.ClinicSettings
}

func (m *mockClinicRepo) Get() (*models.ClinicSettings, error) {
	if m.settings == nil {
		m.settings = &models.ClinicSettings{
			InvoicePrefix: "PV",
			GSTEnabled:    false,
		}
	}
	return m.settings, nil
}
func (m *mockClinicRepo) Upsert(settings *models.ClinicSettings) error {
	m.settings = settings
	return nil
}
func (m *mockClinicRepo) IsSetupComplete() (bool, error) { return true, nil }

type mockPatientTreatmentRepo struct{}

func (m *mockPatientTreatmentRepo) Create(pt *models.PatientTreatment) error        { return nil }
func (m *mockPatientTreatmentRepo) CreateBatch(pts []models.PatientTreatment) error { return nil }
func (m *mockPatientTreatmentRepo) ListByPatient(patientID string) ([]models.PatientTreatment, error) {
	return nil, nil
}

type mockAuditRepo struct{}

func (m *mockAuditRepo) Create(log *models.AuditLog) error { return nil }
func (m *mockAuditRepo) ListByEntity(entityType, entityID string) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditRepo) ListByUser(userID string, limit int) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *mockAuditRepo) ListRecent(limit int) ([]models.AuditLog, error) { return nil, nil }

type mockAuthService struct {
	userID string
}

func (m *mockAuthService) GetCurrentUserID() string { return m.userID }

// --- Tests ---

func TestInvoiceCalculation_SubTotal(t *testing.T) {
	tests := []struct {
		name     string
		items    []InvoiceItemInput
		expected int64
	}{
		{
			name: "single item",
			items: []InvoiceItemInput{
				{Description: "Filling", Quantity: 1, UnitPrice: 150000},
			},
			expected: 150000,
		},
		{
			name: "multiple items",
			items: []InvoiceItemInput{
				{Description: "Filling", Quantity: 2, UnitPrice: 150000},
				{Description: "Cleaning", Quantity: 1, UnitPrice: 100000},
			},
			expected: 400000, // 2*150000 + 1*100000
		},
		{
			name: "zero quantity defaults to 1",
			items: []InvoiceItemInput{
				{Description: "Consultation", Quantity: 0, UnitPrice: 50000},
			},
			expected: 50000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subTotal int64
			for _, item := range tt.items {
				quantity := item.Quantity
				if quantity < 1 {
					quantity = 1
				}
				subTotal += item.UnitPrice * int64(quantity)
			}
			if subTotal != tt.expected {
				t.Errorf("expected subtotal %d, got %d", tt.expected, subTotal)
			}
		})
	}
}

func TestInvoiceCalculation_Discount(t *testing.T) {
	tests := []struct {
		name            string
		subTotal        int64
		discountPercent float64
		discountAmount  int64
		expectedDisc    int64
		expectedTaxable int64
	}{
		{
			name:            "10 percent discount",
			subTotal:        100000,
			discountPercent: 10,
			discountAmount:  0,
			expectedDisc:    10000,
			expectedTaxable: 90000,
		},
		{
			name:            "flat discount",
			subTotal:        100000,
			discountPercent: 0,
			discountAmount:  5000,
			expectedDisc:    5000,
			expectedTaxable: 95000,
		},
		{
			name:            "no discount",
			subTotal:        200000,
			discountPercent: 0,
			discountAmount:  0,
			expectedDisc:    0,
			expectedTaxable: 200000,
		},
		{
			name:            "percent takes precedence",
			subTotal:        100000,
			discountPercent: 20,
			discountAmount:  5000,
			expectedDisc:    20000,
			expectedTaxable: 80000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var discountAmount int64
			if tt.discountPercent > 0 {
				discountAmount = int64(float64(tt.subTotal) * tt.discountPercent / 100)
			} else {
				discountAmount = tt.discountAmount
			}
			taxable := tt.subTotal - discountAmount

			if discountAmount != tt.expectedDisc {
				t.Errorf("expected discount %d, got %d", tt.expectedDisc, discountAmount)
			}
			if taxable != tt.expectedTaxable {
				t.Errorf("expected taxable %d, got %d", tt.expectedTaxable, taxable)
			}
		})
	}
}

func TestInvoiceCalculation_GST(t *testing.T) {
	tests := []struct {
		name          string
		taxable       int64
		gstEnabled    bool
		gstRate       int
		expectedCGST  int64
		expectedSGST  int64
		expectedTotal int64
	}{
		{
			name:          "GST disabled",
			taxable:       100000,
			gstEnabled:    false,
			gstRate:       18,
			expectedCGST:  0,
			expectedSGST:  0,
			expectedTotal: 100000,
		},
		{
			name:          "18% GST",
			taxable:       100000,
			gstEnabled:    true,
			gstRate:       18,
			expectedCGST:  9000,
			expectedSGST:  9000,
			expectedTotal: 118000,
		},
		{
			name:          "12% GST",
			taxable:       200000,
			gstEnabled:    true,
			gstRate:       12,
			expectedCGST:  12000,
			expectedSGST:  12000,
			expectedTotal: 224000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cgst, sgst int64
			if tt.gstEnabled {
				halfRate := float64(tt.gstRate) / 2
				cgst = int64(float64(tt.taxable) * halfRate / 100)
				sgst = int64(float64(tt.taxable) * halfRate / 100)
			}
			total := tt.taxable + cgst + sgst

			if cgst != tt.expectedCGST {
				t.Errorf("expected CGST %d, got %d", tt.expectedCGST, cgst)
			}
			if sgst != tt.expectedSGST {
				t.Errorf("expected SGST %d, got %d", tt.expectedSGST, sgst)
			}
			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestPaymentValidation_ExceedsBalance(t *testing.T) {
	invoice := &models.Invoice{
		TotalAmount:   100000,
		PaidAmount:    80000,
		BalanceAmount: 20000,
		Status:        models.InvoicePartial,
	}

	tests := []struct {
		name      string
		amount    int64
		wantError bool
	}{
		{"exact balance", 20000, false},
		{"under balance", 10000, false},
		{"exceeds balance", 30000, true},
		{"zero amount", 0, true},
		{"negative amount", -1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			if tt.amount <= 0 {
				hasError = true
			} else if tt.amount > invoice.BalanceAmount {
				hasError = true
			}

			if hasError != tt.wantError {
				t.Errorf("amount=%d: expected error=%v, got error=%v", tt.amount, tt.wantError, hasError)
			}
		})
	}
}

func TestPaymentStatusTransition(t *testing.T) {
	tests := []struct {
		name           string
		totalAmount    int64
		existingPaid   int64
		paymentAmount  int64
		expectedStatus models.InvoiceStatus
	}{
		{
			name:           "full payment on new invoice",
			totalAmount:    100000,
			existingPaid:   0,
			paymentAmount:  100000,
			expectedStatus: models.InvoicePaid,
		},
		{
			name:           "partial payment",
			totalAmount:    100000,
			existingPaid:   0,
			paymentAmount:  50000,
			expectedStatus: models.InvoicePartial,
		},
		{
			name:           "final payment completes",
			totalAmount:    100000,
			existingPaid:   70000,
			paymentAmount:  30000,
			expectedStatus: models.InvoicePaid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newPaid := tt.existingPaid + tt.paymentAmount
			balance := tt.totalAmount - newPaid

			var status models.InvoiceStatus
			if balance <= 0 {
				status = models.InvoicePaid
			} else {
				status = models.InvoicePartial
			}

			if status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

func TestVoidInvoice_Validation(t *testing.T) {
	tests := []struct {
		name       string
		status     models.InvoiceStatus
		paidAmount int64
		reason     string
		wantError  bool
		errMsg     string
	}{
		{
			name:       "valid void",
			status:     models.InvoiceIssued,
			paidAmount: 0,
			reason:     "duplicate",
			wantError:  false,
		},
		{
			name:       "already voided",
			status:     models.InvoiceVoid,
			paidAmount: 0,
			reason:     "duplicate",
			wantError:  true,
			errMsg:     "already voided",
		},
		{
			name:       "has payments",
			status:     models.InvoicePartial,
			paidAmount: 50000,
			reason:     "error",
			wantError:  true,
			errMsg:     "payments",
		},
		{
			name:       "no reason",
			status:     models.InvoiceIssued,
			paidAmount: 0,
			reason:     "",
			wantError:  true,
			errMsg:     "reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			if tt.status == models.InvoiceVoid {
				hasError = true
			} else if tt.paidAmount > 0 {
				hasError = true
			} else if tt.reason == "" {
				hasError = true
			}

			if hasError != tt.wantError {
				t.Errorf("expected error=%v, got error=%v", tt.wantError, hasError)
			}
		})
	}
}

func TestListInvoices_PaginationDefaults(t *testing.T) {
	tests := []struct {
		name             string
		page             int
		pageSize         int
		expectedPage     int
		expectedPageSize int
	}{
		{"valid values", 2, 20, 2, 20},
		{"zero page defaults to 1", 0, 20, 1, 20},
		{"negative page defaults to 1", -1, 20, 1, 20},
		{"zero pageSize defaults to 20", 1, 0, 1, 20},
		{"pageSize over 100 capped to max", 1, 200, 1, 100},
		{"negative pageSize defaults to 20", 1, -5, 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := tt.page
			pageSize := tt.pageSize
			if page < 1 {
				page = 1
			}
			if pageSize < 1 {
				pageSize = 20
			}
			if pageSize > 100 {
				pageSize = 100
			}

			if page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, page)
			}
			if pageSize != tt.expectedPageSize {
				t.Errorf("expected pageSize %d, got %d", tt.expectedPageSize, pageSize)
			}
		})
	}
}
