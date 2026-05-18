import { HashRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from '@/components/ui/toaster'
import { useAuthStore } from '@/store/authStore'
import { useSettingsStore } from '@/store/settingsStore'
import { useEffect, useState } from 'react'
import ErrorBoundary from '@/components/ErrorBoundary'

import MainLayout from '@/layouts/MainLayout'
import AuthLayout from '@/layouts/AuthLayout'
import SplashScreen from '@/components/SplashScreen'
import Login from '@/pages/Login'
import SetupWizard from '@/pages/SetupWizard'
import Dashboard from '@/pages/Dashboard'
import Patients from '@/pages/Patients'
import PatientDetail from '@/pages/PatientDetail'
import Billing from '@/pages/Billing'
import InvoiceDetail from '@/pages/InvoiceDetail'
import Appointments from '@/pages/Appointments'
import Reports from '@/pages/Reports'
import Settings from '@/pages/Settings'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <>{children}</>
}

function SetupGuard({ children }: { children: React.ReactNode }) {
  const { isSetupComplete } = useSettingsStore()
  if (!isSetupComplete) return <Navigate to="/setup" replace />
  return <>{children}</>
}

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: true,
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
})

export default function App() {
  const { checkSession } = useAuthStore()
  const { checkSetup, isSetupComplete } = useSettingsStore()
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function init() {
      // Wait for Wails runtime to be ready
      const waitForWails = (): Promise<void> => {
        if (window.go) return Promise.resolve()
        return new Promise((resolve) => {
          const handler = () => resolve()
          document.addEventListener('wails:loaded', handler, { once: true })
          // Fallback: poll briefly in case event already fired
          setTimeout(() => {
            if (window.go) resolve()
          }, 500)
        })
      }

      await waitForWails()

      if (window.go) {
        await checkSetup()
        await checkSession()
      }
      setLoading(false)
    }
    init()
  }, [checkSession, checkSetup])

  if (loading) {
    return <SplashScreen />
  }

  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <HashRouter>
          <Routes>
            {/* Public routes */}
            <Route path="/setup" element={
              isSetupComplete ? <Navigate to="/login" replace /> : <AuthLayout><SetupWizard /></AuthLayout>
            } />
            <Route path="/login" element={
              <SetupGuard><AuthLayout><Login /></AuthLayout></SetupGuard>
            } />

            {/* Protected routes */}
            <Route path="/" element={
              <SetupGuard><ProtectedRoute><MainLayout /></ProtectedRoute></SetupGuard>
            }>
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<ErrorBoundary><Dashboard /></ErrorBoundary>} />
              <Route path="patients" element={<ErrorBoundary><Patients /></ErrorBoundary>} />
              <Route path="patients/:id" element={<ErrorBoundary><PatientDetail /></ErrorBoundary>} />
              <Route path="billing" element={<ErrorBoundary><Billing /></ErrorBoundary>} />
              <Route path="billing/:id" element={<ErrorBoundary><InvoiceDetail /></ErrorBoundary>} />
              <Route path="appointments" element={<ErrorBoundary><Appointments /></ErrorBoundary>} />
              <Route path="reports" element={<ErrorBoundary><Reports /></ErrorBoundary>} />
              <Route path="settings" element={<ErrorBoundary><Settings /></ErrorBoundary>} />
            </Route>

            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
          <Toaster />
        </HashRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  )
}
