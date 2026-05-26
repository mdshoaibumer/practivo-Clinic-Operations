import { create } from 'zustand'
import type { ClinicSettings, Treatment } from '@/types/models'
import type { SetupInput } from '@/types/api'
import { parseError } from '@/lib/api'

interface SettingsState {
  isSetupComplete: boolean
  clinic: ClinicSettings | null
  treatments: Treatment[]
  isLoading: boolean
  checkSetup: () => Promise<void>
  completeSetup: (input: SetupInput) => Promise<void>
  fetchSettings: () => Promise<void>
  updateSettings: (settings: ClinicSettings) => Promise<void>
  fetchTreatments: () => Promise<void>
}

export const useSettingsStore = create<SettingsState>((set) => ({
  isSetupComplete: false,
  clinic: null,
  treatments: [],
  isLoading: false,

  checkSetup: async () => {
    try {
      const complete = await window.go.handler.SettingsHandler.IsSetupComplete()
      // Explicitly coerce to boolean — Wails may return non-boolean truthy values
      const isComplete = complete === true
      console.log('[Setup] IsSetupComplete returned:', complete, '→ coerced:', isComplete)
      set({ isSetupComplete: isComplete })
    } catch (error) {
      console.error('Failed to check setup status:', error)
      set({ isSetupComplete: false })
    }
  },

  completeSetup: async (input: SetupInput) => {
    try {
      await window.go.handler.SettingsHandler.CompleteSetup(input)
      set({ isSetupComplete: true })
    } catch (error) {
      throw parseError(error)
    }
  },

  fetchSettings: async () => {
    set({ isLoading: true })
    try {
      const clinic = await window.go.handler.SettingsHandler.GetClinicSettings()
      const treatments = await window.go.handler.SettingsHandler.ListTreatments()
      set({ clinic, treatments, isLoading: false })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  updateSettings: async (settings: ClinicSettings) => {
    try {
      await window.go.handler.SettingsHandler.UpdateClinicSettings(settings)
      set({ clinic: settings })
    } catch (error) {
      throw parseError(error)
    }
  },

  fetchTreatments: async () => {
    try {
      const treatments = await window.go.handler.SettingsHandler.ListTreatments()
      set({ treatments })
    } catch (error) {
      throw parseError(error)
    }
  },
}))
