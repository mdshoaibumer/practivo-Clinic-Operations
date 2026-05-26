import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAppointmentStore } from './appointmentStore'

// Mock window.go
const mockGetTodayAppointments = vi.fn()
const mockGetAppointmentsByDate = vi.fn()
const mockGetWeekAppointments = vi.fn()
const mockCreateAppointment = vi.fn()
const mockCancelAppointment = vi.fn()
const mockCompleteAppointment = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  useAppointmentStore.setState({
    todayAppointments: [],
    appointments: [],
    selectedDate: new Date().toISOString().split('T')[0],
    isLoading: false,
  })
  ;(window as any).go = {
    handler: {
      AppointmentHandler: {
        GetTodayAppointments: mockGetTodayAppointments,
        GetAppointmentsByDate: mockGetAppointmentsByDate,
        GetWeekAppointments: mockGetWeekAppointments,
        CreateAppointment: mockCreateAppointment,
        CancelAppointment: mockCancelAppointment,
        CompleteAppointment: mockCompleteAppointment,
      },
    },
  }
})

describe('useAppointmentStore', () => {
  it('has correct initial state', () => {
    const state = useAppointmentStore.getState()
    expect(state.todayAppointments).toEqual([])
    expect(state.appointments).toEqual([])
    expect(state.isLoading).toBe(false)
    expect(state.selectedDate).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })

  it('fetchToday loads today appointments', async () => {
    const appointments = [
      { id: '1', patientName: 'Test', date: '2026-05-26', startTime: '10:00' },
    ]
    mockGetTodayAppointments.mockResolvedValue(appointments)

    await useAppointmentStore.getState().fetchToday()

    expect(useAppointmentStore.getState().todayAppointments).toEqual(appointments)
    expect(useAppointmentStore.getState().isLoading).toBe(false)
  })

  it('fetchToday handles null response', async () => {
    mockGetTodayAppointments.mockResolvedValue(null)

    await useAppointmentStore.getState().fetchToday()

    expect(useAppointmentStore.getState().todayAppointments).toEqual([])
  })

  it('fetchToday resets loading on error', async () => {
    mockGetTodayAppointments.mockRejectedValue('[INTERNAL_ERROR] fail')

    await expect(useAppointmentStore.getState().fetchToday()).rejects.toBeDefined()

    expect(useAppointmentStore.getState().isLoading).toBe(false)
  })

  it('fetchByDate loads appointments for a specific date', async () => {
    const appointments = [
      { id: '2', patientName: 'Patient A', date: '2026-05-27', startTime: '14:00' },
    ]
    mockGetAppointmentsByDate.mockResolvedValue(appointments)

    await useAppointmentStore.getState().fetchByDate('2026-05-27')

    expect(useAppointmentStore.getState().appointments).toEqual(appointments)
    expect(mockGetAppointmentsByDate).toHaveBeenCalledWith('2026-05-27')
  })

  it('fetchWeek loads date range', async () => {
    const appointments = [{ id: '3' }, { id: '4' }]
    mockGetWeekAppointments.mockResolvedValue(appointments)

    await useAppointmentStore.getState().fetchWeek('2026-05-25', '2026-05-31')

    expect(useAppointmentStore.getState().appointments).toEqual(appointments)
    expect(mockGetWeekAppointments).toHaveBeenCalledWith('2026-05-25', '2026-05-31')
  })

  it('createAppointment calls backend and refreshes lists', async () => {
    const newApt = { id: '5', patientId: 'p1', date: '2026-05-26' }
    mockCreateAppointment.mockResolvedValue(newApt)
    mockGetTodayAppointments.mockResolvedValue([newApt])
    mockGetAppointmentsByDate.mockResolvedValue([newApt])

    const input = { patientId: 'p1', date: '2026-05-26', startTime: '10:00', endTime: '10:30' }
    const result = await useAppointmentStore.getState().createAppointment(input as any)

    expect(result).toEqual(newApt)
    expect(mockCreateAppointment).toHaveBeenCalledWith(input)
    expect(mockGetTodayAppointments).toHaveBeenCalled()
  })

  it('cancelAppointment calls backend with reason and refreshes', async () => {
    mockCancelAppointment.mockResolvedValue(undefined)
    mockGetTodayAppointments.mockResolvedValue([])
    mockGetAppointmentsByDate.mockResolvedValue([])

    await useAppointmentStore.getState().cancelAppointment('apt-1', 'Patient requested')

    expect(mockCancelAppointment).toHaveBeenCalledWith('apt-1', 'Patient requested')
    expect(mockGetTodayAppointments).toHaveBeenCalled()
  })

  it('completeAppointment calls backend and refreshes', async () => {
    mockCompleteAppointment.mockResolvedValue(undefined)
    mockGetTodayAppointments.mockResolvedValue([])
    mockGetAppointmentsByDate.mockResolvedValue([])

    await useAppointmentStore.getState().completeAppointment('apt-1')

    expect(mockCompleteAppointment).toHaveBeenCalledWith('apt-1')
    expect(mockGetTodayAppointments).toHaveBeenCalled()
  })

  it('setSelectedDate updates the date', () => {
    useAppointmentStore.getState().setSelectedDate('2026-06-01')
    expect(useAppointmentStore.getState().selectedDate).toBe('2026-06-01')
  })

  it('fetchByDate resets loading on error', async () => {
    mockGetAppointmentsByDate.mockRejectedValue('[INTERNAL_ERROR] fail')

    await expect(useAppointmentStore.getState().fetchByDate('2026-05-26')).rejects.toBeDefined()

    expect(useAppointmentStore.getState().isLoading).toBe(false)
  })

  it('fetchByDate handles null response', async () => {
    mockGetAppointmentsByDate.mockResolvedValue(null)

    await useAppointmentStore.getState().fetchByDate('2026-05-26')

    expect(useAppointmentStore.getState().appointments).toEqual([])
  })

  it('fetchWeek resets loading on error', async () => {
    mockGetWeekAppointments.mockRejectedValue('[INTERNAL_ERROR] fail')

    await expect(useAppointmentStore.getState().fetchWeek('2026-05-25', '2026-05-31')).rejects.toBeDefined()

    expect(useAppointmentStore.getState().isLoading).toBe(false)
  })

  it('fetchWeek handles null response', async () => {
    mockGetWeekAppointments.mockResolvedValue(null)

    await useAppointmentStore.getState().fetchWeek('2026-05-25', '2026-05-31')

    expect(useAppointmentStore.getState().appointments).toEqual([])
  })

  it('createAppointment throws parsed error on failure', async () => {
    mockCreateAppointment.mockRejectedValue('[VALIDATION_ERROR] Time conflict')

    const input = { patientId: 'p1', date: '2026-05-26', startTime: '10:00', endTime: '10:30' }
    await expect(useAppointmentStore.getState().createAppointment(input as any)).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Time conflict',
    })
  })

  it('cancelAppointment throws parsed error on failure', async () => {
    mockCancelAppointment.mockRejectedValue('[VALIDATION_ERROR] Already cancelled')

    await expect(
      useAppointmentStore.getState().cancelAppointment('apt-1', 'reason')
    ).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Already cancelled',
    })
  })

  it('completeAppointment throws parsed error on failure', async () => {
    mockCompleteAppointment.mockRejectedValue('[VALIDATION_ERROR] Already completed')

    await expect(
      useAppointmentStore.getState().completeAppointment('apt-1')
    ).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Already completed',
    })
  })
})
