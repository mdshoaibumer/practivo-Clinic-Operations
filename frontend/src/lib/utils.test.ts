import { describe, it, expect } from 'vitest'
import {
  formatCurrency,
  formatDate,
  formatTime,
  rupeesToPaise,
  paiseToRupees,
  getStatusColor,
} from './utils'

describe('formatCurrency', () => {
  it('formats paise to INR currency string', () => {
    expect(formatCurrency(150000)).toBe('₹1,500.00')
  })

  it('handles zero', () => {
    expect(formatCurrency(0)).toBe('₹0.00')
  })

  it('handles small amounts', () => {
    expect(formatCurrency(50)).toBe('₹0.50')
  })

  it('formats large amounts with Indian grouping', () => {
    // ₹1,50,000.00 (Indian numbering system)
    const result = formatCurrency(15000000)
    expect(result).toContain('1,50,000')
  })
})

describe('formatDate', () => {
  it('formats ISO date string', () => {
    const result = formatDate('2026-05-16')
    expect(result).toContain('16')
    expect(result).toContain('May')
    expect(result).toContain('2026')
  })

  it('returns empty string for empty input', () => {
    expect(formatDate('')).toBe('')
  })
})

describe('formatTime', () => {
  it('formats 24h time to 12h AM', () => {
    expect(formatTime('09:30')).toBe('9:30 AM')
  })

  it('formats 24h time to 12h PM', () => {
    expect(formatTime('14:00')).toBe('2:00 PM')
  })

  it('handles noon', () => {
    expect(formatTime('12:00')).toBe('12:00 PM')
  })

  it('handles midnight', () => {
    expect(formatTime('00:00')).toBe('12:00 AM')
  })

  it('returns empty string for empty input', () => {
    expect(formatTime('')).toBe('')
  })
})

describe('rupeesToPaise', () => {
  it('converts whole rupees', () => {
    expect(rupeesToPaise(100)).toBe(10000)
  })

  it('converts rupees with decimal', () => {
    expect(rupeesToPaise(15.50)).toBe(1550)
  })

  it('rounds floating point correctly', () => {
    // 19.99 * 100 can have floating point issues
    expect(rupeesToPaise(19.99)).toBe(1999)
  })

  it('handles zero', () => {
    expect(rupeesToPaise(0)).toBe(0)
  })
})

describe('paiseToRupees', () => {
  it('converts paise to rupees', () => {
    expect(paiseToRupees(10000)).toBe(100)
  })

  it('converts partial rupees', () => {
    expect(paiseToRupees(1550)).toBe(15.5)
  })

  it('handles zero', () => {
    expect(paiseToRupees(0)).toBe(0)
  })
})

describe('getStatusColor', () => {
  it('returns green for paid', () => {
    expect(getStatusColor('paid')).toContain('green')
  })

  it('returns yellow for partial', () => {
    expect(getStatusColor('partial')).toContain('yellow')
  })

  it('returns blue for issued', () => {
    expect(getStatusColor('issued')).toContain('blue')
  })

  it('returns red for void', () => {
    expect(getStatusColor('void')).toContain('red')
  })

  it('returns gray for unknown status', () => {
    expect(getStatusColor('unknown')).toContain('gray')
  })
})
