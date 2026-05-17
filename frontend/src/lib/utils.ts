import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Format paise to rupee display string
export function formatCurrency(paise: number): string {
  const rupees = paise / 100
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: 2,
  }).format(rupees)
}

// Format date for display (handles string ISO dates or number timestamps)
export function formatDate(dateVal: string | number | undefined): string {
  if (!dateVal) return '-'
  
  let date: Date
  if (typeof dateVal === 'number') {
    // If the timestamp is in seconds (Go default), convert to milliseconds
    // 10^12 is approx year 2001 in milliseconds, so anything less is likely seconds
    const ms = dateVal < 10000000000 ? dateVal * 1000 : dateVal
    date = new Date(ms)
  } else {
    date = new Date(dateVal)
  }

  if (isNaN(date.getTime())) return '-'

  return date.toLocaleDateString('en-IN', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

// Format time for display
export function formatTime(timeStr: string): string {
  if (!timeStr) return ''
  const [hours, minutes] = timeStr.split(':')
  const h = parseInt(hours)
  const ampm = h >= 12 ? 'PM' : 'AM'
  const displayHour = h % 12 || 12
  return `${displayHour}:${minutes} ${ampm}`
}

// Get today's date in YYYY-MM-DD format
export function getTodayDate(): string {
  return new Date().toISOString().split('T')[0]
}

// Convert rupees (input) to paise for storage
export function rupeesToPaise(rupees: number): number {
  return Math.round(rupees * 100)
}

// Convert paise to rupees for display in inputs
export function paiseToRupees(paise: number): number {
  return paise / 100
}

// Invoice status badge color
export function getStatusColor(status: string): string {
  switch (status) {
    case 'paid': return 'bg-green-100 text-green-800'
    case 'partial': return 'bg-yellow-100 text-yellow-800'
    case 'issued': return 'bg-blue-100 text-blue-800'
    case 'void': return 'bg-red-100 text-red-800'
    case 'scheduled': return 'bg-blue-100 text-blue-800'
    case 'completed': return 'bg-green-100 text-green-800'
    case 'cancelled': return 'bg-red-100 text-red-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}
