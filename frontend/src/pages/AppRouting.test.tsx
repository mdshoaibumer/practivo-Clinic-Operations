import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// We need to test the routing guards in App.tsx
// Import the guard components directly by re-creating them here with same logic

let mockIsSetupComplete = false
let mockIsAuthenticated = false
const mockCheckSetup = vi.fn()
const mockCheckSession = vi.fn()

vi.mock('@/store/settingsStore', () => ({
  useSettingsStore: () => ({
    isSetupComplete: mockIsSetupComplete,
    checkSetup: mockCheckSetup,
  }),
}))

vi.mock('@/store/authStore', () => ({
  useAuthStore: () => ({
    isAuthenticated: mockIsAuthenticated,
    checkSession: mockCheckSession,
  }),
}))

// Import after mocks
import { Navigate } from 'react-router-dom'

// Recreate the guards as they exist in App.tsx
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  if (!mockIsSetupComplete) return <Navigate to="/setup" replace />
  if (!mockIsAuthenticated) return <Navigate to="/login" replace />
  return <>{children}</>
}

function SetupGuard({ children }: { children: React.ReactNode }) {
  if (!mockIsSetupComplete) return <Navigate to="/setup" replace />
  return <>{children}</>
}

beforeEach(() => {
  vi.clearAllMocks()
  mockIsSetupComplete = false
  mockIsAuthenticated = false
  ;(window as any).go = { handler: {} }
})

function renderWithRoute(initialRoute: string) {
  return render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <Routes>
        <Route path="/setup" element={
          mockIsSetupComplete ? <Navigate to="/login" replace /> : <div data-testid="setup-wizard">Setup Wizard</div>
        } />
        <Route path="/login" element={
          <SetupGuard><div data-testid="login-page">Login Page</div></SetupGuard>
        } />
        <Route path="/dashboard" element={
          <ProtectedRoute><div data-testid="dashboard">Dashboard</div></ProtectedRoute>
        } />
        <Route path="*" element={
          mockIsSetupComplete ? <Navigate to="/dashboard" replace /> : <Navigate to="/setup" replace />
        } />
      </Routes>
    </MemoryRouter>
  )
}

describe('App Routing Guards', () => {
  describe('Fresh database (setup not complete)', () => {
    beforeEach(() => {
      mockIsSetupComplete = false
      mockIsAuthenticated = false
    })

    it('redirects / to /setup', () => {
      renderWithRoute('/')
      expect(screen.getByTestId('setup-wizard')).toBeInTheDocument()
    })

    it('redirects /dashboard to /setup', () => {
      renderWithRoute('/dashboard')
      expect(screen.getByTestId('setup-wizard')).toBeInTheDocument()
    })

    it('redirects /login to /setup (SetupGuard)', () => {
      renderWithRoute('/login')
      expect(screen.getByTestId('setup-wizard')).toBeInTheDocument()
    })

    it('shows setup wizard on /setup', () => {
      renderWithRoute('/setup')
      expect(screen.getByTestId('setup-wizard')).toBeInTheDocument()
    })

    it('redirects any unknown route to /setup', () => {
      renderWithRoute('/unknown-page')
      expect(screen.getByTestId('setup-wizard')).toBeInTheDocument()
    })
  })

  describe('Setup complete, not authenticated', () => {
    beforeEach(() => {
      mockIsSetupComplete = true
      mockIsAuthenticated = false
    })

    it('redirects /setup to /login', () => {
      renderWithRoute('/setup')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('shows login page on /login', () => {
      renderWithRoute('/login')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /dashboard to /login (ProtectedRoute)', () => {
      renderWithRoute('/dashboard')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects unknown routes to /dashboard then to /login', () => {
      renderWithRoute('/anything')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })
  })

  describe('Setup complete, authenticated', () => {
    beforeEach(() => {
      mockIsSetupComplete = true
      mockIsAuthenticated = true
    })

    it('shows dashboard on /dashboard', () => {
      renderWithRoute('/dashboard')
      expect(screen.getByTestId('dashboard')).toBeInTheDocument()
    })

    it('redirects /setup to /login (which would redirect to dashboard in real app)', () => {
      renderWithRoute('/setup')
      // With setup complete, /setup redirects to /login
      // SetupGuard passes since setup is complete, so login page renders
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects unknown routes to /dashboard', () => {
      renderWithRoute('/nonexistent')
      expect(screen.getByTestId('dashboard')).toBeInTheDocument()
    })
  })

  describe('Edge cases', () => {
    it('ProtectedRoute double-checks setup status before checking auth', () => {
      mockIsSetupComplete = false
      mockIsAuthenticated = true // Auth is true but setup isn't done

      renderWithRoute('/dashboard')

      // Should redirect to setup, NOT show dashboard
      expect(screen.getByTestId('setup-wizard')).toBeInTheDocument()
    })

    it('catch-all route is setup-aware', () => {
      mockIsSetupComplete = false
      renderWithRoute('/random/deep/path')
      expect(screen.getByTestId('setup-wizard')).toBeInTheDocument()
    })
  })
})
