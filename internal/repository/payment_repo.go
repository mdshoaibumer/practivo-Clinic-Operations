package repository

import (
	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type paymentRepo struct {
	db *gorm.DB
}

// NewPaymentRepository creates a GORM-backed PaymentRepository implementation.
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

// Create persists a new Payment record to the database.
func (r *paymentRepo) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// FindByInvoiceID returns all payments for a given invoice, newest first.
func (r *paymentRepo) FindByInvoiceID(invoiceID string) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("invoice_id = ?", invoiceID).Order("payment_date DESC").Find(&payments).Error
	return payments, WrapError(err)
}

// GetTotalByInvoice returns the sum of all payments (in paise) for an invoice.
func (r *paymentRepo) GetTotalByInvoice(invoiceID string) (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("invoice_id = ?", invoiceID).
		Scan(&result).Error
	return result.Total, WrapError(err)
}

// GetCollectionByDate returns the total payment amount (in paise) for a single date.
func (r *paymentRepo) GetCollectionByDate(date string) (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("payment_date = ?", date).
		Scan(&result).Error
	return result.Total, WrapError(err)
}

// GetCollectionByDateRange returns the total payment amount (in paise) within
// a date range (inclusive).
func (r *paymentRepo) GetCollectionByDateRange(startDate, endDate string) (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("payment_date >= ? AND payment_date <= ?", startDate, endDate).
		Scan(&result).Error
	return result.Total, WrapError(err)
}

// ListByDateRange returns all payments within a date range with Invoice and
// Invoice.Patient preloaded, ordered newest first.
func (r *paymentRepo) ListByDateRange(startDate, endDate string) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Preload("Invoice").Preload("Invoice.Patient").
		Where("payment_date >= ? AND payment_date <= ?", startDate, endDate).
		Order("payment_date DESC").
		Find(&payments).Error
	return payments, WrapError(err)
}
