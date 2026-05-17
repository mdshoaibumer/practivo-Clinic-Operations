import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { formatCurrency } from '@/lib/utils'
import { Users, Receipt, Calendar, IndianRupee, Plus, ArrowRight } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useToast } from '@/components/ui/use-toast'
import PageTransition from '@/components/PageTransition'
import { useEffect } from 'react'

export default function Dashboard() {
  const navigate = useNavigate()
  const { toast } = useToast()

  const { data: stats, isLoading: statsLoading, error: statsError } = useQuery({
    queryKey: ['dashboardStats'],
    queryFn: () => window.go.handler.DashboardHandler.GetDashboardStats(),
  })

  const { data: todayAppointments = [], isLoading: aptLoading, error: aptError } = useQuery({
    queryKey: ['todayAppointments'],
    queryFn: () => window.go.handler.AppointmentHandler.GetTodayAppointments(),
  })

  useEffect(() => {
    if (statsError || aptError) {
      toast({
        variant: "destructive",
        title: "Error fetching dashboard data",
        description: "Please check your database connection or try again.",
      })
    }
  }, [statsError, aptError, toast])

  const loading = statsLoading || aptLoading

  if (loading) {
    return <div className="text-muted-foreground">Loading dashboard...</div>
  }

  // Empty state: show onboarding guidance when clinic has no data
  const isEmpty = stats && stats.totalPatients === 0 && stats.todayAppointments === 0

  if (isEmpty) {
    return (
      <PageTransition className="space-y-6">
        <h1 className="text-2xl font-bold">Welcome to Clinmitra Dental!</h1>

        <Card className="border-primary/20 bg-primary/5">
          <CardContent className="py-8">
            <div className="text-center space-y-4">
              <div className="inline-flex items-center justify-center w-16 h-16 bg-primary/10 rounded-full">
                <Users className="h-8 w-8 text-primary" />
              </div>
              <h2 className="text-xl font-semibold">Get Started in 3 Steps</h2>
              <p className="text-muted-foreground max-w-md mx-auto">
                Your clinic management system is ready. Follow these steps to start managing your practice efficiently.
              </p>
            </div>
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => navigate('/patients?action=new')}>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="h-8 w-8 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center text-sm font-bold">1</div>
                <h3 className="font-semibold">Register Your First Patient</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                Add patient details — name, phone, and medical history. You can search them instantly later.
              </p>
              <Button size="sm" variant="outline" className="w-full">
                <Plus className="h-4 w-4 mr-2" /> Add Patient <ArrowRight className="h-4 w-4 ml-auto" />
              </Button>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => navigate('/appointments?action=new')}>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="h-8 w-8 rounded-full bg-green-100 text-green-700 flex items-center justify-center text-sm font-bold">2</div>
                <h3 className="font-semibold">Book an Appointment</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                Schedule patient visits. View all appointments at a glance on the calendar.
              </p>
              <Button size="sm" variant="outline" className="w-full">
                <Calendar className="h-4 w-4 mr-2" /> Book Appointment <ArrowRight className="h-4 w-4 ml-auto" />
              </Button>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => navigate('/billing?action=new')}>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="h-8 w-8 rounded-full bg-purple-100 text-purple-700 flex items-center justify-center text-sm font-bold">3</div>
                <h3 className="font-semibold">Create an Invoice</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                Generate professional invoices with treatments, GST, and print them instantly.
              </p>
              <Button size="sm" variant="outline" className="w-full">
                <Receipt className="h-4 w-4 mr-2" /> Create Invoice <ArrowRight className="h-4 w-4 ml-auto" />
              </Button>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardContent className="py-4">
            <p className="text-sm text-muted-foreground text-center">
              <strong>Tip:</strong> Use <kbd className="px-1.5 py-0.5 bg-muted rounded text-xs font-mono">Ctrl+K</kbd> for quick search, <kbd className="px-1.5 py-0.5 bg-muted rounded text-xs font-mono">Ctrl+N</kbd> for new patient, <kbd className="px-1.5 py-0.5 bg-muted rounded text-xs font-mono">Ctrl+B</kbd> for new invoice.
            </p>
          </CardContent>
        </Card>
      </PageTransition>
    )
  }

  return (
    <PageTransition className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="bg-gradient-to-br from-blue-50 to-white">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Today's Appointments</CardTitle>
            <Calendar className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-900">{stats?.todayAppointments || 0}</div>
          </CardContent>
        </Card>

        <Card className="bg-gradient-to-br from-green-50 to-white">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Patients</CardTitle>
            <Users className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-900">{stats?.totalPatients || 0}</div>
          </CardContent>
        </Card>

        <Card className="bg-gradient-to-br from-indigo-50 to-white">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Today's Revenue</CardTitle>
            <IndianRupee className="h-4 w-4 text-indigo-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-indigo-900">{formatCurrency(stats?.todayRevenue || 0)}</div>
          </CardContent>
        </Card>

        <Card className="bg-gradient-to-br from-red-50 to-white">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Outstanding Dues</CardTitle>
            <Receipt className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{formatCurrency(stats?.totalOutstanding || 0)}</div>
          </CardContent>
        </Card>
      </div>

      {/* Monthly Revenue */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">This Month's Revenue</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-3xl font-bold text-green-600">{formatCurrency(stats?.monthRevenue || 0)}</div>
        </CardContent>
      </Card>

      {/* Today's Appointments */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Today's Appointments</CardTitle>
        </CardHeader>
        <CardContent>
          {todayAppointments.length === 0 ? (
            <p className="text-muted-foreground text-sm">No appointments scheduled for today.</p>
          ) : (
            <div className="space-y-2">
              {todayAppointments.map((apt) => (
                <div key={apt.id} className="flex items-center justify-between p-3 bg-muted rounded-md border border-border">
                  <div>
                    <p className="font-medium">{apt.patient?.name || 'Unknown Patient'}</p>
                    <p className="text-sm text-muted-foreground">{apt.purpose || 'General checkup'}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium">{apt.startTime} - {apt.endTime}</p>
                    <p className="text-xs text-muted-foreground capitalize">{apt.status}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </PageTransition>
  )
}
