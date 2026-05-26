import { describe, it, expect, vi, beforeEach } from 'vitest'
import { usePatientStore } from './patientStore'

// Mock window.go
const mockListPatients = vi.fn()
const mockGetPatient = vi.fn()
const mockGetPatientHistory = vi.fn()
const mockCreatePatient = vi.fn()
const mockUpdatePatient = vi.fn()
const mockDeletePatient = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  usePatientStore.setState({
    patients: [],
    totalCount: 0,
    currentPatient: null,
    patientHistory: [],
    searchQuery: '',
    page: 1,
    isLoading: false,
  })
  ;(window as any).go = {
    handler: {
      PatientHandler: {
        ListPatients: mockListPatients,
        GetPatient: mockGetPatient,
        GetPatientHistory: mockGetPatientHistory,
        CreatePatient: mockCreatePatient,
        UpdatePatient: mockUpdatePatient,
        DeletePatient: mockDeletePatient,
      },
    },
  }
})

describe('usePatientStore', () => {
  it('has correct initial state', () => {
    const state = usePatientStore.getState()
    expect(state.patients).toEqual([])
    expect(state.totalCount).toBe(0)
    expect(state.currentPatient).toBeNull()
    expect(state.searchQuery).toBe('')
    expect(state.page).toBe(1)
    expect(state.isLoading).toBe(false)
  })

  it('fetchPatients loads patient list', async () => {
    mockListPatients.mockResolvedValue({
      patients: [
        { id: '1', name: 'Ramesh Kumar', phone: '9876543210' },
        { id: '2', name: 'Suresh Patel', phone: '9876543211' },
      ],
      total: 2,
    })

    await usePatientStore.getState().fetchPatients()

    const state = usePatientStore.getState()
    expect(state.patients).toHaveLength(2)
    expect(state.totalCount).toBe(2)
    expect(state.isLoading).toBe(false)
    expect(mockListPatients).toHaveBeenCalledWith(1, expect.any(Number), '')
  })

  it('fetchPatients handles null response gracefully', async () => {
    mockListPatients.mockResolvedValue({ patients: null, total: 0 })

    await usePatientStore.getState().fetchPatients()

    expect(usePatientStore.getState().patients).toEqual([])
  })

  it('fetchPatients resets loading on error', async () => {
    mockListPatients.mockRejectedValue('[INTERNAL_ERROR] DB error')

    await expect(usePatientStore.getState().fetchPatients()).rejects.toBeDefined()

    expect(usePatientStore.getState().isLoading).toBe(false)
  })

  it('fetchPatient loads single patient', async () => {
    const patient = { id: '1', name: 'Test Patient', phone: '9876543210' }
    mockGetPatient.mockResolvedValue(patient)

    await usePatientStore.getState().fetchPatient('1')

    expect(usePatientStore.getState().currentPatient).toEqual(patient)
    expect(usePatientStore.getState().isLoading).toBe(false)
  })

  it('createPatient calls backend and refreshes list', async () => {
    const newPatient = { id: '3', name: 'New Patient', phone: '9876543212' }
    mockCreatePatient.mockResolvedValue(newPatient)
    mockListPatients.mockResolvedValue({ patients: [newPatient], total: 1 })

    const input = { name: 'New Patient', phone: '9876543212', gender: 'male' }
    const result = await usePatientStore.getState().createPatient(input as any)

    expect(result).toEqual(newPatient)
    expect(mockCreatePatient).toHaveBeenCalledWith(input)
    expect(mockListPatients).toHaveBeenCalled()
  })

  it('updatePatient updates currentPatient', async () => {
    const updated = { id: '1', name: 'Updated Name', phone: '9876543210' }
    mockUpdatePatient.mockResolvedValue(updated)

    const input = { name: 'Updated Name', phone: '9876543210', gender: 'male' }
    const result = await usePatientStore.getState().updatePatient('1', input as any)

    expect(result).toEqual(updated)
    expect(usePatientStore.getState().currentPatient).toEqual(updated)
  })

  it('deletePatient removes patient and refreshes list', async () => {
    mockDeletePatient.mockResolvedValue(undefined)
    mockListPatients.mockResolvedValue({ patients: [], total: 0 })

    await usePatientStore.getState().deletePatient('1')

    expect(mockDeletePatient).toHaveBeenCalledWith('1')
    expect(mockListPatients).toHaveBeenCalled()
  })

  it('setSearch updates query and resets page', () => {
    usePatientStore.setState({ page: 3 })

    usePatientStore.getState().setSearch('ramesh')

    const state = usePatientStore.getState()
    expect(state.searchQuery).toBe('ramesh')
    expect(state.page).toBe(1)
  })

  it('setPage updates current page', () => {
    usePatientStore.getState().setPage(5)
    expect(usePatientStore.getState().page).toBe(5)
  })

  it('fetchPatientHistory handles empty results', async () => {
    mockGetPatientHistory.mockResolvedValue(null)

    await usePatientStore.getState().fetchPatientHistory('patient-1')

    expect(usePatientStore.getState().patientHistory).toEqual([])
  })

  it('fetchPatientHistory stores results', async () => {
    const history = [
      { id: 't1', treatmentName: 'Root Canal', date: '2026-01-15' },
    ]
    mockGetPatientHistory.mockResolvedValue(history)

    await usePatientStore.getState().fetchPatientHistory('patient-1')

    expect(usePatientStore.getState().patientHistory).toEqual(history)
  })

  it('fetchPatientHistory handles error silently', async () => {
    mockGetPatientHistory.mockRejectedValue(new Error('DB error'))

    // Should not throw
    await usePatientStore.getState().fetchPatientHistory('patient-1')

    expect(usePatientStore.getState().patientHistory).toEqual([])
  })

  it('fetchPatient resets loading on error', async () => {
    mockGetPatient.mockRejectedValue('[NOT_FOUND] Patient not found')

    await expect(usePatientStore.getState().fetchPatient('bad-id')).rejects.toBeDefined()

    expect(usePatientStore.getState().isLoading).toBe(false)
  })

  it('createPatient throws parsed error on failure', async () => {
    mockCreatePatient.mockRejectedValue('[VALIDATION_ERROR] Phone already exists')

    await expect(
      usePatientStore.getState().createPatient({ name: 'Test', phone: '9876543210' } as any)
    ).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Phone already exists',
    })
  })

  it('updatePatient throws parsed error on failure', async () => {
    mockUpdatePatient.mockRejectedValue('[NOT_FOUND] Patient not found')

    await expect(
      usePatientStore.getState().updatePatient('bad-id', {} as any)
    ).rejects.toEqual({
      code: 'NOT_FOUND',
      message: 'Patient not found',
    })
  })

  it('deletePatient throws parsed error on failure', async () => {
    mockDeletePatient.mockRejectedValue('[VALIDATION_ERROR] Cannot delete patient with invoices')

    await expect(
      usePatientStore.getState().deletePatient('p1')
    ).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Cannot delete patient with invoices',
    })
  })

  it('fetchPatients uses search query from state', async () => {
    usePatientStore.setState({ searchQuery: 'kumar', page: 2 })
    mockListPatients.mockResolvedValue({ patients: [], total: 0 })

    await usePatientStore.getState().fetchPatients()

    expect(mockListPatients).toHaveBeenCalledWith(2, expect.any(Number), 'kumar')
  })
})
