import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Billing from './Billing'

// Mock stores
const mockTreatments = [
  { id: 't1', name: 'Cleaning', code: 'CLEAN', defaultPrice: 150000, category: 'Preventive', isActive: true },
  { id: 't2', name: 'Filling', code: 'FILL', defaultPrice: 200000, category: 'Restorative', isActive: true },
]
const mockPatients = [
  { id: 'p1', name: 'John Doe', phone: '9876543210' },
  { id: 'p2', name: 'Jane Smith', phone: '9876543211' },
]
const mockFetchTreatments = vi.fn()
const mockFetchPatients = vi.fn()

vi.mock('@/store/settingsStore', () => ({
  useSettingsStore: () => ({
    treatments: mockTreatments,
    fetchTreatments: mockFetchTreatments,
  }),
}))

vi.mock('@/store/patientStore', () => ({
  usePatientStore: () => ({
    patients: mockPatients,
    fetchPatients: mockFetchPatients,
  }),
}))

// Mock toast
const mockToast = vi.fn()
vi.mock('@/components/ui/use-toast', () => ({
  useToast: () => ({ toast: mockToast }),
}))

// Mock PageTransition to just render children
vi.mock('@/components/PageTransition', () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

// Mock skeletons
vi.mock('@/components/ui/skeletons', () => ({
  InvoiceListSkeleton: () => <div data-testid="loading-skeleton" />,
}))

// Mock InvoiceConfirmDialog
vi.mock('@/components/billing/InvoiceConfirmDialog', () => ({
  default: ({ open, onConfirm, onCancel }: any) =>
    open ? (
      <div data-testid="confirm-dialog">
        <button onClick={onConfirm}>Confirm</button>
        <button onClick={onCancel}>Cancel</button>
      </div>
    ) : null,
}))

// Mock window.go
const mockListInvoices = vi.fn()
const mockCreateInvoice = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  ;(window as any).go = {
    handler: {
      InvoiceHandler: {
        ListInvoices: mockListInvoices,
        CreateInvoice: mockCreateInvoice,
      },
    },
  }
})

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderBilling() {
  const queryClient = createQueryClient()
  return {
    queryClient,
    ...render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/billing']}>
          <Billing />
        </MemoryRouter>
      </QueryClientProvider>
    ),
  }
}

describe('Billing Page', () => {
  describe('invoice list', () => {
    it('shows loading skeleton while fetching', () => {
      mockListInvoices.mockReturnValue(new Promise(() => {})) // Never resolves
      renderBilling()
      expect(screen.getByTestId('loading-skeleton')).toBeInTheDocument()
    })

    it('renders invoice list after loading', async () => {
      mockListInvoices.mockResolvedValue({
        invoices: [
          {
            id: 'inv-1',
            invoiceNumber: 'CD-2605-0001',
            status: 'issued',
            invoiceDate: '2026-05-26',
            totalAmount: 450000,
            balanceAmount: 450000,
            patient: { name: 'John Doe' },
          },
        ],
        total: 1,
      })

      renderBilling()

      await waitFor(() => {
        expect(screen.getByText('CD-2605-0001')).toBeInTheDocument()
      })
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    it('shows empty state when no invoices', async () => {
      mockListInvoices.mockResolvedValue({ invoices: [], total: 0 })

      renderBilling()

      await waitFor(() => {
        expect(screen.queryByText('CD-2605-0001')).not.toBeInTheDocument()
      })
    })

    it('handles null invoices array from backend', async () => {
      mockListInvoices.mockResolvedValue({ invoices: null, total: 0 })

      renderBilling()

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalled()
      })
      // Should not crash
    })
  })

  describe('status filtering', () => {
    beforeEach(() => {
      mockListInvoices.mockResolvedValue({ invoices: [], total: 0 })
    })

    it('fetches all invoices by default (empty status filter)', async () => {
      renderBilling()

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalledWith(1, expect.any(Number), '', '', '', '', '')
      })
    })

    it('refetches when status filter changes', async () => {
      renderBilling()

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalledTimes(1)
      })

      // Find and click a filter button (Unpaid, Paid, etc.)
      const unpaidButton = screen.getByRole('button', { name: /unpaid/i })
      fireEvent.click(unpaidButton)

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalledWith(
          1, expect.any(Number), expect.stringMatching(/issued|unpaid/i), '', '', '', ''
        )
      })
    })
  })

  describe('pagination', () => {
    it('passes page number to query', async () => {
      mockListInvoices.mockResolvedValue({
        invoices: Array(20).fill(null).map((_, i) => ({
          id: `inv-${i}`,
          invoiceNumber: `CD-001-${i}`,
          status: 'issued',
          invoiceDate: '2026-05-26',
          totalAmount: 100000,
          balanceAmount: 100000,
          patient: { name: `Patient ${i}` },
        })),
        total: 40, // More than one page
      })

      renderBilling()

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalledWith(1, expect.any(Number), '', '', '', '', '')
      })
    })
  })

  describe('React Query cache behavior', () => {
    it('uses queryKey with page and statusFilter for proper cache management', async () => {
      const queryClient = createQueryClient()
      mockListInvoices.mockResolvedValue({ invoices: [], total: 0 })

      render(
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>
            <Billing />
          </MemoryRouter>
        </QueryClientProvider>
      )

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalled()
      })

      // Verify the query was stored with the correct key
      const queries = queryClient.getQueryCache().getAll()
      const invoiceQuery = queries.find(q => q.queryKey[0] === 'invoices')
      expect(invoiceQuery).toBeDefined()
      expect(invoiceQuery?.queryKey).toEqual(['invoices', 1, ''])
    })

    it('invalidating invoices query causes refetch', async () => {
      const queryClient = createQueryClient()
      mockListInvoices.mockResolvedValue({
        invoices: [{ id: 'inv-1', invoiceNumber: 'CD-001', status: 'issued', invoiceDate: '2026-05-26', totalAmount: 100000, balanceAmount: 100000, patient: { name: 'Test' } }],
        total: 1,
      })

      render(
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>
            <Billing />
          </MemoryRouter>
        </QueryClientProvider>
      )

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalledTimes(1)
      })

      // Simulate what InvoiceDetail does after recording payment
      mockListInvoices.mockResolvedValue({
        invoices: [{ id: 'inv-1', invoiceNumber: 'CD-001', status: 'paid', invoiceDate: '2026-05-26', totalAmount: 100000, balanceAmount: 0, patient: { name: 'Test' } }],
        total: 1,
      })

      queryClient.invalidateQueries({ queryKey: ['invoices'] })

      await waitFor(() => {
        expect(mockListInvoices).toHaveBeenCalledTimes(2)
      })
    })
  })

  describe('error handling', () => {
    it('shows toast on fetch error', async () => {
      mockListInvoices.mockRejectedValue(new Error('Network error'))

      renderBilling()

      await waitFor(() => {
        expect(mockToast).toHaveBeenCalledWith(
          expect.objectContaining({
            variant: 'destructive',
            title: 'Error fetching invoices',
          })
        )
      })
    })
  })
})
