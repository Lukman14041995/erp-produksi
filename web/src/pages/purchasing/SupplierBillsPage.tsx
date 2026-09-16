import { useState } from 'react'
import { uid } from '@/lib/uid'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { StatusBadge } from '@/components/StatusBadge'
import { useCreateSupplierInvoice, useSupplierInvoices } from '@/hooks/usePurchasing'
import { useSuppliers } from '@/hooks/useSuppliers'
import { useMaterials } from '@/hooks/useMaterials'
import { formatCurrency, formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { SupplierInvoice } from '@/types/purchasing'

interface LineDraft {
  key: string
  materialId: string
  qty: string
  unitCost: string
}

export function SupplierBillsPage() {
  const { data, isLoading, error } = useSupplierInvoices()
  const { data: suppliers } = useSuppliers()
  const { data: materials } = useMaterials()
  const createBill = useCreateSupplierInvoice()
  const navigate = useNavigate()

  const [open, setOpen] = useState(false)
  const [supplierId, setSupplierId] = useState('')
  const [debitAccountCode, setDebitAccountCode] = useState('1-1200')
  const [taxRate, setTaxRate] = useState('0')
  const [lines, setLines] = useState<LineDraft[]>([{ key: uid(), materialId: '', qty: '', unitCost: '' }])

  const supplierName = (id: string) => suppliers?.find((s) => s.id === id)?.name ?? id.slice(0, 8)

  function resetForm() {
    setSupplierId('')
    setDebitAccountCode('1-1200')
    setTaxRate('0')
    setLines([{ key: uid(), materialId: '', qty: '', unitCost: '' }])
  }

  function submit() {
    const items = lines.filter((l) => l.materialId && l.qty && l.unitCost).map((l) => ({ material_id: l.materialId, qty: l.qty, unit_cost: l.unitCost }))
    if (!supplierId || items.length === 0) return toast.error('Pilih pemasok dan tambahkan minimal satu item')

    createBill
      .mutateAsync({ supplier_id: supplierId, debit_account_code: debitAccountCode, tax_rate: taxRate, items })
      .then((bill) => {
        toast.success('Tagihan pemasok dibuat', bill.bill_number)
        setOpen(false)
        resetForm()
      })
      .catch((err) => toast.error('Gagal membuat tagihan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<SupplierInvoice>[] = [
    { key: 'bill', header: 'No. Tagihan', render: (b) => b.bill_number, sortValue: (b) => b.bill_number, csvValue: (b) => b.bill_number },
    { key: 'supplier', header: 'Pemasok', render: (b) => supplierName(b.supplier_id), csvValue: (b) => supplierName(b.supplier_id) },
    { key: 'date', header: 'Tanggal Tagihan', render: (b) => formatDate(b.bill_date), sortValue: (b) => b.bill_date, csvValue: (b) => b.bill_date },
    { key: 'total', header: 'Total Keseluruhan', render: (b) => formatCurrency(b.grand_total), sortValue: (b) => Number(b.grand_total), csvValue: (b) => b.grand_total },
    { key: 'balance', header: 'Sisa Tagihan', render: (b) => formatCurrency(b.balance_due), sortValue: (b) => Number(b.balance_due), csvValue: (b) => b.balance_due },
    { key: 'status', header: 'Status', render: (b) => <StatusBadge kind="bill" value={b.status} />, csvValue: (b) => b.status },
  ]

  return (
    <div>
      <PageHeader
        title="Tagihan Pemasok"
        description="Subledger utang: tagihan dari pemasok, memicu penerimaan bahan baku dan utang usaha."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Tagihan Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(b) => b.id}
          exportFilename="supplier-bills"
          searchPlaceholder="Cari nomor tagihan..."
          searchFn={(b, q) => b.bill_number.toLowerCase().includes(q)}
          onRowClick={(b) => navigate(`/purchasing/bills/${b.id}`)}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Tagihan Pemasok Baru</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Pemasok</Label>
                <Select value={supplierId} onValueChange={setSupplierId}>
                  <SelectTrigger>
                    <SelectValue placeholder="Pilih pemasok" />
                  </SelectTrigger>
                  <SelectContent>
                    {suppliers?.map((s) => (
                      <SelectItem key={s.id} value={s.id}>
                        {s.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>Akun Debit</Label>
                <Input value={debitAccountCode} onChange={(e) => setDebitAccountCode(e.target.value)} placeholder="1-1200" />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Tarif Pajak</Label>
              <Input type="number" step="0.01" value={taxRate} onChange={(e) => setTaxRate(e.target.value)} />
            </div>

            <div className="space-y-2">
              <Label>Item Baris</Label>
              {lines.map((line) => (
                <div key={line.key} className="flex items-end gap-2">
                  <div className="flex-1 space-y-1">
                    <Select value={line.materialId} onValueChange={(v) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, materialId: v } : l)))}>
                      <SelectTrigger>
                        <SelectValue placeholder="Bahan Baku" />
                      </SelectTrigger>
                      <SelectContent>
                        {materials?.map((m) => (
                          <SelectItem key={m.id} value={m.id}>
                            {m.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <Input
                    type="number"
                    placeholder="Jml"
                    className="w-24"
                    value={line.qty}
                    onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, qty: e.target.value } : l)))}
                  />
                  <Input
                    type="number"
                    placeholder="Biaya satuan"
                    className="w-28"
                    value={line.unitCost}
                    onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, unitCost: e.target.value } : l)))}
                  />
                  <Button type="button" variant="ghost" size="icon" onClick={() => setLines((ls) => ls.filter((l) => l.key !== line.key))}>
                    <Trash2 className="h-4 w-4 text-[var(--color-danger)]" />
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => setLines((ls) => [...ls, { key: uid(), materialId: '', qty: '', unitCost: '' }])}
              >
                <Plus className="h-3.5 w-3.5" /> Tambah Baris
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Batal
            </Button>
            <Button onClick={submit} loading={createBill.isPending}>
              Simpan Tagihan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
