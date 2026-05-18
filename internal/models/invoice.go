package models

type InvoiceStatus string

const (
	InvoiceIssued  InvoiceStatus = "issued"
	InvoicePartial InvoiceStatus = "partial"
	InvoicePaid    InvoiceStatus = "paid"
	InvoiceVoid    InvoiceStatus = "void"
)

type Invoice struct {
	BaseModel
	InvoiceNumber   string        `gorm:"type:text;uniqueIndex;not null" json:"invoiceNumber"`
	PatientID       string        `gorm:"type:text;not null;index" json:"patientId"`
	InvoiceDate     string        `gorm:"type:text;not null;index" json:"invoiceDate"`  // YYYY-MM-DD
	SubTotal        int64         `gorm:"type:integer;not null" json:"subTotal"`        // paise
	DiscountAmount  int64         `gorm:"type:integer;default:0" json:"discountAmount"` // paise
	DiscountPercent float64       `gorm:"type:real;default:0" json:"discountPercent"`
	TaxableAmount   int64         `gorm:"type:integer;not null" json:"taxableAmount"`  // paise
	CGSTAmount      int64         `gorm:"type:integer;default:0" json:"cgstAmount"`    // paise
	SGSTAmount      int64         `gorm:"type:integer;default:0" json:"sgstAmount"`    // paise
	TotalAmount     int64         `gorm:"type:integer;not null" json:"totalAmount"`    // paise
	PaidAmount      int64         `gorm:"type:integer;default:0" json:"paidAmount"`    // paise
	BalanceAmount   int64         `gorm:"type:integer;default:0" json:"balanceAmount"` // paise
	Status          InvoiceStatus `gorm:"type:text;not null;default:'issued'" json:"status"`
	Notes           string        `gorm:"type:text" json:"notes"`
	VoidReason      string        `gorm:"type:text" json:"voidReason"`
	CreatedBy       string        `gorm:"type:text" json:"createdBy"`

	// Relationships
	Patient  Patient       `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Items    []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`
	Payments []Payment     `gorm:"foreignKey:InvoiceID" json:"payments,omitempty"`
}

type InvoiceItem struct {
	BaseModel
	InvoiceID   string  `gorm:"type:text;not null;index" json:"invoiceId"`
	TreatmentID *string `gorm:"type:text" json:"treatmentId"` // nullable for custom items
	Description string  `gorm:"type:text;not null" json:"description"`
	Quantity    int     `gorm:"type:integer;not null;default:1" json:"quantity"`
	UnitPrice   int64   `gorm:"type:integer;not null" json:"unitPrice"` // paise
	Amount      int64   `gorm:"type:integer;not null" json:"amount"`    // paise (quantity * unitPrice)
	ToothNumber string  `gorm:"type:text" json:"toothNumber"`

	// Relationships
	Treatment Treatment `gorm:"foreignKey:TreatmentID" json:"treatment,omitempty"`
}
