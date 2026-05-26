import { z } from 'zod'

export const loginSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
})

export const setupSchema = z.object({
  clinicName: z.string().min(2, 'Clinic name is required'),
  doctorName: z.string().min(2, 'Doctor name is required'),
  doctorQualification: z.string().optional(),
  address: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  pincode: z.string().optional(),
  phone: z.string().min(10, 'Valid phone number required'),
  email: z.string().email().optional().or(z.literal('')),
  gstin: z.string().optional(),
  gstEnabled: z.boolean(),
  invoicePrefix: z.string().min(1).max(5).optional().or(z.literal('')),
  adminUsername: z.string().min(3, 'Username must be at least 3 characters'),
  adminPassword: z.string().min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one digit'),
  adminFullName: z.string().min(2, 'Full name is required'),
})

export const patientSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters').max(100),
  phone: z.string().min(10, 'Valid 10-digit phone number required'),
  email: z.string().email().optional().or(z.literal('')),
  gender: z.enum(['male', 'female', 'other']),
  age: z.number().min(0).max(120).optional().nullable().or(z.literal('')),
  dateOfBirth: z.string().optional(),
  address: z.string().optional(),
  city: z.string().optional(),
  bloodGroup: z.string().optional(),
  medicalHistory: z.string().optional(),
  allergies: z.string().optional(),
  notes: z.string().optional(),
})

export const appointmentSchema = z.object({
  patientId: z.string().min(1, 'Patient is required'),
  date: z.string().min(1, 'Date is required'),
  startTime: z.string().min(1, 'Start time is required'),
  endTime: z.string().min(1, 'End time is required'),
  duration: z.number().min(15).max(180).optional().default(30),
  purpose: z.string().optional(),
  notes: z.string().optional(),
})

export const invoiceItemSchema = z.object({
  treatmentId: z.string().optional(),
  description: z.string().min(1, 'Description is required'),
  quantity: z.number().min(1, 'Quantity must be at least 1'),
  unitPrice: z.number().min(1, 'Price must be greater than 0'), // in rupees for input
  toothNumber: z.string().optional(),
})

export const paymentSchema = z.object({
  amount: z.number().min(1, 'Amount must be greater than 0'), // in rupees for input
  method: z.enum(['cash', 'upi', 'card', 'bank_transfer', 'other']),
  paymentDate: z.string().min(1, 'Payment date is required'),
  reference: z.string().optional(),
  notes: z.string().optional(),
})

export const changePasswordSchema = z.object({
  oldPassword: z.string().min(1, 'Current password is required'),
  newPassword: z.string().min(8, 'New password must be at least 8 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one digit'),
  confirmPassword: z.string().min(8),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ['confirmPassword'],
})

export type LoginFormData = z.infer<typeof loginSchema>
export type SetupFormData = z.infer<typeof setupSchema>
export type PatientFormData = z.infer<typeof patientSchema>
export type AppointmentFormData = z.infer<typeof appointmentSchema>
export type InvoiceItemFormData = z.infer<typeof invoiceItemSchema>
export type PaymentFormData = z.infer<typeof paymentSchema>
export type ChangePasswordFormData = z.infer<typeof changePasswordSchema>
