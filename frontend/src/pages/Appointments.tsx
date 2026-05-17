import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { usePatientStore } from '@/store/patientStore'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { appointmentSchema, type AppointmentFormData } from '@/lib/validators'
import { formatTime, getStatusColor, getTodayDate } from '@/lib/utils'
import { APPOINTMENT_STATUS_LABELS } from '@/lib/constants'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { DatePicker } from '@/components/ui/date-picker'
import PatientAvatar from '@/components/PatientAvatar'
import WeekView from '@/components/appointments/WeekView'
import { Plus, ChevronLeft, ChevronRight, X, CheckCircle, CalendarDays, List } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@/components/ui/use-toast'
import PageTransition from '@/components/PageTransition'

type ViewMode = 'day' | 'week'

function getWeekStart(dateStr: string): string {
  const d = new Date(dateStr)
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1) // Monday start
  d.setDate(diff)
  return d.toISOString().split('T')[0]
}

function getWeekEnd(startStr: string): string {
  const d = new Date(startStr)
  d.setDate(d.getDate() + 6)
  return d.toISOString().split('T')[0]
}

export default function Appointments() {
  const [searchParams] = useSearchParams()
  const { toast } = useToast()
  const queryClient = useQueryClient()

  const { patients, fetchPatients } = usePatientStore()
  const [selectedDate, setSelectedDate] = useState(getTodayDate())
  const [viewMode, setViewMode] = useState<ViewMode>('day')
  const [showForm, setShowForm] = useState(false)
  const [formError, setFormError] = useState('')

  const { register, handleSubmit, formState: { errors }, reset, setValue, control } = useForm<AppointmentFormData>({
    resolver: zodResolver(appointmentSchema),
    defaultValues: { date: selectedDate },
  })

  // Fetch appointments based on view mode
  const { data: appointments = [], isLoading, isError } = useQuery({
    queryKey: ['appointments', viewMode, selectedDate],
    queryFn: async () => {
      if (viewMode === 'day') {
        const res = await window.go.handler.AppointmentHandler.GetAppointmentsByDateRange(selectedDate, selectedDate)
        return res || []
      } else {
        const weekStart = getWeekStart(selectedDate)
        const weekEnd = getWeekEnd(weekStart)
        const res = await window.go.handler.AppointmentHandler.GetAppointmentsByDateRange(weekStart, weekEnd)
        return res || []
      }
    }
  })

  useEffect(() => {
    if (isError) {
      toast({
        variant: "destructive",
        title: "Error fetching appointments",
        description: "Could not load appointment data.",
      })
    }
  }, [isError, toast])

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: any) => window.go.handler.AppointmentHandler.CreateAppointment(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appointments'] })
      queryClient.invalidateQueries({ queryKey: ['dashboardStats'] })
      reset()
      setShowForm(false)
      toast({ title: "Appointment booked" })
    },
    onError: (err: any) => {
      setFormError(typeof err === 'string' ? err : err.message || 'Failed to create appointment')
    }
  })

  const cancelMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string, reason: string }) => window.go.handler.AppointmentHandler.CancelAppointment(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appointments'] })
      toast({ title: "Appointment cancelled" })
    }
  })

  const completeMutation = useMutation({
    mutationFn: (id: string) => window.go.handler.AppointmentHandler.CompleteAppointment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appointments'] })
      toast({ title: "Appointment marked as completed" })
    }
  })

  // Handle action=new from keyboard shortcut or global search
  useEffect(() => {
    if (searchParams.get('action') === 'new') {
      setShowForm(true)
    }
  }, [searchParams])

  useEffect(() => {
    if (showForm) fetchPatients()
  }, [showForm, fetchPatients])

  // Close form on Escape
  useEffect(() => {
    function handleEsc(e: KeyboardEvent) {
      if (e.key === 'Escape' && showForm) {
        setShowForm(false)
      }
    }
    window.addEventListener('keydown', handleEsc)
    return () => window.removeEventListener('keydown', handleEsc)
  }, [showForm])

  const changeDate = (days: number) => {
    const d = new Date(selectedDate)
    d.setDate(d.getDate() + days)
    const newDate = d.toISOString().split('T')[0]
    setSelectedDate(newDate)
    setValue('date', newDate)
  }

  const changeWeek = (weeks: number) => {
    const d = new Date(selectedDate)
    d.setDate(d.getDate() + (weeks * 7))
    setSelectedDate(d.toISOString().split('T')[0])
  }

  const onSubmit = (data: AppointmentFormData) => {
    setFormError('')
    createMutation.mutate({
      patientId: data.patientId,
      date: data.date,
      startTime: data.startTime,
      endTime: data.endTime,
      duration: data.duration || 30,
      notes: data.notes || '',
      purpose: data.purpose || '',
    })
  }

  const handleCancel = (id: string) => {
    if (confirm('Cancel this appointment?')) {
      cancelMutation.mutate({ id, reason: 'Cancelled by user' })
    }
  }

  const handleComplete = (id: string) => {
    completeMutation.mutate(id)
  }

  const isToday = selectedDate === getTodayDate()

  return (
    <PageTransition className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Appointments</h1>
        <div className="flex gap-2">
          {/* View Mode Toggle */}
          <div className="flex border rounded-md">
            <Button
              variant={viewMode === 'day' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setViewMode('day')}
              className="rounded-r-none"
            >
              <List className="h-4 w-4 mr-1" /> Day
            </Button>
            <Button
              variant={viewMode === 'week' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setViewMode('week')}
              className="rounded-l-none"
            >
              <CalendarDays className="h-4 w-4 mr-1" /> Week
            </Button>
          </div>
          <Button onClick={() => setShowForm(!showForm)}>
            <Plus className="h-4 w-4 mr-2" /> New Appointment
          </Button>
        </div>
      </div>

      {/* New Appointment Form */}
      {showForm && (
        <Card>
          <CardHeader><CardTitle>Book Appointment</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {formError && <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">{formError}</div>}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>Patient *</Label>
                  <select
                    {...register('patientId')}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="">Select patient...</option>
                    {patients.map(p => <option key={p.id} value={p.id}>{p.name} ({p.phone})</option>)}
                  </select>
                  {errors.patientId && <p className="text-sm text-red-500">{errors.patientId.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>Date *</Label>
                  <Controller
                    name="date"
                    control={control}
                    render={({ field }) => (
                      <DatePicker
                        value={field.value}
                        onChange={field.onChange}
                        placeholder="Select appointment date"
                      />
                    )}
                  />
                  {errors.date && <p className="text-sm text-red-500">{errors.date.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>Start Time *</Label>
                  <Input type="time" {...register('startTime')} />
                  {errors.startTime && <p className="text-sm text-red-500">{errors.startTime.message}</p>}
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>End Time *</Label>
                  <Input type="time" {...register('endTime')} />
                  {errors.endTime && <p className="text-sm text-red-500">{errors.endTime.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>Purpose</Label>
                  <Input {...register('purpose')} placeholder="Treatment purpose" />
                </div>
                <div className="space-y-2">
                  <Label>Notes</Label>
                  <Input {...register('notes')} placeholder="Additional notes" />
                </div>
              </div>
              <div className="flex gap-2">
                <Button type="submit" disabled={createMutation.isPending}>
                  {createMutation.isPending ? 'Booking...' : 'Book'}
                </Button>
                <Button type="button" variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Date Navigation */}
      <div className="flex items-center justify-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => viewMode === 'day' ? changeDate(-1) : changeWeek(-1)}>
          <ChevronLeft className="h-5 w-5" />
        </Button>
        <div className="text-center">
          <DatePicker
            value={selectedDate}
            onChange={(date) => setSelectedDate(date)}
            className="text-center"
          />
          {isToday && <p className="text-xs text-primary font-medium mt-1">Today</p>}
          {viewMode === 'week' && (
            <p className="text-xs text-muted-foreground mt-1">
              Week of {new Date(getWeekStart(selectedDate)).toLocaleDateString('en-IN', { day: 'numeric', month: 'short' })}
            </p>
          )}
        </div>
        <Button variant="ghost" size="icon" onClick={() => viewMode === 'day' ? changeDate(1) : changeWeek(1)}>
          <ChevronRight className="h-5 w-5" />
        </Button>
        {!isToday && (
          <Button variant="outline" size="sm" onClick={() => setSelectedDate(getTodayDate())}>
            Today
          </Button>
        )}
      </div>

      {/* Week View */}
      {viewMode === 'week' && (
        <WeekView
          appointments={appointments}
          weekStart={getWeekStart(selectedDate)}
          onComplete={handleComplete}
          onCancel={handleCancel}
        />
      )}

      {/* Day View - Appointments List */}
      {viewMode === 'day' && (
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 text-center text-muted-foreground">Loading...</div>
          ) : appointments.length === 0 ? (
            <div className="p-6 text-center text-muted-foreground">No appointments for this date.</div>
          ) : (
            <div className="divide-y">
              {appointments
                .sort((a: any, b: any) => a.startTime.localeCompare(b.startTime))
                .map((apt: any) => (
                <div key={apt.id} className="flex items-center justify-between p-4 hover:bg-muted/30 transition-colors">
                  <div className="flex items-center gap-4">
                    <div className="text-center min-w-[60px]">
                      <p className="text-sm font-bold">{formatTime(apt.startTime)}</p>
                      <p className="text-xs text-muted-foreground">{apt.duration}min</p>
                    </div>
                    <PatientAvatar name={apt.patient?.name || 'Patient'} size="sm" />
                    <div>
                      <p className="font-medium">{apt.patient?.name || 'Patient'}</p>
                      <p className="text-sm text-muted-foreground">
                        {apt.purpose || 'General'}
                        {apt.notes && ` • ${apt.notes}`}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(apt.status)}`}>
                      {APPOINTMENT_STATUS_LABELS[apt.status]}
                    </span>
                    {apt.status === 'scheduled' && (
                      <>
                        <Button variant="ghost" size="icon" title="Complete" onClick={() => handleComplete(apt.id)}>
                          <CheckCircle className="h-4 w-4 text-green-600" />
                        </Button>
                        <Button variant="ghost" size="icon" title="Cancel" onClick={() => handleCancel(apt.id)}>
                          <X className="h-4 w-4 text-red-500" />
                        </Button>
                      </>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
      )}
    </PageTransition>
  )
}
