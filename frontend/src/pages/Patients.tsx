import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { patientSchema, type PatientFormData } from '@/lib/validators'
import { GENDER_OPTIONS, BLOOD_GROUP_OPTIONS, INDIAN_STATES, INDIAN_CITIES, PAGE_SIZE } from '@/lib/constants'
import { useDebounce } from '@/lib/useDebounce'
import { exportToCSV } from '@/lib/exportCSV'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import PatientAvatar from '@/components/PatientAvatar'
import { Plus, Search, Download } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@/components/ui/use-toast'
import PageTransition from '@/components/PageTransition'

export default function Patients() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { toast } = useToast()
  const queryClient = useQueryClient()
  
  const [showForm, setShowForm] = useState(false)
  const [formError, setFormError] = useState('')
  const [localSearch, setLocalSearch] = useState('')
  const debouncedSearch = useDebounce(localSearch, 300)
  const [page, setPage] = useState(1)
  const [selectedState, setSelectedState] = useState('')

  const { register, handleSubmit, formState: { errors }, reset } = useForm<PatientFormData>({
    resolver: zodResolver(patientSchema),
    defaultValues: { gender: 'male' },
  })

  // React Query for patients list
  const { data: response, isLoading, isError } = useQuery({
    queryKey: ['patients', page, debouncedSearch],
    queryFn: () => window.go.handler.PatientHandler.ListPatients(page, PAGE_SIZE, debouncedSearch),
  })

  const patients = response?.patients || []
  const totalCount = response?.total || 0

  useEffect(() => {
    if (isError) {
      toast({
        variant: "destructive",
        title: "Error fetching patients",
        description: "Could not load patient data.",
      })
    }
  }, [isError, toast])

  // Mutation for creating a patient
  const createPatientMutation = useMutation({
    mutationFn: (data: any) => window.go.handler.PatientHandler.CreatePatient(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients'] })
      queryClient.invalidateQueries({ queryKey: ['dashboardStats'] })
      reset()
      setShowForm(false)
      toast({
        title: "Patient Created",
        description: "Patient registered successfully.",
      })
    },
    onError: (err: any) => {
      const message = typeof err === 'string' ? err : err.message || 'Failed to create patient'
      setFormError(message)
    }
  })

  // Handle action=new from keyboard shortcut or global search
  useEffect(() => {
    if (searchParams.get('action') === 'new') {
      setShowForm(true)
    }
  }, [searchParams])

  // Reset page when search changes
  useEffect(() => {
    setPage(1)
  }, [debouncedSearch])

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

  const onSubmit = (data: PatientFormData) => {
    setFormError('')
    createPatientMutation.mutate({
      name: data.name,
      phone: data.phone,
      email: data.email || '',
      gender: data.gender,
      age: data.age || 0,
      dateOfBirth: data.dateOfBirth || '',
      address: data.address || '',
      city: data.city || '',
      bloodGroup: data.bloodGroup || '',
      medicalHistory: data.medicalHistory || '',
      allergies: data.allergies || '',
      notes: data.notes || '',
    })
  }

  const handleExportCSV = () => {
    if (patients.length === 0) return
    exportToCSV(
      patients.map(p => ({
        name: p.name,
        phone: p.phone,
        gender: p.gender,
        age: p.age || '',
        email: p.email || '',
        city: p.city || '',
        bloodGroup: p.bloodGroup || '',
        medicalHistory: p.medicalHistory || '',
        allergies: p.allergies || '',
      })),
      `patients-export-${new Date().toISOString().split('T')[0]}`,
      [
        { key: 'name', header: 'Name' },
        { key: 'phone', header: 'Phone' },
        { key: 'gender', header: 'Gender' },
        { key: 'age', header: 'Age' },
        { key: 'email', header: 'Email' },
        { key: 'city', header: 'City' },
        { key: 'bloodGroup', header: 'Blood Group' },
        { key: 'medicalHistory', header: 'Medical History' },
        { key: 'allergies', header: 'Allergies' },
      ]
    )
  }

  return (
    <PageTransition className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Patients</h1>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={handleExportCSV} disabled={patients.length === 0}>
            <Download className="h-4 w-4 mr-2" /> Export CSV
          </Button>
          <Button onClick={() => setShowForm(!showForm)}>
            <Plus className="h-4 w-4 mr-2" /> New Patient
          </Button>
        </div>
      </div>

      {/* New Patient Form - Improved 2-column layout with logical grouping */}
      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle>Register New Patient</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {formError && (
                <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">{formError}</div>
              )}

              {/* Personal Details */}
              <div>
                <h3 className="text-sm font-medium text-muted-foreground mb-3">Personal Details</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Name *</Label>
                    <Input {...register('name')} placeholder="Patient full name" autoFocus />
                    {errors.name && <p className="text-sm text-red-500">{errors.name.message}</p>}
                  </div>
                  <div className="space-y-2">
                    <Label>Phone *</Label>
                    <Input {...register('phone')} placeholder="10-digit mobile" />
                    {errors.phone && <p className="text-sm text-red-500">{errors.phone.message}</p>}
                  </div>
                  <div className="space-y-2">
                    <Label>Gender *</Label>
                    <select {...register('gender')} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                      {GENDER_OPTIONS.map(opt => (
                        <option key={opt.value} value={opt.value}>{opt.label}</option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-2">
                    <Label>Age</Label>
                    <Input type="number" {...register('age', { valueAsNumber: true })} />
                  </div>
                </div>
              </div>

              {/* Contact Details */}
              <div>
                <h3 className="text-sm font-medium text-muted-foreground mb-3">Contact Details</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Email</Label>
                    <Input type="email" {...register('email')} />
                  </div>
                  <div className="space-y-2">
                    <Label>State</Label>
                    <select
                      value={selectedState}
                      onChange={(e) => setSelectedState(e.target.value)}
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    >
                      <option value="">Select state...</option>
                      {INDIAN_STATES.map(s => (
                        <option key={s} value={s}>{s}</option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-2">
                    <Label>City</Label>
                    <select {...register('city')} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                      <option value="">Select city...</option>
                      {(selectedState ? (INDIAN_CITIES[selectedState] || []) : Object.values(INDIAN_CITIES).flat()).map(c => (
                        <option key={c} value={c}>{c}</option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-2">
                    <Label>Blood Group</Label>
                    <select {...register('bloodGroup')} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                      {BLOOD_GROUP_OPTIONS.map(opt => (
                        <option key={opt.value} value={opt.value}>{opt.label}</option>
                      ))}
                    </select>
                  </div>
                </div>
              </div>

              {/* Medical Information */}
              <div>
                <h3 className="text-sm font-medium text-muted-foreground mb-3">Medical Information</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Medical History</Label>
                    <Input {...register('medicalHistory')} placeholder="Any known conditions" />
                  </div>
                  <div className="space-y-2">
                    <Label>Allergies</Label>
                    <Input {...register('allergies')} placeholder="Known allergies" />
                  </div>
                </div>
              </div>

              <div className="flex gap-2 pt-2">
                <Button type="submit" disabled={createPatientMutation.isPending}>
                  {createPatientMutation.isPending ? 'Saving...' : 'Save Patient'}
                </Button>
                <Button type="button" variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
                <span className="text-xs text-muted-foreground self-center ml-2">Press Esc to close</span>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
        <Input
          data-search-input
          placeholder="Search by name or phone..."
          value={localSearch}
          onChange={(e) => setLocalSearch(e.target.value)}
          className="pl-10"
        />
      </div>

      {/* Patient List */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 text-center text-muted-foreground">Loading...</div>
          ) : patients.length === 0 ? (
            <div className="p-6 text-center text-muted-foreground">
              {debouncedSearch ? 'No patients found matching your search.' : 'No patients registered yet. Click "New Patient" to get started.'}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="border-b bg-muted/50">
                  <tr>
                    <th className="px-4 py-3 text-left text-sm font-medium">Name</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Phone</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Gender</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Age</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">City</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {patients.map((patient: any) => (
                    <tr key={patient.id} className="border-b hover:bg-muted/30 cursor-pointer" onClick={() => navigate(`/patients/${patient.id}`)}>
                      <td className="px-4 py-3 text-sm">
                        <div className="flex items-center gap-3">
                          <PatientAvatar name={patient.name} size="sm" />
                          <span className="font-medium">{patient.name}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm">{patient.phone}</td>
                      <td className="px-4 py-3 text-sm capitalize">{patient.gender}</td>
                      <td className="px-4 py-3 text-sm">{patient.age || '-'}</td>
                      <td className="px-4 py-3 text-sm">{patient.city || '-'}</td>
                      <td className="px-4 py-3 text-sm">
                        <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); navigate(`/patients/${patient.id}`) }}>
                          View
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalCount > PAGE_SIZE && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * PAGE_SIZE + 1} to {Math.min(page * PAGE_SIZE, totalCount)} of {totalCount}
          </p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
            <Button variant="outline" size="sm" disabled={page * PAGE_SIZE >= totalCount} onClick={() => setPage(page + 1)}>Next</Button>
          </div>
        </div>
      )}
    </PageTransition>
  )
}
