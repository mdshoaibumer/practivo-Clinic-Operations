import { HashRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from '@/components/ui/toaster'
import { useAuthStore } from '@/store/authStore'
import { useSettingsStore } from '@/store/settingsStore'
import { useEffect, useState, lazy, Suspense } from 'react'
import ErrorBoundary from '@/components/ErrorBoundary'

import MainLayout from '@/layouts/MainLayout'
import AuthLayout from '@/layouts/AuthLayout'
import SplashScreen from '@/components/SplashScreen'
import Login from '@/pages/Login'
import SetupWizard from '@/pages/SetupWizard'

// Lazy-load pages for code splitting and faster initial load
const Dashboard = lazy(() => import('@/pages/Dashboard'))
const Patients = lazy(() => import('@/pages/Patients'))
const PatientDetail = lazy(() => import('@/pages/PatientDetail'))
const Billing = lazy(() => import('@/pages/Billing'))
const InvoiceDetail = lazy(() => import('@/pages/InvoiceDetail'))
const Appointments = lazy(() => import('@/pages/Appointments'))
const Reports = lazy(() => import('@/pages/Reports'))
const Settings = lazy(() => import('@/pages/Settings'))

// Lightweight fallback for lazy-loaded pages
function PageLoader() {
  return (
    <div className="flex items-center justify-center h-48">
      <div className="h-5 w-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
    </div>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore()
  const { isSetupComplete } = useSettingsStore()
  if (!isSetupComplete) return <Navigate to="/setup" replace />
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
          const interval = setInterval(() => {
            if (window.go) {
              clearInterval(interval)
              resolve()
            }
          }, 100)
          // Hard timeout: resolve after 3s even without window.go
          setTimeout(() => {
            clearInterval(interval)
            resolve()
          }, 3000)
        })
      }

      await waitForWails()

      if (window.go) {
        await checkSetup()
        await checkSession()
      } else {
        console.warn('[Init] window.go not available after wait — treating as fresh setup')
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
              <Route path="dashboard" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><Dashboard /></Suspense></ErrorBoundary>} />
              <Route path="patients" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><Patients /></Suspense></ErrorBoundary>} />
              <Route path="patients/:id" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><PatientDetail /></Suspense></ErrorBoundary>} />
              <Route path="billing" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><Billing /></Suspense></ErrorBoundary>} />
              <Route path="billing/:id" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><InvoiceDetail /></Suspense></ErrorBoundary>} />
              <Route path="appointments" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><Appointments /></Suspense></ErrorBoundary>} />
              <Route path="reports" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><Reports /></Suspense></ErrorBoundary>} />
              <Route path="settings" element={<ErrorBoundary><Suspense fallback={<PageLoader />}><Settings /></Suspense></ErrorBoundary>} />
            </Route>

            <Route path="*" element={
              isSetupComplete ? <Navigate to="/dashboard" replace /> : <Navigate to="/setup" replace />
            } />
          </Routes>
          <Toaster />
        </HashRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  )
}
