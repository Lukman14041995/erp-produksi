import { useState } from 'react'
import { Plus, Send, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { StatusBadge } from '@/components/StatusBadge'
import { useCreatePayment, usePayments, usePostPayment, useVoidPayment } from '@/hooks/useFinance'
import { useCustomers } from '@/hooks/useCustomers'
import { useSuppliers } from '@/hooks/useSuppliers'
import { useInvoices } from '@/hooks/useSales'
import { useSupplierInvoices } from '@/hooks/usePurchasing'
import { formatCurrency, formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Payment, PaymentType } from '@/types/finance'

export function PaymentsPage() {
  const { data, isLoading, error } = usePayments()
  const { data: customers } = useCustomers()
  const { data: suppliers } = useSuppliers()
  const { data: invoices } = useInvoices()
  const { data: bills } = useSupplierInvoices()

  const createPayment = useCreatePayment()
  const postPayment = usePostPayment()
  const voidPayment = useVoidPayment()

  const [open, setOpen] = useState(false)
  const [type, setType] = useState<PaymentType>('RECEIPT')
  const [partyId, setPartyId] = useState('')
  const [accountCode, setAccountCode] = useState('1-1002')
  const [amount, setAmount] = useState('')
  const [targetDocId, setTargetDocId] = useState('')

  const openInvoices = invoices?.filter((i) => i.customer_id === partyId && (i.status === 'POSTED' || i.status === 'PARTIALLY_PAID'))
  const openBills = bills?.filter((b) => b.supplier_id === partyId && (b.status === 'POSTED' || b.status === 'PARTIALLY_PAID'))

  function resetForm() {
    setType('RECEIPT')
    setPartyId('')
    setAmount('')
    setTargetDocId('')
  }

  function submit() {
    if (!partyId || !amount) return toast.error('Isi pihak dan jumlah')

    const input =
      type === 'RECEIPT'
        ? {
            payment_type: type,
            customer_id: partyId,
            cash_bank_account_code: accountCode,
            amount,
            allocations: targetDocId ? [{ invoice_id: targetDocId, amount }] : [],
          }
        : {
            payment_type: type,
            supplier_id: partyId,
            cash_bank_account_code: accountCode,
            amount,
            bill_allocations: targetDocId ? [{ supplier_invoice_id: targetDocId, amount }] : [],
          }

    createPayment
      .mutateAsync(input)
      .then((p) => {
        toast.success('Pembayaran dibuat', p.payment_number)
        setOpen(false)
        resetForm()
      })
      .catch((err) => toast.error('Gagal membuat pembayaran', err instanceof ApiError ? err.message : undefined))
  }

  const partyName = (p: Payment): string => {
    if (p.payment_type === 'RECEIPT') return customers?.find((c) => c.id === p.customer_id)?.name ?? p.customer_id?.slice(0, 8) ?? ''
    return suppliers?.find((s) => s.id === p.supplier_id)?.name ?? p.supplier_id?.slice(0, 8) ?? ''
  }

  const columns: Column<Payment>[] = [
    { key: 'no', header: 'No. Pembayaran', render: (p) => p.payment_number, sortValue: (p) => p.payment_number, csvValue: (p) => p.payment_number },
    { key: 'type', header: 'Jenis', render: (p) => p.payment_type, csvValue: (p) => p.payment_type },
    { key: 'party', header: 'Pihak', render: partyName, csvValue: partyName },
    { key: 'date', header: 'Tanggal', render: (p) => formatDate(p.payment_date), sortValue: (p) => p.payment_date, csvValue: (p) => p.payment_date },
    { key: 'amount', header: 'Jumlah', render: (p) => formatCurrency(p.amount), sortValue: (p) => Number(p.amount), csvValue: (p) => p.amount },
    { key: 'status', header: 'Status', render: (p) => <StatusBadge kind="doc" value={p.status} />, csvValue: (p) => p.status },
    {
      key: 'actions',
      header: '',
      render: (p) =>
        p.status === 'DRAFT' ? (
          <div className="flex justify-end gap-1">
            <Button size="sm" variant="secondary" onClick={() => postPayment.mutate({ id: p.id })}>
              <Send className="h-3.5 w-3.5" /> Posting
            </Button>
            <Button size="sm" variant="ghost" onClick={() => voidPayment.mutate({ id: p.id })}>
              <XCircle className="h-3.5 w-3.5 text-[var(--color-danger)]" />
            </Button>
          </div>
        ) : p.status === 'POSTED' ? (
          <Button size="sm" variant="ghost" onClick={() => voidPayment.mutate({ id: p.id, body: { reason: 'Voided from UI' } })}>
            <XCircle className="h-3.5 w-3.5 text-[var(--color-danger)]" /> Batalkan
          </Button>
        ) : null,
      csvValue: () => '',
      className: 'text-right',
    },
  ]

  return (
    <div>
      <PageHeader
        title="Pembayaran"
        description="Penerimaan dari pelanggan dan pengeluaran ke pemasok, dengan alokasi faktur/tagihan."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Pembayaran Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && <DataTable data={data} columns={columns} rowKey={(p) => p.id} exportFilename="payments" pageSize={15} />}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Pembayaran Baru</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label>Jenis Pembayaran</Label>
              <Select
                value={type}
                onValueChange={(v) => {
                  setType(v as PaymentType)
                  setPartyId('')
                  setTargetDocId('')
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="RECEIPT">Penerimaan (dari Pelanggan)</SelectItem>
                  <SelectItem value="DISBURSEMENT">Pengeluaran (ke Pemasok)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label>{type === 'RECEIPT' ? 'Pelanggan' : 'Pemasok'}</Label>
              <Select value={partyId} onValueChange={setPartyId}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih" />
                </SelectTrigger>
                <SelectContent>
                  {(type === 'RECEIPT' ? customers : suppliers)?.map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {partyId && (
              <div className="space-y-1.5">
                <Label>Terapkan ke {type === 'RECEIPT' ? 'Faktur' : 'Tagihan'} (opsional)</Label>
                <Select value={targetDocId || '__none'} onValueChange={(v) => setTargetDocId(v === '__none' ? '' : v)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Belum diterapkan" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none">Belum diterapkan</SelectItem>
                    {(type === 'RECEIPT' ? openInvoices : openBills)?.map((doc) => (
                      <SelectItem key={doc.id} value={doc.id}>
                        {'invoice_number' in doc ? doc.invoice_number : doc.bill_number} &middot; Saldo {formatCurrency(doc.balance_due)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kode Akun Kas/Bank</Label>
                <Input value={accountCode} onChange={(e) => setAccountCode(e.target.value)} placeholder="1-1002" />
              </div>
              <div className="space-y-1.5">
                <Label>Jumlah</Label>
                <Input type="number" value={amount} onChange={(e) => setAmount(e.target.value)} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Batal
            </Button>
            <Button onClick={submit} loading={createPayment.isPending}>
              Simpan Pembayaran
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
