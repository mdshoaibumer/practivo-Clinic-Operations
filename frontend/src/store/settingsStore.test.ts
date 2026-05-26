import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useSettingsStore } from './settingsStore'

// Mock window.go
const mockIsSetupComplete = vi.fn()
const mockCompleteSetup = vi.fn()
const mockGetClinicSettings = vi.fn()
const mockUpdateClinicSettings = vi.fn()
const mockListTreatments = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  useSettingsStore.setState({
    isSetupComplete: false,
    clinic: null,
    treatments: [],
    isLoading: false,
  })
  ;(window as any).go = {
    handler: {
      SettingsHandler: {
        IsSetupComplete: mockIsSetupComplete,
        CompleteSetup: mockCompleteSetup,
        GetClinicSettings: mockGetClinicSettings,
        UpdateClinicSettings: mockUpdateClinicSettings,
        ListTreatments: mockListTreatments,
      },
    },
  }
})

describe('useSettingsStore', () => {
  describe('initial state', () => {
    it('has correct initial state', () => {
      const state = useSettingsStore.getState()
      expect(state.isSetupComplete).toBe(false)
      expect(state.clinic).toBeNull()
      expect(state.treatments).toEqual([])
      expect(state.isLoading).toBe(false)
    })
  })

  describe('checkSetup', () => {
    it('sets isSetupComplete to true when backend returns true', async () => {
      mockIsSetupComplete.mockResolvedValue(true)

      await useSettingsStore.getState().checkSetup()

      expect(useSettingsStore.getState().isSetupComplete).toBe(true)
    })

    it('sets isSetupComplete to false when backend returns false', async () => {
      mockIsSetupComplete.mockResolvedValue(false)

      await useSettingsStore.getState().checkSetup()

      expect(useSettingsStore.getState().isSetupComplete).toBe(false)
    })

    it('coerces truthy non-boolean values to false (strict equality)', async () => {
      // Wails might return 1, "true", or other truthy values
      mockIsSetupComplete.mockResolvedValue(1)

      await useSettingsStore.getState().checkSetup()

      expect(useSettingsStore.getState().isSetupComplete).toBe(false)
    })

    it('coerces string "true" to false (strict equality)', async () => {
      mockIsSetupComplete.mockResolvedValue('true')

      await useSettingsStore.getState().checkSetup()

      expect(useSettingsStore.getState().isSetupComplete).toBe(false)
    })

    it('coerces undefined to false', async () => {
      mockIsSetupComplete.mockResolvedValue(undefined)

      await useSettingsStore.getState().checkSetup()

      expect(useSettingsStore.getState().isSetupComplete).toBe(false)
    })

    it('coerces null to false', async () => {
      mockIsSetupComplete.mockResolvedValue(null)

      await useSettingsStore.getState().checkSetup()

      expect(useSettingsStore.getState().isSetupComplete).toBe(false)
    })

    it('sets isSetupComplete to false on error', async () => {
      // Start with true to verify it resets
      useSettingsStore.setState({ isSetupComplete: true })
      mockIsSetupComplete.mockRejectedValue(new Error('DB not ready'))

      await useSettingsStore.getState().checkSetup()

      expect(useSettingsStore.getState().isSetupComplete).toBe(false)
    })

    it('does not throw on error', async () => {
      mockIsSetupComplete.mockRejectedValue(new Error('network error'))

      await expect(useSettingsStore.getState().checkSetup()).resolves.not.toThrow()
    })
  })

  describe('completeSetup', () => {
    it('calls backend and sets isSetupComplete to true', async () => {
      mockCompleteSetup.mockResolvedValue(undefined)

      const input = {
        clinicName: 'Test Clinic',
        doctorName: 'Dr. Test',
        phone: '9876543210',
        adminFullName: 'Admin',
        adminUsername: 'admin',
        adminPassword: 'Password1!',
      }

      await useSettingsStore.getState().completeSetup(input as any)

      expect(mockCompleteSetup).toHaveBeenCalledWith(input)
      expect(useSettingsStore.getState().isSetupComplete).toBe(true)
    })

    it('throws parsed error on failure', async () => {
      mockCompleteSetup.mockRejectedValue('[VALIDATION_ERROR] Clinic name is required')

      await expect(
        useSettingsStore.getState().completeSetup({} as any)
      ).rejects.toEqual({
        code: 'VALIDATION_ERROR',
        message: 'Clinic name is required',
      })
    })

    it('does not set isSetupComplete on failure', async () => {
      mockCompleteSetup.mockRejectedValue('[VALIDATION_ERROR] Error')

      try {
        await useSettingsStore.getState().completeSetup({} as any)
      } catch {
        // Expected
      }

      expect(useSettingsStore.getState().isSetupComplete).toBe(false)
    })
  })

  describe('fetchSettings', () => {
    it('fetches clinic and treatments', async () => {
      const clinic = { id: '1', clinicName: 'Test Clinic', setupComplete: true }
      const treatments = [{ id: 't1', name: 'Cleaning', defaultPrice: 150000 }]
      mockGetClinicSettings.mockResolvedValue(clinic)
      mockListTreatments.mockResolvedValue(treatments)

      await useSettingsStore.getState().fetchSettings()

      const state = useSettingsStore.getState()
      expect(state.clinic).toEqual(clinic)
      expect(state.treatments).toEqual(treatments)
      expect(state.isLoading).toBe(false)
    })

    it('sets isLoading during fetch', async () => {
      mockGetClinicSettings.mockImplementation(() => {
        expect(useSettingsStore.getState().isLoading).toBe(true)
        return Promise.resolve({})
      })
      mockListTreatments.mockResolvedValue([])

      await useSettingsStore.getState().fetchSettings()

      expect(useSettingsStore.getState().isLoading).toBe(false)
    })

    it('resets isLoading on error', async () => {
      mockGetClinicSettings.mockRejectedValue('[INTERNAL_ERROR] DB error')

      try {
        await useSettingsStore.getState().fetchSettings()
      } catch {
        // Expected
      }

      expect(useSettingsStore.getState().isLoading).toBe(false)
    })

    it('throws parsed error on failure', async () => {
      mockGetClinicSettings.mockRejectedValue('[NOT_FOUND] Clinic not configured')

      await expect(useSettingsStore.getState().fetchSettings()).rejects.toEqual({
        code: 'NOT_FOUND',
        message: 'Clinic not configured',
      })
    })
  })

  describe('updateSettings', () => {
    it('updates clinic settings in store after success', async () => {
      mockUpdateClinicSettings.mockResolvedValue(undefined)
      const settings = { id: '1', clinicName: 'Updated Clinic' } as any

      await useSettingsStore.getState().updateSettings(settings)

      expect(mockUpdateClinicSettings).toHaveBeenCalledWith(settings)
      expect(useSettingsStore.getState().clinic).toEqual(settings)
    })

    it('throws parsed error on failure', async () => {
      mockUpdateClinicSettings.mockRejectedValue('[VALIDATION_ERROR] Invalid phone')

      await expect(
        useSettingsStore.getState().updateSettings({} as any)
      ).rejects.toEqual({
        code: 'VALIDATION_ERROR',
        message: 'Invalid phone',
      })
    })

    it('does not update store on failure', async () => {
      const originalClinic = { id: '1', clinicName: 'Original' } as any
      useSettingsStore.setState({ clinic: originalClinic })
      mockUpdateClinicSettings.mockRejectedValue('[INTERNAL_ERROR] Error')

      try {
        await useSettingsStore.getState().updateSettings({ clinicName: 'New' } as any)
      } catch {
        // Expected
      }

      expect(useSettingsStore.getState().clinic).toEqual(originalClinic)
    })
  })

  describe('fetchTreatments', () => {
    it('fetches and sets treatments', async () => {
      const treatments = [
        { id: 't1', name: 'Cleaning', defaultPrice: 150000 },
        { id: 't2', name: 'Filling', defaultPrice: 200000 },
      ]
      mockListTreatments.mockResolvedValue(treatments)

      await useSettingsStore.getState().fetchTreatments()

      expect(useSettingsStore.getState().treatments).toEqual(treatments)
    })

    it('throws parsed error on failure', async () => {
      mockListTreatments.mockRejectedValue('[INTERNAL_ERROR] Error')

      await expect(useSettingsStore.getState().fetchTreatments()).rejects.toEqual({
        code: 'INTERNAL_ERROR',
        message: 'Error',
      })
    })
  })
})
