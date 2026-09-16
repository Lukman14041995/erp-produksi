import { useState } from 'react'
import { Plus, Send, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { StatusBadge } from '@/components/StatusBadge'
import { useCreateExpense, useExpenses, usePostExpense, useVoidExpense } from '@/hooks/useFinance'
import { formatCurrency, formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Expense } from '@/types/finance'

export function ExpensesPage() {
  const { data, isLoading, error } = useExpenses()
  const createExpense = useCreateExpense()
  const postExpense = usePostExpense()
  const voidExpense = useVoidExpense()

  const [open, setOpen] = useState(false)
  const [expenseAccountCode, setExpenseAccountCode] = useState('6-1100')
  const [paidFromAccountCode, setPaidFromAccountCode] = useState('1-1002')
  const [amount, setAmount] = useState('')
  const [description, setDescription] = useState('')

  function submit() {
    if (!amount || !description) return toast.error('Isi jumlah dan deskripsi')
    createExpense
      .mutateAsync({ expense_account_code: expenseAccountCode, paid_from_account_code: paidFromAccountCode, amount, description })
      .then((e) => {
        toast.success('Pengeluaran dibuat', e.expense_number)
        setOpen(false)
        setAmount('')
        setDescription('')
      })
      .catch((err) => toast.error('Gagal membuat pengeluaran', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<Expense>[] = [
    { key: 'no', header: 'No. Pengeluaran', render: (e) => e.expense_number, sortValue: (e) => e.expense_number, csvValue: (e) => e.expense_number },
    { key: 'date', header: 'Tanggal', render: (e) => formatDate(e.expense_date), sortValue: (e) => e.expense_date, csvValue: (e) => e.expense_date },
    { key: 'desc', header: 'Deskripsi', render: (e) => e.description, csvValue: (e) => e.description },
    { key: 'amount', header: 'Jumlah', render: (e) => formatCurrency(e.amount), sortValue: (e) => Number(e.amount), csvValue: (e) => e.amount },
    { key: 'status', header: 'Status', render: (e) => <StatusBadge kind="doc" value={e.status} />, csvValue: (e) => e.status },
    {
      key: 'actions',
      header: '',
      render: (e) =>
        e.status === 'DRAFT' ? (
          <div className="flex justify-end gap-1">
            <Button size="sm" variant="secondary" onClick={() => postExpense.mutate({ id: e.id })}>
              <Send className="h-3.5 w-3.5" /> Posting
            </Button>
            <Button size="sm" variant="ghost" onClick={() => voidExpense.mutate({ id: e.id })}>
              <XCircle className="h-3.5 w-3.5 text-[var(--color-danger)]" />
            </Button>
          </div>
        ) : e.status === 'POSTED' ? (
          <Button size="sm" variant="ghost" onClick={() => voidExpense.mutate({ id: e.id, body: { reason: 'Voided from UI' } })}>
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
        title="Pengeluaran"
        description="Beban operasional yang diposting langsung terhadap kas/bank."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Pengeluaran Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && <DataTable data={data} columns={columns} rowKey={(e) => e.id} exportFilename="expenses" pageSize={15} />}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Pengeluaran Baru</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kode Akun Beban</Label>
                <Input value={expenseAccountCode} onChange={(e) => setExpenseAccountCode(e.target.value)} placeholder="6-1100" />
              </div>
              <div className="space-y-1.5">
                <Label>Kode Akun Sumber Dana</Label>
                <Input value={paidFromAccountCode} onChange={(e) => setPaidFromAccountCode(e.target.value)} placeholder="1-1002" />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Jumlah</Label>
              <Input type="number" value={amount} onChange={(e) => setAmount(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label>Deskripsi</Label>
              <Input value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Batal
            </Button>
            <Button onClick={submit} loading={createExpense.isPending}>
              Simpan Pengeluaran
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
