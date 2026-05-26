import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useInvoiceStore } from '@/store/invoiceStore'
import { useSettingsStore } from '@/store/settingsStore'
import { formatCurrency, formatDate, getStatusColor, rupeesToPaise, getTodayDate } from '@/lib/utils'
import { INVOICE_STATUS_LABELS, PAYMENT_METHODS } from '@/lib/constants'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import InvoicePrintView from '@/components/billing/InvoicePrintView'
import WhatsAppDialog from '@/components/WhatsAppDialog'
import { ArrowLeft, Printer, CreditCard } from 'lucide-react'
import type { WhatsAppMessageResult } from '@/types/api'

export default function InvoiceDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { currentInvoice, isLoading, fetchInvoice, recordPayment } = useInvoiceStore()
  const { clinic, fetchSettings } = useSettingsStore()
  const [showPayment, setShowPayment] = useState(false)
  const [paymentAmount, setPaymentAmount] = useState('')
  const [paymentMethod, setPaymentMethod] = useState('cash')
  const [paymentRef, setPaymentRef] = useState('')
  const [payError, setPayError] = useState('')
  const [whatsAppOpen, setWhatsAppOpen] = useState(false)
  const [whatsAppMessage, setWhatsAppMessage] = useState<WhatsAppMessageResult | null>(null)

  useEffect(() => {
    if (id) fetchInvoice(id)
    if (!clinic) fetchSettings()
  }, [id, fetchInvoice, clinic, fetchSettings])

  const handleRecordPayment = async () => {
    setPayError('')
    const amount = parseFloat(paymentAmount)
    if (!amount || amount <= 0) {
      setPayError('Enter a valid amount')
      return
    }

    try {
      await recordPayment({
        invoiceId: currentInvoice!.id,
        amount: rupeesToPaise(amount),
        method: paymentMethod,
        paymentDate: getTodayDate(),
        reference: paymentRef,
        notes: '',
      })
      const paidMethod = paymentMethod
      setShowPayment(false)
      setPaymentAmount('')
      setPaymentRef('')

      // Invalidate billing list cache so it shows updated status
      queryClient.invalidateQueries({ queryKey: ['invoices'] })

      // Prepare WhatsApp invoice message
      try {
        const templates = await window.go.handler.WhatsAppHandler.GetWhatsAppTemplates()
        if (templates.enabled) {
          const result = await window.go.handler.WhatsAppHandler.PrepareInvoiceMessage(currentInvoice!.id, paidMethod)
          setWhatsAppMessage(result)
          setWhatsAppOpen(true)
        }
      } catch (waErr) {
        console.error('WhatsApp message preparation failed:', waErr)
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to record payment'
      setPayError(message)
    }
  }

  if (isLoading || !currentInvoice) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 animate-pulse rounded-md bg-muted" />
        <div className="h-64 animate-pulse rounded-md bg-muted" />
      </div>
    )
  }

  const invoice = currentInvoice

  return (
    <>
      <div className="space-y-6 print:hidden">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={() => navigate('/billing')}>
              <ArrowLeft className="h-5 w-5" />
            </Button>
            <div>
              <h1 className="text-2xl font-bold">Invoice {invoice.invoiceNumber}</h1>
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(invoice.status)}`}>
                {INVOICE_STATUS_LABELS[invoice.status]}
              </span>
            </div>
          </div>
          <div className="flex gap-2">
            {invoice.status !== 'paid' && invoice.status !== 'void' && (
              <Button onClick={() => setShowPayment(true)}>
                <CreditCard className="h-4 w-4 mr-2" /> Record Payment
              </Button>
            )}
            <Button variant="outline" onClick={() => window.print()}>
              <Printer className="h-4 w-4 mr-2" /> Print
            </Button>
          </div>
        </div>

        {/* Invoice Details */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card>
            <CardHeader><CardTitle className="text-lg">Patient</CardTitle></CardHeader>
            <CardContent>
              <p className="font-medium">{invoice.patient?.name}</p>
              <p className="text-sm text-muted-foreground">{invoice.patient?.phone}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle className="text-lg">Invoice Info</CardTitle></CardHeader>
            <CardContent className="space-y-1">
              <p className="text-sm"><span className="text-muted-foreground">Date:</span> {formatDate(invoice.invoiceDate)}</p>
              <p className="text-sm"><span className="text-muted-foreground">Number:</span> {invoice.invoiceNumber}</p>
            </CardContent>
          </Card>
        </div>

        {/* Line Items */}
        <Card>
          <CardHeader><CardTitle className="text-lg">Items</CardTitle></CardHeader>
          <CardContent>
            <table className="w-full">
              <thead className="border-b">
                <tr>
                  <th className="pb-2 text-left text-sm font-medium">Description</th>
                  <th className="pb-2 text-center text-sm font-medium">Qty</th>
                  <th className="pb-2 text-right text-sm font-medium">Price</th>
                  <th className="pb-2 text-right text-sm font-medium">Amount</th>
                </tr>
              </thead>
              <tbody>
                {invoice.items?.map((item) => (
                  <tr key={item.id} className="border-b">
                    <td className="py-2 text-sm">
                      {item.description}
                      {item.toothNumber && <span className="text-muted-foreground"> (Tooth: {item.toothNumber})</span>}
                    </td>
                    <td className="py-2 text-sm text-center">{item.quantity}</td>
                    <td className="py-2 text-sm text-right">{formatCurrency(item.unitPrice)}</td>
                    <td className="py-2 text-sm text-right">{formatCurrency(item.amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div className="mt-4 border-t pt-4 space-y-1 text-sm">
              <div className="flex justify-between"><span>Subtotal</span><span>{formatCurrency(invoice.subTotal)}</span></div>
              {invoice.discountAmount > 0 && (
                <div className="flex justify-between text-red-600"><span>Discount</span><span>-{formatCurrency(invoice.discountAmount)}</span></div>
              )}
              {invoice.cgstAmount > 0 && (
                <>
                  <div className="flex justify-between"><span>CGST</span><span>{formatCurrency(invoice.cgstAmount)}</span></div>
                  <div className="flex justify-between"><span>SGST</span><span>{formatCurrency(invoice.sgstAmount)}</span></div>
                </>
              )}
              <div className="flex justify-between font-bold text-lg border-t pt-2">
                <span>Total</span><span>{formatCurrency(invoice.totalAmount)}</span>
              </div>
              <div className="flex justify-between text-green-600"><span>Paid</span><span>{formatCurrency(invoice.paidAmount)}</span></div>
              {invoice.balanceAmount > 0 && (
                <div className="flex justify-between text-red-600 font-medium"><span>Balance Due</span><span>{formatCurrency(invoice.balanceAmount)}</span></div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Payment History */}
        {invoice.payments && invoice.payments.length > 0 && (
          <Card>
            <CardHeader><CardTitle className="text-lg">Payment History</CardTitle></CardHeader>
            <CardContent>
              <div className="space-y-2">
                {invoice.payments.map((payment) => (
                  <div key={payment.id} className="flex justify-between items-center p-3 bg-muted rounded-md">
                    <div>
                      <p className="text-sm font-medium">{formatCurrency(payment.amount)}</p>
                      <p className="text-xs text-muted-foreground capitalize">{payment.method} • {formatDate(payment.paymentDate)}</p>
                    </div>
                    {payment.reference && <span className="text-xs text-muted-foreground">Ref: {payment.reference}</span>}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Record Payment Dialog */}
        {showPayment && (
          <Card className="border-primary">
            <CardHeader><CardTitle className="text-lg">Record Payment</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              {payError && <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">{payError}</div>}
              <p className="text-sm text-muted-foreground">Balance: {formatCurrency(invoice.balanceAmount)}</p>
              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>Amount (₹) *</Label>
                  <Input
                    type="number"
                    step="0.01"
                    value={paymentAmount}
                    onChange={(e) => setPaymentAmount(e.target.value)}
                    placeholder="Amount in rupees"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Method</Label>
                  <select
                    value={paymentMethod}
                    onChange={(e) => setPaymentMethod(e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    {PAYMENT_METHODS.map(m => <option key={m.value} value={m.value}>{m.label}</option>)}
                  </select>
                </div>
                <div className="space-y-2">
                  <Label>Reference</Label>
                  <Input value={paymentRef} onChange={(e) => setPaymentRef(e.target.value)} placeholder="UPI ref / Card last 4" />
                </div>
              </div>
              <div className="flex gap-2">
                <Button onClick={handleRecordPayment}>Save Payment</Button>
                <Button variant="outline" onClick={() => setShowPayment(false)}>Cancel</Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Print View: Always in DOM but only visible during print */}
      <InvoicePrintView invoice={invoice} clinic={clinic} />

      {/* WhatsApp Dialog */}
      <WhatsAppDialog
        isOpen={whatsAppOpen}
        onClose={() => { setWhatsAppOpen(false); setWhatsAppMessage(null) }}
        messageResult={whatsAppMessage}
        title="Send Payment Receipt"
      />
    </>
  )
}
