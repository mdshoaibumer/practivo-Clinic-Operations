import { describe, it, expect } from 'vitest'
import {
  loginSchema,
  patientSchema,
  appointmentSchema,
  changePasswordSchema,
  paymentSchema,
} from './validators'

describe('loginSchema', () => {
  it('accepts valid credentials', () => {
    const result = loginSchema.safeParse({ username: 'admin', password: 'secret123' })
    expect(result.success).toBe(true)
  })

  it('rejects short username', () => {
    const result = loginSchema.safeParse({ username: 'ab', password: 'secret123' })
    expect(result.success).toBe(false)
  })

  it('rejects short password', () => {
    const result = loginSchema.safeParse({ username: 'admin', password: '12345' })
    expect(result.success).toBe(false)
  })

  it('rejects empty fields', () => {
    const result = loginSchema.safeParse({ username: '', password: '' })
    expect(result.success).toBe(false)
  })
})

describe('patientSchema', () => {
  it('accepts valid patient data', () => {
    const result = patientSchema.safeParse({
      name: 'Ramesh Kumar',
      phone: '9876543210',
      gender: 'male',
    })
    expect(result.success).toBe(true)
  })

  it('rejects name too short', () => {
    const result = patientSchema.safeParse({
      name: 'R',
      phone: '9876543210',
      gender: 'male',
    })
    expect(result.success).toBe(false)
  })

  it('rejects phone too short', () => {
    const result = patientSchema.safeParse({
      name: 'Ramesh Kumar',
      phone: '12345',
      gender: 'male',
    })
    expect(result.success).toBe(false)
  })

  it('rejects invalid gender', () => {
    const result = patientSchema.safeParse({
      name: 'Ramesh Kumar',
      phone: '9876543210',
      gender: 'invalid',
    })
    expect(result.success).toBe(false)
  })

  it('accepts optional email', () => {
    const result = patientSchema.safeParse({
      name: 'Ramesh Kumar',
      phone: '9876543210',
      gender: 'female',
      email: '',
    })
    expect(result.success).toBe(true)
  })

  it('rejects invalid email format', () => {
    const result = patientSchema.safeParse({
      name: 'Ramesh Kumar',
      phone: '9876543210',
      gender: 'male',
      email: 'not-an-email',
    })
    expect(result.success).toBe(false)
  })
})

describe('appointmentSchema', () => {
  it('accepts valid appointment', () => {
    const result = appointmentSchema.safeParse({
      patientId: 'uuid-123',
      date: '2026-05-16',
      startTime: '10:00',
      endTime: '10:30',
    })
    expect(result.success).toBe(true)
  })

  it('rejects missing patient', () => {
    const result = appointmentSchema.safeParse({
      patientId: '',
      date: '2026-05-16',
      startTime: '10:00',
      endTime: '10:30',
    })
    expect(result.success).toBe(false)
  })

  it('rejects missing date', () => {
    const result = appointmentSchema.safeParse({
      patientId: 'uuid-123',
      date: '',
      startTime: '10:00',
      endTime: '10:30',
    })
    expect(result.success).toBe(false)
  })

  it('defaults duration to 30', () => {
    const result = appointmentSchema.safeParse({
      patientId: 'uuid-123',
      date: '2026-05-16',
      startTime: '10:00',
      endTime: '10:30',
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.duration).toBe(30)
    }
  })
})

describe('paymentSchema', () => {
  it('accepts valid payment', () => {
    const result = paymentSchema.safeParse({
      amount: 500,
      method: 'cash',
      paymentDate: '2026-05-16',
    })
    expect(result.success).toBe(true)
  })

  it('rejects zero amount', () => {
    const result = paymentSchema.safeParse({
      amount: 0,
      method: 'cash',
      paymentDate: '2026-05-16',
    })
    expect(result.success).toBe(false)
  })

  it('rejects invalid payment method', () => {
    const result = paymentSchema.safeParse({
      amount: 500,
      method: 'bitcoin',
      paymentDate: '2026-05-16',
    })
    expect(result.success).toBe(false)
  })
})

describe('changePasswordSchema', () => {
  it('accepts matching passwords', () => {
    const result = changePasswordSchema.safeParse({
      oldPassword: 'current123',
      newPassword: 'newpass123',
      confirmPassword: 'newpass123',
    })
    expect(result.success).toBe(true)
  })

  it('rejects mismatched passwords', () => {
    const result = changePasswordSchema.safeParse({
      oldPassword: 'current123',
      newPassword: 'newpass123',
      confirmPassword: 'different',
    })
    expect(result.success).toBe(false)
  })

  it('rejects short new password', () => {
    const result = changePasswordSchema.safeParse({
      oldPassword: 'current123',
      newPassword: '12345',
      confirmPassword: '12345',
    })
    expect(result.success).toBe(false)
  })
})
