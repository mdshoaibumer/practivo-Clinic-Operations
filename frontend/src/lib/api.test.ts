import { describe, it, expect } from 'vitest'
import { parseError, apiCall, getErrorMessage } from './api'

describe('parseError', () => {
  it('parses backend AppError format [CODE] Message', () => {
    const result = parseError('[VALIDATION_ERROR] Name is required')
    expect(result.code).toBe('VALIDATION_ERROR')
    expect(result.message).toBe('Name is required')
  })

  it('parses NOT_FOUND error', () => {
    const result = parseError('[NOT_FOUND] Patient not found')
    expect(result.code).toBe('NOT_FOUND')
    expect(result.message).toBe('Patient not found')
  })

  it('handles plain string errors', () => {
    const result = parseError('Something went wrong')
    expect(result.code).toBe('UNKNOWN')
    expect(result.message).toBe('Something went wrong')
  })

  it('handles null/undefined', () => {
    expect(parseError(null).code).toBe('UNKNOWN')
    expect(parseError(undefined).code).toBe('UNKNOWN')
  })

  it('handles Error objects via String()', () => {
    const result = parseError(new Error('test error'))
    expect(result.code).toBe('UNKNOWN')
    expect(result.message).toContain('test error')
  })

  it('handles empty string', () => {
    const result = parseError('')
    expect(result.code).toBe('UNKNOWN')
    expect(result.message).toBe('An unexpected error occurred')
  })

  it('parses UNAUTHORIZED error', () => {
    const result = parseError('[UNAUTHORIZED] Session expired')
    expect(result.code).toBe('UNAUTHORIZED')
    expect(result.message).toBe('Session expired')
  })

  it('parses ACCOUNT_LOCKED error', () => {
    const result = parseError('[ACCOUNT_LOCKED] Too many login attempts')
    expect(result.code).toBe('ACCOUNT_LOCKED')
    expect(result.message).toBe('Too many login attempts')
  })
})

describe('apiCall', () => {
  it('returns [result, null] on success', async () => {
    const [result, err] = await apiCall(() => Promise.resolve({ id: '1', name: 'Test' }))
    expect(err).toBeNull()
    expect(result).toEqual({ id: '1', name: 'Test' })
  })

  it('returns [null, AppError] on failure', async () => {
    const [result, err] = await apiCall(() => Promise.reject('[NOT_FOUND] Item not found'))
    expect(result).toBeNull()
    expect(err).not.toBeNull()
    expect(err!.code).toBe('NOT_FOUND')
    expect(err!.message).toBe('Item not found')
  })

  it('handles async function that throws', async () => {
    const [result, err] = await apiCall(async () => {
      throw new Error('network timeout')
    })
    expect(result).toBeNull()
    expect(err!.code).toBe('UNKNOWN')
  })
})

describe('getErrorMessage', () => {
  it('returns login prompt for UNAUTHORIZED', () => {
    expect(getErrorMessage({ code: 'UNAUTHORIZED', message: '' })).toBe('Please log in to continue')
  })

  it('returns permission message for FORBIDDEN', () => {
    expect(getErrorMessage({ code: 'FORBIDDEN', message: '' })).toBe(
      'You do not have permission to perform this action'
    )
  })

  it('returns not found message', () => {
    expect(getErrorMessage({ code: 'NOT_FOUND', message: '' })).toBe(
      'The requested resource was not found'
    )
  })

  it('returns account locked message', () => {
    expect(getErrorMessage({ code: 'ACCOUNT_LOCKED', message: '' })).toBe(
      'Account is temporarily locked. Please try again later.'
    )
  })

  it('passes through validation error message', () => {
    expect(getErrorMessage({ code: 'VALIDATION_ERROR', message: 'Phone is required' })).toBe(
      'Phone is required'
    )
  })

  it('returns generic message for INTERNAL_ERROR', () => {
    expect(getErrorMessage({ code: 'INTERNAL_ERROR', message: 'sql: connection reset' })).toBe(
      'Something went wrong. Please try again.'
    )
  })

  it('falls back to message for unknown codes', () => {
    expect(getErrorMessage({ code: 'CUSTOM', message: 'Custom message' })).toBe('Custom message')
  })
})
