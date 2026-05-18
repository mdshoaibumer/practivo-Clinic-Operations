package handler

import (
	"log/slog"

	"clinmitra/internal/models"
	"clinmitra/internal/service"
)

type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

// NewInvoiceHandler creates an InvoiceHandler backed by the given service.
func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// CreateInvoice generates a new invoice with line items, GST calculation,
// and treatment history recording.
func (h *InvoiceHandler) CreateInvoice(input service.CreateInvoiceInput) (*models.Invoice, error) {
	slog.Info("creating invoice", "patientId", input.PatientID, "items", len(input.Items))
	result, err := h.invoiceService.CreateInvoice(input)
	if err != nil {
		slog.Warn("create invoice failed", "patientId", input.PatientID, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("invoice created", "id", result.ID, "number", result.InvoiceNumber, "total", result.TotalAmount)
	return result, nil
}

// GetInvoice retrieves a single invoice by ID with items and payments.
func (h *InvoiceHandler) GetInvoice(id string) (*models.Invoice, error) {
	result, err := h.invoiceService.GetInvoice(id)
	if err != nil {
		slog.Debug("get invoice failed", "id", id, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// ListInvoices returns a paginated, filterable list of invoices.
func (h *InvoiceHandler) ListInvoices(page, pageSize int, status, startDate, endDate, patientID, search string) (*service.InvoiceListResponse, error) {
	page, pageSize = sanitizePagination(page, pageSize)
	search = sanitizeSearch(search)
	result, err := h.invoiceService.ListInvoices(page, pageSize, status, startDate, endDate, patientID, search)
	if err != nil {
		slog.Error("list invoices failed", "page", page, "status", status, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// RecordPayment records a payment against an invoice and updates its status.
func (h *InvoiceHandler) RecordPayment(input service.RecordPaymentInput) (*models.Payment, error) {
	slog.Info("recording payment", "invoiceId", input.InvoiceID, "amount", input.Amount, "method", input.Method)
	result, err := h.invoiceService.RecordPayment(input)
	if err != nil {
		slog.Warn("record payment failed", "invoiceId", input.InvoiceID, "amount", input.Amount, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("payment recorded", "paymentId", result.ID, "invoiceId", input.InvoiceID, "amount", input.Amount)
	return result, nil
}

// VoidInvoice marks an unpaid invoice as void with a reason.
func (h *InvoiceHandler) VoidInvoice(id, reason string) error {
	slog.Info("voiding invoice", "id", id, "reason", reason)
	err := h.invoiceService.VoidInvoice(id, reason)
	if err != nil {
		slog.Warn("void invoice failed", "id", id, "error", err.Error())
		return safeError(err)
	}
	slog.Info("invoice voided", "id", id)
	return nil
}

// GetPatientOutstanding returns the total unpaid balance for a patient.
func (h *InvoiceHandler) GetPatientOutstanding(patientID string) (int64, error) {
	result, err := h.invoiceService.GetPatientOutstanding(patientID)
	if err != nil {
		slog.Error("get patient outstanding failed", "patientId", patientID, "error", err.Error())
		return 0, safeError(err)
	}
	return result, nil
}

// GetPatientInvoices returns all invoices for a specific patient.
func (h *InvoiceHandler) GetPatientInvoices(patientID string) ([]models.Invoice, error) {
	result, err := h.invoiceService.GetPatientInvoices(patientID)
	if err != nil {
		slog.Error("get patient invoices failed", "patientId", patientID, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}
