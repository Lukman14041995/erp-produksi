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
import { StatusBadge } from '@/components/StatusBadge'
import { useJournals, usePostManualJournal } from '@/hooks/useAccounting'
import { formatCurrency, formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { JournalEntry } from '@/types/accounting'

interface LineDraft {
  key: string
  accountCode: string
  debit: string
  credit: string
  description: string
}

export function JournalsPage() {
  const { data, isLoading, error } = useJournals()
  const postManual = usePostManualJournal()
  const navigate = useNavigate()

  const [open, setOpen] = useState(false)
  const [description, setDescription] = useState('')
  const [lines, setLines] = useState<LineDraft[]>([
    { key: uid(), accountCode: '', debit: '', credit: '', description: '' },
    { key: uid(), accountCode: '', debit: '', credit: '', description: '' },
  ])

  const totalDebit = lines.reduce((s, l) => s + (Number(l.debit) || 0), 0)
  const totalCredit = lines.reduce((s, l) => s + (Number(l.credit) || 0), 0)
  const balanced = totalDebit === totalCredit && totalDebit > 0

  function submit() {
    if (!balanced) return toast.error('Jurnal tidak seimbang', 'Total debit harus sama dengan total kredit')
    postManual
      .mutateAsync({
        description,
        lines: lines
          .filter((l) => l.accountCode && (l.debit || l.credit))
          .map((l) => ({ account_code: l.accountCode, debit: l.debit || '0', credit: l.credit || '0', description: l.description })),
      })
      .then((j) => {
        toast.success('Jurnal diposting', j.journal_number)
        setOpen(false)
        setDescription('')
        setLines([
          { key: uid(), accountCode: '', debit: '', credit: '', description: '' },
          { key: uid(), accountCode: '', debit: '', credit: '', description: '' },
        ])
      })
      .catch((err) => toast.error('Gagal posting jurnal', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<JournalEntry>[] = [
    { key: 'no', header: 'No. Jurnal', render: (j) => j.journal_number, sortValue: (j) => j.journal_number, csvValue: (j) => j.journal_number },
    { key: 'date', header: 'Tanggal', render: (j) => formatDate(j.journal_date), sortValue: (j) => j.journal_date, csvValue: (j) => j.journal_date },
    { key: 'source', header: 'Sumber', render: (j) => j.source_type, csvValue: (j) => j.source_type },
    { key: 'desc', header: 'Deskripsi', render: (j) => j.description, csvValue: (j) => j.description },
    { key: 'debit', header: 'Total Debit', render: (j) => formatCurrency(j.total_debit), sortValue: (j) => Number(j.total_debit), csvValue: (j) => j.total_debit },
    { key: 'status', header: 'Status', render: (j) => <StatusBadge kind="journal" value={j.status} />, csvValue: (j) => j.status },
  ]

  return (
    <div>
      <PageHeader
        title="Entri Jurnal"
        description="Posting akuntansi berpasangan yang tidak dapat diubah. Koreksi dilakukan melalui jurnal balik, bukan pengeditan."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Jurnal Manual
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(j) => j.id}
          exportFilename="journals"
          searchPlaceholder="Cari no. jurnal..."
          searchFn={(j, q) => j.journal_number.toLowerCase().includes(q) || j.description.toLowerCase().includes(q)}
          onRowClick={(j) => navigate(`/accounting/journals/${j.id}`)}
          pageSize={15}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>Jurnal Manual Baru</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label>Deskripsi</Label>
              <Input value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>
            {lines.map((line) => (
              <div key={line.key} className="flex items-end gap-2">
                <Input
                  placeholder="Kode akun"
                  className="w-28"
                  value={line.accountCode}
                  onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, accountCode: e.target.value } : l)))}
                />
                <Input
                  placeholder="Deskripsi baris"
                  className="flex-1"
                  value={line.description}
                  onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, description: e.target.value } : l)))}
                />
                <Input
                  type="number"
                  placeholder="Debit"
                  className="w-24"
                  value={line.debit}
                  onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, debit: e.target.value, credit: '' } : l)))}
                />
                <Input
                  type="number"
                  placeholder="Kredit"
                  className="w-24"
                  value={line.credit}
                  onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, credit: e.target.value, debit: '' } : l)))}
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
              onClick={() => setLines((ls) => [...ls, { key: uid(), accountCode: '', debit: '', credit: '', description: '' }])}
            >
              <Plus className="h-3.5 w-3.5" /> Tambah Baris
            </Button>

            <div className={`flex justify-between rounded-md px-3 py-2 text-sm ${balanced ? 'bg-[var(--color-success-surface)] text-[var(--color-success)]' : 'bg-[var(--color-danger-surface)] text-[var(--color-danger)]'}`}>
              <span>Debit: {formatCurrency(totalDebit)}</span>
              <span>Kredit: {formatCurrency(totalCredit)}</span>
              <span>{balanced ? 'Seimbang' : 'Tidak Seimbang'}</span>
            </div>
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Batal
            </Button>
            <Button onClick={submit} loading={postManual.isPending} disabled={!balanced}>
              Posting Jurnal
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
