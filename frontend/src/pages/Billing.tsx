import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useSettingsStore } from '@/store/settingsStore'
import { usePatientStore } from '@/store/patientStore'
import { formatCurrency, getStatusColor, formatDate, rupeesToPaise } from '@/lib/utils'
import { INVOICE_STATUS_LABELS, PAGE_SIZE } from '@/lib/constants'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import InvoiceConfirmDialog from '@/components/billing/InvoiceConfirmDialog'
import { Plus, Trash2 } from 'lucide-react'
import type { Treatment } from '@/types/models'
import type { InvoiceItemInput } from '@/types/api'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@/components/ui/use-toast'
import PageTransition from '@/components/PageTransition'

export default function Billing() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { toast } = useToast()
  const queryClient = useQueryClient()
  
  const { treatments, fetchTreatments } = useSettingsStore()
  const { patients, fetchPatients } = usePatientStore()
  const [showForm, setShowForm] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [selectedPatient, setSelectedPatient] = useState('')
  const [items, setItems] = useState<(InvoiceItemInput & { key: string })[]>([])
  const [discount, setDiscount] = useState('')
  const [formError, setFormError] = useState('')
  
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')

  // Fetch invoices using React Query
  const { data: response, isLoading, isError } = useQuery({
    queryKey: ['invoices', page, statusFilter],
    queryFn: () => window.go.handler.InvoiceHandler.ListInvoices(page, PAGE_SIZE, statusFilter, "", "", "", ""),
  })

  const invoices = response?.invoices || []
  const totalCount = response?.total || 0

  useEffect(() => {
    if (isError) {
      toast({
        variant: "destructive",
        title: "Error fetching invoices",
        description: "Could not load invoice data.",
      })
    }
  }, [isError, toast])

  // Mutation for creating invoice
  const createInvoiceMutation = useMutation({
    mutationFn: (data: any) => window.go.handler.InvoiceHandler.CreateInvoice(data),
    onSuccess: (invoice) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] })
      queryClient.invalidateQueries({ queryKey: ['dashboardStats'] })
      setShowForm(false)
      setShowConfirm(false)
      setItems([])
      setSelectedPatient('')
      setDiscount('')
      navigate(`/billing/${invoice.id}`)
      toast({
        title: "Invoice Created",
        description: "Invoice has been generated.",
      })
    },
    onError: (err: any) => {
      const message = typeof err === 'string' ? err : err.message || 'Failed to create invoice'
      setFormError(message)
      setShowConfirm(false)
    }
  })

  // Handle action=new from keyboard shortcut or global search
  useEffect(() => {
    if (searchParams.get('action') === 'new') {
      setShowForm(true)
      if (items.length === 0) {
        setItems([{
          key: Date.now().toString(),
          treatmentId: '',
          description: '',
          quantity: 1,
          unitPrice: 0,
          toothNumber: '',
        }])
      }
    }
  }, [searchParams])

  useEffect(() => {
    if (showForm) {
      fetchTreatments()
      fetchPatients()
    }
  }, [showForm, fetchTreatments, fetchPatients])

  const addItem = () => {
    setItems([...items, {
      key: Date.now().toString(),
      treatmentId: '',
      description: '',
      quantity: 1,
      unitPrice: 0,
      toothNumber: '',
    }])
  }

  const removeItem = (key: string) => {
    setItems(items.filter(i => i.key !== key))
  }

  const updateItem = (key: string, field: string, value: string | number) => {
    setItems(items.map(i => {
      if (i.key !== key) return i
      const updated = { ...i, [field]: value }
      // Auto-fill from treatment selection
      if (field === 'treatmentId' && value) {
        const treatment = treatments.find((t: Treatment) => t.id === value)
        if (treatment) {
          updated.description = treatment.name
          updated.unitPrice = treatment.defaultPrice / 100 // Display in rupees
        }
      }
      return updated
    }))
  }

  const getSubtotal = () => {
    return items.reduce((sum, item) => sum + (rupeesToPaise(item.unitPrice) * item.quantity), 0)
  }

  const handleCreateInvoice = async () => {
    setFormError('')
    if (!selectedPatient) {
      setFormError('Please select a patient')
      return
    }
    if (items.length === 0) {
      setFormError('Add at least one item')
      return
    }

    // Show confirmation dialog first
    setShowConfirm(true)
  }

  const handleConfirmInvoice = () => {
    createInvoiceMutation.mutate({
      patientId: selectedPatient,
      items: items.map(i => ({
        treatmentId: i.treatmentId,
        description: i.description,
        quantity: i.quantity,
        unitPrice: rupeesToPaise(i.unitPrice),
        toothNumber: i.toothNumber,
      })),
      discountPercent: parseFloat(discount) || 0,
      discountAmount: 0,
      notes: '',
    })
  }

  return (
    <PageTransition className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Billing</h1>
        <Button onClick={() => { setShowForm(!showForm); if (!showForm) addItem() }}>
          <Plus className="h-4 w-4 mr-2" /> New Invoice
        </Button>
      </div>

      {/* Create Invoice Form */}
      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle>Create Invoice</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {formError && <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">{formError}</div>}

            <div className="space-y-2">
              <Label>Patient *</Label>
              <select
                value={selectedPatient}
                onChange={(e) => setSelectedPatient(e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">Select patient...</option>
                {patients.map(p => (
                  <option key={p.id} value={p.id}>{p.name} ({p.phone})</option>
                ))}
              </select>
            </div>

            {/* Line Items */}
            <div className="space-y-3">
              <Label>Items</Label>
              {items.map((item) => (
                <div key={item.key} className="grid grid-cols-12 gap-2 items-end">
                  <div className="col-span-4">
                    <select
                      value={item.treatmentId}
                      onChange={(e) => updateItem(item.key, 'treatmentId', e.target.value)}
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    >
                      <option value="">Custom item</option>
                      {treatments.map((t: Treatment) => (
                        <option key={t.id} value={t.id}>{t.name} - {formatCurrency(t.defaultPrice)}</option>
                      ))}
                    </select>
                  </div>
                  <div className="col-span-3">
                    <Input
                      placeholder="Description"
                      value={item.description}
                      onChange={(e) => updateItem(item.key, 'description', e.target.value)}
                    />
                  </div>
                  <div className="col-span-1">
                    <Input
                      type="number"
                      min={1}
                      value={item.quantity}
                      onChange={(e) => updateItem(item.key, 'quantity', parseInt(e.target.value) || 1)}
                    />
                  </div>
                  <div className="col-span-2">
                    <Input
                      type="number"
                      step="0.01"
                      placeholder="₹ Price"
                      value={item.unitPrice || ''}
                      onChange={(e) => updateItem(item.key, 'unitPrice', parseFloat(e.target.value) || 0)}
                    />
                  </div>
                  <div className="col-span-1">
                    <Input placeholder="Tooth" value={item.toothNumber} onChange={(e) => updateItem(item.key, 'toothNumber', e.target.value)} />
                  </div>
                  <div className="col-span-1">
                    <Button variant="ghost" size="icon" onClick={() => removeItem(item.key)}>
                      <Trash2 className="h-4 w-4 text-red-500" />
                    </Button>
                  </div>
                </div>
              ))}
              <Button variant="outline" size="sm" onClick={addItem}>+ Add Item</Button>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label>Discount %</Label>
                <Input type="number" min={0} max={100} value={discount} onChange={(e) => setDiscount(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label>Subtotal</Label>
                <p className="text-lg font-bold">{formatCurrency(getSubtotal())}</p>
              </div>
            </div>

            <div className="flex gap-2">
              <Button onClick={handleCreateInvoice} disabled={createInvoiceMutation.isPending}>
                {createInvoiceMutation.isPending ? 'Creating...' : 'Create Invoice'}
              </Button>
              <Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Invoice Confirmation Dialog */}
      {showConfirm && (
        <InvoiceConfirmDialog
          patient={patients.find(p => p.id === selectedPatient)}
          items={items}
          discount={parseFloat(discount) || 0}
          treatments={treatments}
          subtotal={getSubtotal()}
          onConfirm={handleConfirmInvoice}
          onBack={() => setShowConfirm(false)}
        />
      )}

      {/* Filter */}
      <div className="flex gap-2">
        {['', 'issued', 'partial', 'paid', 'void'].map(status => (
          <Button
            key={status}
            variant={statusFilter === status ? 'default' : 'outline'}
            size="sm"
            onClick={() => setStatusFilter(status)}
          >
            {status ? INVOICE_STATUS_LABELS[status] : 'All'}
          </Button>
        ))}
      </div>

      {/* Invoice List */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 text-center text-muted-foreground">Loading...</div>
          ) : invoices.length === 0 ? (
            <div className="p-6 text-center text-muted-foreground">No invoices found.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="border-b bg-muted/50">
                  <tr>
                    <th className="px-4 py-3 text-left text-sm font-medium">Invoice #</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Patient</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Date</th>
                    <th className="px-4 py-3 text-right text-sm font-medium">Total</th>
                    <th className="px-4 py-3 text-right text-sm font-medium">Balance</th>
                    <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {invoices.map((invoice: any) => (
                    <tr key={invoice.id} className="border-b hover:bg-muted/30 cursor-pointer" onClick={() => navigate(`/billing/${invoice.id}`)}>
                      <td className="px-4 py-3 text-sm font-mono">{invoice.invoiceNumber}</td>
                      <td className="px-4 py-3 text-sm">{invoice.patient?.name || '-'}</td>
                      <td className="px-4 py-3 text-sm">{formatDate(invoice.invoiceDate)}</td>
                      <td className="px-4 py-3 text-sm text-right">{formatCurrency(invoice.totalAmount)}</td>
                      <td className="px-4 py-3 text-sm text-right">{formatCurrency(invoice.balanceAmount)}</td>
                      <td className="px-4 py-3 text-sm">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(invoice.status)}`}>
                          {INVOICE_STATUS_LABELS[invoice.status]}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

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
