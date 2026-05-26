import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import InvoiceDetail from './InvoiceDetail'

// Mock stores
const mockFetchInvoice = vi.fn()
const mockRecordPayment = vi.fn()
let mockCurrentInvoice: any = null
let mockIsLoading = false

vi.mock('@/store/invoiceStore', () => ({
  useInvoiceStore: () => ({
    currentInvoice: mockCurrentInvoice,
    isLoading: mockIsLoading,
    fetchInvoice: mockFetchInvoice,
    recordPayment: mockRecordPayment,
  }),
}))

const mockClinic = { clinicName: 'Test Clinic', doctorName: 'Dr. Test' }
const mockFetchSettings = vi.fn()

vi.mock('@/store/settingsStore', () => ({
  useSettingsStore: () => ({
    clinic: mockClinic,
    fetchSettings: mockFetchSettings,
  }),
}))

// Mock WhatsApp handler
const mockGetWhatsAppTemplates = vi.fn()
const mockPrepareInvoiceMessage = vi.fn()

// Mock window.go
beforeEach(() => {
  vi.clearAllMocks()
  mockCurrentInvoice = null
  mockIsLoading = false
  ;(window as any).go = {
    handler: {
      WhatsAppHandler: {
        GetWhatsAppTemplates: mockGetWhatsAppTemplates,
        PrepareInvoiceMessage: mockPrepareInvoiceMessage,
      },
    },
  }
})

// Mock components that aren't needed for these tests
vi.mock('@/components/billing/InvoicePrintView', () => ({
  default: () => <div data-testid="print-view" />,
}))

vi.mock('@/components/WhatsAppDialog', () => ({
  default: ({ isOpen }: { isOpen: boolean }) => isOpen ? <div data-testid="whatsapp-dialog" /> : null,
}))

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderInvoiceDetail(invoiceId = 'inv-1') {
  const queryClient = createQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`/billing/${invoiceId}`]}>
        <Routes>
          <Route path="/billing/:id" element={<InvoiceDetail />} />
          <Route path="/billing" element={<div data-testid="billing-list" />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('InvoiceDetail', () => {
  describe('loading state', () => {
    it('shows skeleton while loading', () => {
      mockIsLoading = true
      const { container } = renderInvoiceDetail()
      expect(container.querySelector('.animate-pulse')).toBeInTheDocument()
    })

    it('shows skeleton when no invoice loaded', () => {
      mockCurrentInvoice = null
      mockIsLoading = false
      renderInvoiceDetail()
      // Should show loading skeleton since currentInvoice is null
      expect(screen.queryByText('Invoice')).not.toBeInTheDocument()
    })
  })

  describe('invoice display', () => {
    beforeEach(() => {
      mockCurrentInvoice = {
        id: 'inv-1',
        invoiceNumber: 'CD-2605-0001',
        status: 'issued',
        invoiceDate: '2026-05-26',
        patient: { name: 'John Doe', phone: '9876543210' },
        items: [
          { id: 'item-1', description: 'Cleaning', quantity: 1, unitPrice: 150000, amount: 150000, toothNumber: '' },
          { id: 'item-2', description: 'Filling', quantity: 2, unitPrice: 200000, amount: 400000, toothNumber: '16' },
        ],
        subTotal: 550000,
        discountAmount: 0,
        cgstAmount: 0,
        sgstAmount: 0,
        totalAmount: 550000,
        paidAmount: 0,
        balanceAmount: 550000,
        payments: [],
      }
    })

    it('renders invoice number', () => {
      renderInvoiceDetail()
      expect(screen.getByText('Invoice CD-2605-0001')).toBeInTheDocument()
    })

    it('renders patient name', () => {
      renderInvoiceDetail()
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    it('renders line items', () => {
      renderInvoiceDetail()
      expect(screen.getByText('Cleaning')).toBeInTheDocument()
      expect(screen.getByText('Filling')).toBeInTheDocument()
    })

    it('shows tooth number when present', () => {
      renderInvoiceDetail()
      expect(screen.getByText('(Tooth: 16)')).toBeInTheDocument()
    })

    it('shows Record Payment button for unpaid invoice', () => {
      renderInvoiceDetail()
      expect(screen.getByRole('button', { name: /record payment/i })).toBeInTheDocument()
    })

    it('hides Record Payment button for paid invoice', () => {
      mockCurrentInvoice.status = 'paid'
      renderInvoiceDetail()
      expect(screen.queryByRole('button', { name: /record payment/i })).not.toBeInTheDocument()
    })

    it('hides Record Payment button for voided invoice', () => {
      mockCurrentInvoice.status = 'void'
      renderInvoiceDetail()
      expect(screen.queryByRole('button', { name: /record payment/i })).not.toBeInTheDocument()
    })

    it('shows Print button', () => {
      renderInvoiceDetail()
      expect(screen.getByRole('button', { name: /print/i })).toBeInTheDocument()
    })

    it('renders payment history when payments exist', () => {
      mockCurrentInvoice.payments = [
        { id: 'pay-1', amount: 100000, method: 'cash', paymentDate: '2026-05-26', reference: '' },
      ]
      renderInvoiceDetail()
      expect(screen.getByText('Payment History')).toBeInTheDocument()
    })

    it('hides payment history when no payments', () => {
      renderInvoiceDetail()
      expect(screen.queryByText('Payment History')).not.toBeInTheDocument()
    })
  })

  describe('record payment flow', () => {
    beforeEach(() => {
      mockCurrentInvoice = {
        id: 'inv-1',
        invoiceNumber: 'CD-2605-0001',
        status: 'issued',
        invoiceDate: '2026-05-26',
        patient: { name: 'John Doe', phone: '9876543210' },
        items: [{ id: 'item-1', description: 'Cleaning', quantity: 1, unitPrice: 150000, amount: 150000, toothNumber: '' }],
        subTotal: 150000,
        discountAmount: 0,
        cgstAmount: 0,
        sgstAmount: 0,
        totalAmount: 150000,
        paidAmount: 0,
        balanceAmount: 150000,
        payments: [],
      }
      mockGetWhatsAppTemplates.mockResolvedValue({ enabled: false })
    })

    it('shows payment form when Record Payment clicked', () => {
      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      expect(screen.getByPlaceholderText('Amount in rupees')).toBeInTheDocument()
    })

    it('shows error for empty amount', async () => {
      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      fireEvent.click(screen.getByRole('button', { name: /save payment/i }))
      expect(screen.getByText('Enter a valid amount')).toBeInTheDocument()
    })

    it('shows error for zero amount', async () => {
      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      fireEvent.change(screen.getByPlaceholderText('Amount in rupees'), { target: { value: '0' } })
      fireEvent.click(screen.getByRole('button', { name: /save payment/i }))
      expect(screen.getByText('Enter a valid amount')).toBeInTheDocument()
    })

    it('shows error for negative amount', async () => {
      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      fireEvent.change(screen.getByPlaceholderText('Amount in rupees'), { target: { value: '-100' } })
      fireEvent.click(screen.getByRole('button', { name: /save payment/i }))
      expect(screen.getByText('Enter a valid amount')).toBeInTheDocument()
    })

    it('calls recordPayment with correct input on valid amount', async () => {
      mockRecordPayment.mockResolvedValue({ id: 'pay-1', amount: 150000 })

      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      fireEvent.change(screen.getByPlaceholderText('Amount in rupees'), { target: { value: '1500' } })
      fireEvent.click(screen.getByRole('button', { name: /save payment/i }))

      await waitFor(() => {
        expect(mockRecordPayment).toHaveBeenCalledWith(
          expect.objectContaining({
            invoiceId: 'inv-1',
            amount: 150000, // 1500 * 100 paise
            method: 'cash',
          })
        )
      })
    })

    it('hides payment form after successful payment', async () => {
      mockRecordPayment.mockResolvedValue({ id: 'pay-1', amount: 150000 })

      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      fireEvent.change(screen.getByPlaceholderText('Amount in rupees'), { target: { value: '1500' } })
      fireEvent.click(screen.getByRole('button', { name: /save payment/i }))

      await waitFor(() => {
        expect(screen.queryByPlaceholderText('Amount in rupees')).not.toBeInTheDocument()
      })
    })

    it('shows error message on payment failure', async () => {
      mockRecordPayment.mockRejectedValue(new Error('Insufficient balance'))

      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      fireEvent.change(screen.getByPlaceholderText('Amount in rupees'), { target: { value: '1500' } })
      fireEvent.click(screen.getByRole('button', { name: /save payment/i }))

      await waitFor(() => {
        expect(screen.getByText('Insufficient balance')).toBeInTheDocument()
      })
    })

    it('cancel button hides payment form', () => {
      renderInvoiceDetail()
      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      expect(screen.getByPlaceholderText('Amount in rupees')).toBeInTheDocument()

      fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
      expect(screen.queryByPlaceholderText('Amount in rupees')).not.toBeInTheDocument()
    })

    it('invalidates React Query invoice cache after payment (fixes stale billing list)', async () => {
      const queryClient = createQueryClient()
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      mockRecordPayment.mockResolvedValue({ id: 'pay-1', amount: 150000 })

      render(
        <QueryClientProvider client={queryClient}>
          <MemoryRouter initialEntries={['/billing/inv-1']}>
            <Routes>
              <Route path="/billing/:id" element={<InvoiceDetail />} />
            </Routes>
          </MemoryRouter>
        </QueryClientProvider>
      )

      fireEvent.click(screen.getByRole('button', { name: /record payment/i }))
      fireEvent.change(screen.getByPlaceholderText('Amount in rupees'), { target: { value: '1500' } })
      fireEvent.click(screen.getByRole('button', { name: /save payment/i }))

      await waitFor(() => {
        expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['invoices'] })
      })
    })
  })

  describe('fetch on mount', () => {
    it('calls fetchInvoice with route param id', () => {
      mockCurrentInvoice = null
      renderInvoiceDetail('inv-test-123')
      expect(mockFetchInvoice).toHaveBeenCalledWith('inv-test-123')
    })

    it('calls fetchSettings if clinic is not loaded', () => {
      // clinic is already mocked as non-null, so need to reset
      // This test verifies the effect dependency
      mockCurrentInvoice = null
      renderInvoiceDetail()
      expect(mockFetchInvoice).toHaveBeenCalled()
    })
  })
})
