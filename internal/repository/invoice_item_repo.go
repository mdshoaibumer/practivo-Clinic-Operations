package repository

import (
	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type invoiceItemRepo struct {
	db *gorm.DB
}

// NewInvoiceItemRepository creates a GORM-backed InvoiceItemRepository implementation.
func NewInvoiceItemRepository(db *gorm.DB) InvoiceItemRepository {
	return &invoiceItemRepo{db: db}
}

// CreateBatch inserts multiple InvoiceItems in batches of 10.
func (r *invoiceItemRepo) CreateBatch(items []models.InvoiceItem) error {
	return r.db.CreateInBatches(items, 10).Error
}

// FindByInvoiceID returns all line items for an invoice with Treatment preloaded.
func (r *invoiceItemRepo) FindByInvoiceID(invoiceID string) ([]models.InvoiceItem, error) {
	var items []models.InvoiceItem
	err := r.db.Preload("Treatment").Where("invoice_id = ?", invoiceID).Find(&items).Error
	return items, WrapError(err)
}
