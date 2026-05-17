import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'
import { useUIStore } from '@/store/uiStore'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Kbd } from '@/components/ui/kbd'
import { useKeyboardShortcuts } from '@/lib/useKeyboardShortcuts'
import GlobalSearch from '@/components/GlobalSearch'
import {
  LayoutDashboard,
  Users,
  Receipt,
  Calendar,
  BarChart3,
  Settings,
  LogOut,
  Menu,
  ChevronLeft,
  Search,
  DatabaseBackup,
} from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useToast } from '@/components/ui/use-toast'

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, badgeKey: null },
  { path: '/patients', label: 'Patients', icon: Users, badgeKey: null },
  { path: '/billing', label: 'Billing', icon: Receipt, badgeKey: 'outstanding' as const },
  { path: '/appointments', label: 'Appointments', icon: Calendar, badgeKey: 'appointments' as const },
  { path: '/reports', label: 'Reports', icon: BarChart3, badgeKey: null },
  { path: '/settings', label: 'Settings', icon: Settings, badgeKey: null },
]

export default function MainLayout() {
  const { user, logout } = useAuthStore()
  const { sidebarCollapsed, toggleSidebar } = useUIStore()
  const navigate = useNavigate()
  const { toast } = useToast()

  useKeyboardShortcuts()

  // Fetch badge counts using React Query
  const { data: stats } = useQuery({
    queryKey: ['dashboardStats'],
    queryFn: () => window.go.handler.DashboardHandler.GetDashboardStats(),
    refetchInterval: 60000, // Refresh every 60 seconds
  })

  const badges = {
    appointments: stats?.todayAppointments || 0,
    outstanding: (stats?.totalOutstanding || 0) > 0 ? 1 : 0,
  }

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const handleBackup = async () => {
    try {
      const result = await window.go.handler.BackupHandler.CreateBackup("")
      toast({
        title: "Backup Successful",
        description: `Saved to ${result.filepath}`,
      })
    } catch (err: any) {
      toast({
        variant: "destructive",
        title: "Backup Failed",
        description: typeof err === 'string' ? err : "Could not complete backup.",
      })
    }
  }

  return (
    <div className="flex h-screen overflow-hidden print:h-auto print:overflow-visible">
      {/* Global Search Overlay */}
      <div className="print:hidden">
        <GlobalSearch />
      </div>

      {/* Sidebar */}
      <aside
        className={cn(
          "flex flex-col bg-white border-r transition-all duration-300 print:hidden",
          sidebarCollapsed ? "w-16" : "w-64"
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          {!sidebarCollapsed && (
            <h1 className="text-xl font-bold text-primary">Clinmitra Dental</h1>
          )}
          <Button variant="ghost" size="icon" onClick={toggleSidebar}>
            {sidebarCollapsed ? <Menu className="h-5 w-5" /> : <ChevronLeft className="h-5 w-5" />}
          </Button>
        </div>

        {/* Quick Search Trigger */}
        {!sidebarCollapsed && (
          <div className="px-3 pt-3">
            <button
              onClick={() => {
                const event = new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
                window.dispatchEvent(event)
              }}
              className="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-muted-foreground border rounded-md hover:bg-accent transition-colors"
            >
              <Search className="h-3.5 w-3.5" />
              <span className="flex-1 text-left">Search...</span>
              <Kbd>Ctrl+K</Kbd>
            </button>
          </div>
        )}

        {/* Navigation */}
        <nav className="flex-1 p-2 space-y-1 mt-1 overflow-y-auto">
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors relative",
                  isActive
                    ? "bg-primary/10 text-primary"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                  sidebarCollapsed && "justify-center"
                )
              }
            >
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {!sidebarCollapsed && <span className="flex-1">{item.label}</span>}
              {/* Badge */}
              {item.badgeKey === 'appointments' && badges.appointments > 0 && (
                <span className={cn(
                  "bg-primary text-white text-[10px] font-bold rounded-full flex items-center justify-center",
                  sidebarCollapsed ? "absolute top-0 right-0 h-4 w-4" : "h-5 min-w-[20px] px-1"
                )}>
                  {badges.appointments}
                </span>
              )}
              {item.badgeKey === 'outstanding' && badges.outstanding > 0 && (
                <span className={cn(
                  "bg-red-500 rounded-full",
                  sidebarCollapsed ? "absolute top-1 right-1 h-2 w-2" : "h-2 w-2"
                )} />
              )}
            </NavLink>
          ))}
        </nav>

        {/* Footer Actions */}
        <div className="p-4 border-t space-y-2">
          {!sidebarCollapsed && (
            <div className="mb-4">
              <p className="text-sm font-medium truncate">{user?.fullName}</p>
              <p className="text-xs text-muted-foreground capitalize">{user?.role}</p>
            </div>
          )}
          
          <Button
            variant="outline"
            size={sidebarCollapsed ? "icon" : "sm"}
            onClick={handleBackup}
            className={cn("w-full justify-start text-blue-600 hover:text-blue-700 hover:bg-blue-50 border-blue-200", sidebarCollapsed && "justify-center")}
            title="Backup Data"
          >
            <DatabaseBackup className="h-4 w-4" />
            {!sidebarCollapsed && <span className="ml-2">Backup Data</span>}
          </Button>

          <Button
            variant="ghost"
            size={sidebarCollapsed ? "icon" : "sm"}
            onClick={handleLogout}
            className={cn("w-full justify-start", sidebarCollapsed && "justify-center")}
            title="Logout"
          >
            <LogOut className="h-4 w-4" />
            {!sidebarCollapsed && <span className="ml-2">Logout</span>}
          </Button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto bg-gray-50 print:overflow-visible print:bg-white print:m-0 print:p-0">
        <div className="p-6 print:p-0">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
