package service

import (
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"

	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceItemInput struct {
	TreatmentID string `json:"treatmentId"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unitPrice"` // paise
	ToothNumber string `json:"toothNumber"`
}

type CreateInvoiceInput struct {
	PatientID       string             `json:"patientId"`
	Items           []InvoiceItemInput `json:"items"`
	DiscountPercent float64            `json:"discountPercent"`
	DiscountAmount  int64              `json:"discountAmount"` // paise, manual discount
	Notes           string             `json:"notes"`
}

type RecordPaymentInput struct {
	InvoiceID   string `json:"invoiceId"`
	Amount      int64  `json:"amount"` // paise
	Method      string `json:"method"`
	PaymentDate string `json:"paymentDate"`
	Reference   string `json:"reference"`
	Notes       string `json:"notes"`
}

type InvoiceListResponse struct {
	Invoices []models.Invoice `json:"invoices"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type InvoiceService struct {
	invoiceRepo          repository.InvoiceRepository
	invoiceItemRepo      repository.InvoiceItemRepository
	paymentRepo          repository.PaymentRepository
	patientRepo          repository.PatientRepository
	treatmentRepo        repository.TreatmentRepository
	clinicRepo           repository.ClinicRepository
	patientTreatmentRepo repository.PatientTreatmentRepository
	authService          *AuthService
	auditService         *AuditService
	db                   *gorm.DB
}

// NewInvoiceService creates an InvoiceService with all required dependencies
// for invoice creation, payment recording, and voiding.
func NewInvoiceService(
	invoiceRepo repository.InvoiceRepository,
	invoiceItemRepo repository.InvoiceItemRepository,
	paymentRepo repository.PaymentRepository,
	patientRepo repository.PatientRepository,
	treatmentRepo repository.TreatmentRepository,
	clinicRepo repository.ClinicRepository,
	patientTreatmentRepo repository.PatientTreatmentRepository,
	authService *AuthService,
	auditService *AuditService,
	db *gorm.DB,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:          invoiceRepo,
		invoiceItemRepo:      invoiceItemRepo,
		paymentRepo:          paymentRepo,
		patientRepo:          patientRepo,
		treatmentRepo:        treatmentRepo,
		clinicRepo:           clinicRepo,
		patientTreatmentRepo: patientTreatmentRepo,
		authService:          authService,
		auditService:         auditService,
		db:                   db,
	}
}

// CreateInvoice generates a new invoice with line items. Calculates subtotal,
// discount, GST (if enabled), and total. Creates patient treatment records
// for each item with a TreatmentID. All operations run inside a transaction.
func (s *InvoiceService) CreateInvoice(input CreateInvoiceInput) (*models.Invoice, error) {
	if err := utils.ValidateRequired("Patient", input.PatientID); err != nil {
		return nil, err
	}
	if len(input.Items) == 0 {
		return nil, utils.ValidationError("At least one item is required")
	}

	// Verify patient
	if _, err := s.patientRepo.FindByID(input.PatientID); err != nil {
		return nil, utils.ValidationError("Patient not found")
	}

	// Get clinic settings for GST and prefix
	settings, err := s.clinicRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get clinic settings: %w", err)
	}

	// Calculate totals
	var subTotal int64
	invoiceItems := make([]models.InvoiceItem, len(input.Items))
	patientTreatments := make([]models.PatientTreatment, 0)

	for i, item := range input.Items {
		quantity := item.Quantity
		if quantity < 1 {
			quantity = 1
		}
		amount := item.UnitPrice * int64(quantity)
		subTotal += amount

		invItem := models.InvoiceItem{
			BaseModel:   models.BaseModel{ID: uuid.New().String()},
			Description: item.Description,
			Quantity:    quantity,
			UnitPrice:   item.UnitPrice,
			Amount:      amount,
			ToothNumber: item.ToothNumber,
		}
		if item.TreatmentID != "" {
			invItem.TreatmentID = &item.TreatmentID
		}
		invoiceItems[i] = invItem

		// Build patient treatment history
		if item.TreatmentID != "" {
			patientTreatments = append(patientTreatments, models.PatientTreatment{
				BaseModel:     models.BaseModel{ID: uuid.New().String()},
				PatientID:     input.PatientID,
				TreatmentID:   item.TreatmentID,
				TreatmentDate: utils.TodayDate(),
				ToothNumber:   item.ToothNumber,
				PerformedBy:   s.authService.GetCurrentUserID(),
			})
		}
	}

	// Calculate discount
	var discountAmount int64
	if input.DiscountPercent > 0 {
		discountAmount = int64(math.Round(float64(subTotal) * input.DiscountPercent / 100))
	} else {
		discountAmount = input.DiscountAmount
	}
	if discountAmount > subTotal {
		return nil, utils.ValidationError("Discount cannot exceed subtotal")
	}

	taxableAmount := subTotal - discountAmount

	// Calculate GST
	var cgst, sgst int64
	if settings.GSTEnabled {
		halfRate := float64(settings.GSTRate) / 2
		cgst = int64(math.Round(float64(taxableAmount) * halfRate / 100))
		sgst = int64(math.Round(float64(taxableAmount) * halfRate / 100))
	}

	totalAmount := taxableAmount + cgst + sgst

	var invoice *models.Invoice

	// Use transaction for atomicity — invoice number generation is inside
	// the transaction to prevent duplicate numbers under concurrent access.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Generate invoice number inside transaction for race safety
		invoiceNumber, err := s.generateInvoiceNumber(settings.InvoicePrefix)
		if err != nil {
			return fmt.Errorf("failed to generate invoice number: %w", err)
		}

		invoice = &models.Invoice{
			BaseModel:       models.BaseModel{ID: uuid.New().String()},
			InvoiceNumber:   invoiceNumber,
			PatientID:       input.PatientID,
			InvoiceDate:     utils.TodayDate(),
			SubTotal:        subTotal,
			DiscountAmount:  discountAmount,
			DiscountPercent: input.DiscountPercent,
			TaxableAmount:   taxableAmount,
			CGSTAmount:      cgst,
			SGSTAmount:      sgst,
			TotalAmount:     totalAmount,
			PaidAmount:      0,
			BalanceAmount:   totalAmount,
			Status:          models.InvoiceIssued,
			Notes:           input.Notes,
			CreatedBy:       s.authService.GetCurrentUserID(),
		}

		if err := tx.Create(invoice).Error; err != nil {
			return err
		}

		for i := range invoiceItems {
			invoiceItems[i].InvoiceID = invoice.ID
		}
		if err := tx.CreateInBatches(invoiceItems, 10).Error; err != nil {
			return err
		}

		// Record patient treatments
		for i := range patientTreatments {
			patientTreatments[i].InvoiceID = invoice.ID
		}
		if len(patientTreatments) > 0 {
			if err := tx.CreateInBatches(patientTreatments, 10).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("invoice creation transaction failed",
			"patientId", input.PatientID,
			"items", len(input.Items),
			"subTotal", subTotal,
			"error", err.Error(),
		)
		return nil, err
	}

	invoice.Items = invoiceItems
	slog.Info("invoice created in service",
		"invoiceId", invoice.ID,
		"number", invoice.InvoiceNumber,
		"patientId", input.PatientID,
		"total", invoice.TotalAmount,
		"items", len(invoiceItems),
	)
	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditCreate, "invoice", invoice.ID, nil, invoice)
	return invoice, nil
}

// GetInvoice retrieves a single invoice by ID with patient, items, and payments.
func (s *InvoiceService) GetInvoice(id string) (*models.Invoice, error) {
	invoice, err := s.invoiceRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

// ListInvoices returns a paginated, filterable list of invoices.
// Filters: status, date range, patient ID, and search (invoice number).
func (s *InvoiceService) ListInvoices(page, pageSize int, status, startDate, endDate, patientID, search string) (*InvoiceListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filters := repository.InvoiceFilters{
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
		PatientID: patientID,
		Search:    search,
	}

	invoices, total, err := s.invoiceRepo.List(page, pageSize, filters)
	if err != nil {
		return nil, err
	}

	return &InvoiceListResponse{
		Invoices: invoices,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// RecordPayment records a payment against an invoice. Updates the invoice's
// paid/balance amounts and transitions status (issued -> partial -> paid).
// The balance check and update are performed inside a single transaction to
// prevent overpayment from concurrent requests.
func (s *InvoiceService) RecordPayment(input RecordPaymentInput) (*models.Payment, error) {
	if err := utils.ValidateRequired("Invoice", input.InvoiceID); err != nil {
		return nil, err
	}
	if err := utils.ValidatePositiveAmount("Payment amount", input.Amount); err != nil {
		return nil, err
	}

	paymentDate := input.PaymentDate
	if paymentDate == "" {
		paymentDate = utils.TodayDate()
	}

	payment := &models.Payment{
		BaseModel:   models.BaseModel{ID: uuid.New().String()},
		InvoiceID:   input.InvoiceID,
		Amount:      input.Amount,
		Method:      models.PaymentMethod(input.Method),
		PaymentDate: paymentDate,
		Reference:   input.Reference,
		Notes:       input.Notes,
		ReceivedBy:  s.authService.GetCurrentUserID(),
	}

	var invoice *models.Invoice

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Re-read invoice inside transaction to get accurate balance
		var inv models.Invoice
		if err := tx.Preload("Patient").Preload("Items").Preload("Payments").
			Where("id = ?", input.InvoiceID).First(&inv).Error; err != nil {
			return utils.ErrNotFound
		}
		invoice = &inv

		if invoice.Status == models.InvoiceVoid {
			return utils.ValidationError("Cannot record payment for a voided invoice")
		}
		if invoice.Status == models.InvoicePaid {
			return utils.ValidationError("Invoice is already fully paid")
		}
		if input.Amount > invoice.BalanceAmount {
			return utils.ErrInsufficientBalance
		}

		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		invoice.PaidAmount += input.Amount
		invoice.BalanceAmount = invoice.TotalAmount - invoice.PaidAmount

		if invoice.BalanceAmount <= 0 {
			invoice.Status = models.InvoicePaid
			invoice.BalanceAmount = 0
		} else {
			invoice.Status = models.InvoicePartial
		}

		return tx.Model(invoice).Updates(map[string]interface{}{
			"paid_amount":    invoice.PaidAmount,
			"balance_amount": invoice.BalanceAmount,
			"status":         invoice.Status,
		}).Error
	})

	if err != nil {
		slog.Error("payment recording transaction failed",
			"invoiceId", input.InvoiceID,
			"amount", input.Amount,
			"method", input.Method,
			"error", err.Error(),
		)
		return nil, err
	}

	slog.Info("payment recorded in service",
		"paymentId", payment.ID,
		"invoiceId", input.InvoiceID,
		"amount", input.Amount,
		"newStatus", invoice.Status,
		"balance", invoice.BalanceAmount,
	)
	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditCreate, "payment", payment.ID, nil, payment)
	return payment, nil
}

// VoidInvoice marks an invoice as void with a reason. Only unpaid invoices
// (no recorded payments) can be voided. Requires admin role.
func (s *InvoiceService) VoidInvoice(id, reason string) error {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return err
	}

	invoice, err := s.invoiceRepo.FindByID(id)
	if err != nil {
		return err
	}

	if invoice.Status == models.InvoiceVoid {
		return utils.ValidationError("Invoice is already voided")
	}
	if invoice.PaidAmount > 0 {
		return utils.ValidationError("Cannot void an invoice with recorded payments. Refund first.")
	}
	if reason == "" {
		return utils.ValidationError("Void reason is required")
	}

	invoice.Status = models.InvoiceVoid
	invoice.VoidReason = reason
	invoice.BalanceAmount = 0

	err = s.db.Model(invoice).Updates(map[string]interface{}{
		"status":         invoice.Status,
		"void_reason":    invoice.VoidReason,
		"balance_amount": invoice.BalanceAmount,
	}).Error
	if err != nil {
		return err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "invoice", id, nil, map[string]string{
		"action": "void",
		"reason": reason,
	})
	return nil
}

// GetPatientOutstanding returns the total unpaid balance across all invoices
// for a specific patient.
func (s *InvoiceService) GetPatientOutstanding(patientID string) (int64, error) {
	return s.invoiceRepo.GetOutstandingByPatient(patientID)
}

// GetPatientInvoices returns all invoices for a specific patient.
func (s *InvoiceService) GetPatientInvoices(patientID string) ([]models.Invoice, error) {
	return s.invoiceRepo.ListByPatient(patientID)
}

// generateInvoiceNumber creates a sequential invoice number in the format
// PREFIX-YYMM-NNNN (e.g., PV-2601-0001). Queries the last used number
// and increments the sequence.
func (s *InvoiceService) generateInvoiceNumber(prefix string) (string, error) {
	yearMonth := utils.CurrentMonth()
	lastNumber, err := s.invoiceRepo.GetLastInvoiceNumber(prefix, yearMonth)
	if err != nil {
		return "", err
	}

	var seq int
	if lastNumber == "" {
		seq = 1
	} else {
		parts := strings.Split(lastNumber, "-")
		if len(parts) == 3 {
			seq, _ = strconv.Atoi(parts[2])
			seq++
		} else {
			seq = 1
		}
	}

	return fmt.Sprintf("%s-%s-%04d", prefix, yearMonth, seq), nil
}
