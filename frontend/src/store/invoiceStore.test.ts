import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useInvoiceStore } from './invoiceStore'

// Mock window.go
const mockListInvoices = vi.fn()
const mockGetInvoice = vi.fn()
const mockCreateInvoice = vi.fn()
const mockRecordPayment = vi.fn()
const mockVoidInvoice = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  useInvoiceStore.setState({
    invoices: [],
    totalCount: 0,
    currentInvoice: null,
    page: 1,
    statusFilter: '',
    isLoading: false,
  })
  ;(window as any).go = {
    handler: {
      InvoiceHandler: {
        ListInvoices: mockListInvoices,
        GetInvoice: mockGetInvoice,
        CreateInvoice: mockCreateInvoice,
        RecordPayment: mockRecordPayment,
        VoidInvoice: mockVoidInvoice,
      },
    },
  }
})

describe('useInvoiceStore', () => {
  it('has correct initial state', () => {
    const state = useInvoiceStore.getState()
    expect(state.invoices).toEqual([])
    expect(state.totalCount).toBe(0)
    expect(state.currentInvoice).toBeNull()
    expect(state.page).toBe(1)
    expect(state.statusFilter).toBe('')
    expect(state.isLoading).toBe(false)
  })

  it('fetchInvoices loads invoice list', async () => {
    mockListInvoices.mockResolvedValue({
      invoices: [
        { id: 'inv-1', invoiceNumber: 'INV-202601-001', status: 'issued' },
      ],
      total: 1,
    })

    await useInvoiceStore.getState().fetchInvoices()

    const state = useInvoiceStore.getState()
    expect(state.invoices).toHaveLength(1)
    expect(state.totalCount).toBe(1)
    expect(state.isLoading).toBe(false)
  })

  it('fetchInvoices handles null response', async () => {
    mockListInvoices.mockResolvedValue({ invoices: null, total: 0 })

    await useInvoiceStore.getState().fetchInvoices()

    expect(useInvoiceStore.getState().invoices).toEqual([])
  })

  it('fetchInvoices uses status filter', async () => {
    useInvoiceStore.setState({ statusFilter: 'paid' })
    mockListInvoices.mockResolvedValue({ invoices: [], total: 0 })

    await useInvoiceStore.getState().fetchInvoices()

    expect(mockListInvoices).toHaveBeenCalledWith(1, expect.any(Number), 'paid', '', '', '', '')
  })

  it('fetchInvoice loads single invoice', async () => {
    const invoice = { id: 'inv-1', invoiceNumber: 'INV-202601-001', totalAmount: 150000 }
    mockGetInvoice.mockResolvedValue(invoice)

    await useInvoiceStore.getState().fetchInvoice('inv-1')

    expect(useInvoiceStore.getState().currentInvoice).toEqual(invoice)
    expect(useInvoiceStore.getState().isLoading).toBe(false)
  })

  it('createInvoice creates and refreshes list', async () => {
    const newInvoice = { id: 'inv-2', invoiceNumber: 'INV-202601-002' }
    mockCreateInvoice.mockResolvedValue(newInvoice)
    mockListInvoices.mockResolvedValue({ invoices: [newInvoice], total: 1 })

    const input = { patientId: 'p1', items: [] }
    const result = await useInvoiceStore.getState().createInvoice(input as any)

    expect(result).toEqual(newInvoice)
    expect(mockCreateInvoice).toHaveBeenCalledWith(input)
    expect(mockListInvoices).toHaveBeenCalled()
  })

  it('recordPayment records and refreshes current invoice', async () => {
    useInvoiceStore.setState({
      currentInvoice: { id: 'inv-1' } as any,
    })

    const payment = { id: 'pay-1', amount: 50000, method: 'cash' }
    mockRecordPayment.mockResolvedValue(payment)
    mockGetInvoice.mockResolvedValue({ id: 'inv-1', paidAmount: 50000 })

    const input = { invoiceId: 'inv-1', amount: 50000, method: 'cash', paymentDate: '2026-05-26' }
    const result = await useInvoiceStore.getState().recordPayment(input as any)

    expect(result).toEqual(payment)
    expect(mockGetInvoice).toHaveBeenCalledWith('inv-1')
  })

  it('recordPayment throws parsed error on failure', async () => {
    useInvoiceStore.setState({ currentInvoice: { id: 'inv-1' } as any })
    mockRecordPayment.mockRejectedValue('[VALIDATION_ERROR] Amount exceeds balance')

    const input = { invoiceId: 'inv-1', amount: 999999, method: 'cash', paymentDate: '2026-05-26' }
    await expect(useInvoiceStore.getState().recordPayment(input as any)).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Amount exceeds balance',
    })
  })

  it('recordPayment skips refresh if no currentInvoice', async () => {
    useInvoiceStore.setState({ currentInvoice: null })
    const payment = { id: 'pay-1', amount: 50000 }
    mockRecordPayment.mockResolvedValue(payment)

    const input = { invoiceId: 'inv-1', amount: 50000, method: 'cash', paymentDate: '2026-05-26' }
    await useInvoiceStore.getState().recordPayment(input as any)

    expect(mockGetInvoice).not.toHaveBeenCalled()
  })

  it('voidInvoice calls backend and refreshes', async () => {
    mockVoidInvoice.mockResolvedValue(undefined)
    mockGetInvoice.mockResolvedValue({ id: 'inv-1', status: 'void' })

    await useInvoiceStore.getState().voidInvoice('inv-1', 'Duplicate invoice')

    expect(mockVoidInvoice).toHaveBeenCalledWith('inv-1', 'Duplicate invoice')
    expect(mockGetInvoice).toHaveBeenCalledWith('inv-1')
  })

  it('voidInvoice throws parsed error on failure', async () => {
    mockVoidInvoice.mockRejectedValue('[VALIDATION_ERROR] Cannot void paid invoice')

    await expect(
      useInvoiceStore.getState().voidInvoice('inv-1', 'test')
    ).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Cannot void paid invoice',
    })
  })

  it('createInvoice throws parsed error on failure', async () => {
    mockCreateInvoice.mockRejectedValue('[VALIDATION_ERROR] No items provided')

    await expect(
      useInvoiceStore.getState().createInvoice({ patientId: 'p1', items: [] } as any)
    ).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'No items provided',
    })
  })

  it('fetchInvoices resets loading on error', async () => {
    mockListInvoices.mockRejectedValue('[INTERNAL_ERROR] DB error')

    await expect(useInvoiceStore.getState().fetchInvoices()).rejects.toBeDefined()

    expect(useInvoiceStore.getState().isLoading).toBe(false)
  })

  it('fetchInvoice resets loading on error', async () => {
    mockGetInvoice.mockRejectedValue('[NOT_FOUND] Invoice not found')

    await expect(useInvoiceStore.getState().fetchInvoice('bad-id')).rejects.toBeDefined()

    expect(useInvoiceStore.getState().isLoading).toBe(false)
  })

  it('setPage updates page', () => {
    useInvoiceStore.getState().setPage(3)
    expect(useInvoiceStore.getState().page).toBe(3)
  })

  it('setStatusFilter updates filter and resets page', () => {
    useInvoiceStore.setState({ page: 5 })

    useInvoiceStore.getState().setStatusFilter('paid')

    const state = useInvoiceStore.getState()
    expect(state.statusFilter).toBe('paid')
    expect(state.page).toBe(1)
  })
})
