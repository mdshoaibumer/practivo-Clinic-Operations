package repository

import (
	"errors"
	"fmt"

	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type invoiceRepo struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a GORM-backed InvoiceRepository implementation.
func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepo{db: db}
}

// Create persists a new Invoice record to the database.
func (r *invoiceRepo) Create(invoice *models.Invoice) error {
	return r.db.Create(invoice).Error
}

// FindByID retrieves an invoice by ID with Patient, Items, and Payments preloaded.
func (r *invoiceRepo) FindByID(id string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.Preload("Patient").Preload("Items").Preload("Payments").
		Where("id = ?", id).First(&invoice).Error
	if err != nil {
		return nil, WrapError(err)
	}
	return &invoice, nil
}

// Update saves changed fields on an existing invoice record.
func (r *invoiceRepo) Update(invoice *models.Invoice) error {
	return r.db.Model(invoice).Updates(invoice).Error
}

// List returns a paginated list of invoices filtered by status, date range,
// patient, and/or search term (invoice number). Patient relation is preloaded.
func (r *invoiceRepo) List(page, pageSize int, filters InvoiceFilters) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{})

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.StartDate != "" {
		query = query.Where("invoice_date >= ?", filters.StartDate)
	}
	if filters.EndDate != "" {
		query = query.Where("invoice_date <= ?", filters.EndDate)
	}
	if filters.PatientID != "" {
		query = query.Where("patient_id = ?", filters.PatientID)
	}
	if filters.Search != "" {
		query = query.Where("invoice_number LIKE ?", "%"+filters.Search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, WrapError(err)
	}

	offset := (page - 1) * pageSize
	err = query.Preload("Patient").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&invoices).Error
	if err != nil {
		return nil, 0, WrapError(err)
	}

	return invoices, total, nil
}

// ListByPatient returns all invoices for a patient with Items and Payments
// preloaded, ordered newest first.
func (r *invoiceRepo) ListByPatient(patientID string) ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.Preload("Items").Preload("Payments").
		Where("patient_id = ?", patientID).
		Order("created_at DESC").
		Find(&invoices).Error
	return invoices, WrapError(err)
}

// GetLastInvoiceNumber returns the latest invoice number matching the given
// prefix and year-month pattern (e.g. "PV-2601-%"). Used for sequential numbering.
func (r *invoiceRepo) GetLastInvoiceNumber(prefix, yearMonth string) (string, error) {
	var invoice models.Invoice
	pattern := fmt.Sprintf("%s-%s-%%", prefix, yearMonth)
	err := r.db.Where("invoice_number LIKE ?", pattern).
		Order("invoice_number DESC").
		First(&invoice).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", WrapError(err)
	}
	return invoice.InvoiceNumber, nil
}

// GetOutstandingByPatient returns the total unpaid balance (in paise) across
// all issued or partially-paid invoices for a given patient.
func (r *invoiceRepo) GetOutstandingByPatient(patientID string) (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.Model(&models.Invoice{}).
		Select("COALESCE(SUM(balance_amount), 0) as total").
		Where("patient_id = ? AND status IN (?, ?)", patientID, models.InvoiceIssued, models.InvoicePartial).
		Scan(&result).Error
	return result.Total, WrapError(err)
}

// GetTotalOutstanding returns the system-wide total unpaid balance (in paise)
// across all issued or partially-paid invoices.
func (r *invoiceRepo) GetTotalOutstanding() (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.Model(&models.Invoice{}).
		Select("COALESCE(SUM(balance_amount), 0) as total").
		Where("status IN (?, ?)", models.InvoiceIssued, models.InvoicePartial).
		Scan(&result).Error
	return result.Total, WrapError(err)
}

// GetRevenueByDateRange returns total payment amount (in paise) collected
// within the specified date range (inclusive).
func (r *invoiceRepo) GetRevenueByDateRange(startDate, endDate string) (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("payment_date >= ? AND payment_date <= ?", startDate, endDate).
		Scan(&result).Error
	return result.Total, WrapError(err)
}
