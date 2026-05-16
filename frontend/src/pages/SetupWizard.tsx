import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useSettingsStore } from '@/store/settingsStore'
import { setupSchema, type SetupFormData } from '@/lib/validators'
import { INDIAN_STATES, INDIAN_CITIES } from '@/lib/constants'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'

export default function SetupWizard() {
  const navigate = useNavigate()
  const { completeSetup } = useSettingsStore()
  const [step, setStep] = useState(1)
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const { register, handleSubmit, formState: { errors }, watch, control, trigger } = useForm<SetupFormData>({
    resolver: zodResolver(setupSchema),
    defaultValues: {
      gstEnabled: false,
      invoicePrefix: 'CD',
    },
  })

  const nextStep = async () => {
    const fieldsToValidate = step === 1 
      ? ['clinicName', 'doctorName', 'phone'] as const
      : ['adminFullName', 'adminUsername', 'adminPassword'] as const
    
    const isValid = await trigger(fieldsToValidate)
    if (isValid) {
      setStep(step + 1)
    }
  }

  const onSubmit = async (data: SetupFormData) => {
    setError('')
    setIsSubmitting(true)
    try {
      await completeSetup({
        clinicName: data.clinicName,
        doctorName: data.doctorName,
        address: data.address || '',
        city: data.city || '',
        state: data.state || '',
        pincode: data.pincode || '',
        phone: data.phone,
        email: data.email || '',
        gstin: data.gstin || '',
        gstEnabled: data.gstEnabled,
        invoicePrefix: data.invoicePrefix || 'PV',
        adminUsername: data.adminUsername,
        adminPassword: data.adminPassword,
        adminFullName: data.adminFullName,
      })
      navigate('/login')
    } catch (err: any) {
      const message = err.message || 'Setup failed'
      setError(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Card className="w-full max-w-lg mx-auto">
      <CardHeader>
        <CardTitle>Setup Clinmitra Dental</CardTitle>
        <CardDescription>Step {step} of 3 — {step === 1 ? 'Clinic Details' : step === 2 ? 'Admin Account' : 'Confirm'}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">{error}</div>
          )}

          {step === 1 && (
            <>
              <div className="space-y-2">
                <Label htmlFor="clinicName">Clinic Name *</Label>
                <Input id="clinicName" {...register('clinicName')} placeholder="e.g., Smile Dental Clinic" />
                {errors.clinicName && <p className="text-sm text-red-500">{errors.clinicName.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="doctorName">Doctor Name *</Label>
                <Input id="doctorName" {...register('doctorName')} placeholder="e.g., Dr. Sharma" />
                {errors.doctorName && <p className="text-sm text-red-500">{errors.doctorName.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="phone">Phone *</Label>
                <Input id="phone" {...register('phone')} placeholder="10-digit mobile number" />
                {errors.phone && <p className="text-sm text-red-500">{errors.phone.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="address">Address</Label>
                <Input id="address" {...register('address')} placeholder="Clinic address" />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="state">State</Label>
                  <Controller
                    name="state"
                    control={control}
                    render={({ field }) => (
                      <select
                        id="state"
                        value={field.value || ''}
                        onChange={field.onChange}
                        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      >
                        <option value="">Select state...</option>
                        {INDIAN_STATES.map(s => (
                          <option key={s} value={s}>{s}</option>
                        ))}
                      </select>
                    )}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="city">City</Label>
                  <Controller
                    name="city"
                    control={control}
                    render={({ field }) => {
                      const selectedState = watch('state')
                      const cities = selectedState ? (INDIAN_CITIES[selectedState] || []) : []
                      return (
                        <select
                          id="city"
                          value={field.value || ''}
                          onChange={field.onChange}
                          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                        >
                          <option value="">Select city...</option>
                          {cities.map(c => (
                            <option key={c} value={c}>{c}</option>
                          ))}
                        </select>
                      )
                    }}
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" type="email" {...register('email')} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="gstin">GSTIN</Label>
                  <Input id="gstin" {...register('gstin')} placeholder="Optional" />
                </div>
              </div>
              <Button type="button" className="w-full" onClick={nextStep}>Next</Button>
            </>
          )}

          {step === 2 && (
            <>
              <p className="text-sm text-muted-foreground">Create your admin login credentials.</p>
              <div className="space-y-2">
                <Label htmlFor="adminFullName">Full Name *</Label>
                <Input id="adminFullName" {...register('adminFullName')} placeholder="Your full name" />
                {errors.adminFullName && <p className="text-sm text-red-500">{errors.adminFullName.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="adminUsername">Username *</Label>
                <Input id="adminUsername" {...register('adminUsername')} placeholder="Login username" />
                {errors.adminUsername && <p className="text-sm text-red-500">{errors.adminUsername.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="adminPassword">Password *</Label>
                <Input id="adminPassword" type="password" {...register('adminPassword')} placeholder="Min 6 characters" />
                {errors.adminPassword && <p className="text-sm text-red-500">{errors.adminPassword.message}</p>}
              </div>
              <div className="flex gap-2">
                <Button type="button" variant="outline" onClick={() => setStep(1)}>Back</Button>
                <Button type="button" className="flex-1" onClick={nextStep}>Next</Button>
              </div>
            </>
          )}

          {step === 3 && (
            <>
              <p className="text-sm text-muted-foreground">Review and confirm your setup.</p>
              <div className="bg-muted p-4 rounded-md space-y-2 text-sm">
                <p><strong>Clinic:</strong> Will be configured with your details</p>
                <p><strong>Admin:</strong> Account will be created for login</p>
                <p><strong>Treatments:</strong> Default dental treatment catalog will be loaded</p>
                <p><strong>Backup:</strong> Auto-backup enabled by default</p>
              </div>
              <div className="flex gap-2">
                <Button type="button" variant="outline" onClick={() => setStep(2)}>Back</Button>
                <Button type="submit" className="flex-1" disabled={isSubmitting}>
                  {isSubmitting ? 'Setting up...' : 'Complete Setup'}
                </Button>
              </div>
            </>
          )}
        </form>
      </CardContent>
    </Card>
  )
}
