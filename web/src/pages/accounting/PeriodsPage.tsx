import { useState } from 'react'
import { Lock, Plus, Unlock } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { StatusBadge } from '@/components/StatusBadge'
import { useAccountingPeriods, useClosePeriod, useCreatePeriod, useReopenPeriod } from '@/hooks/useAccounting'
import { formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { AccountingPeriod } from '@/types/accounting'

export function PeriodsPage() {
  const { data, isLoading, error } = useAccountingPeriods()
  const createPeriod = useCreatePeriod()
  const closePeriod = useClosePeriod()
  const reopenPeriod = useReopenPeriod()

  const [open, setOpen] = useState(false)
  const [period, setPeriod] = useState('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  function submit() {
    if (!period || !startDate || !endDate) return toast.error('Isi semua kolom')
    createPeriod
      .mutateAsync({ period, start_date: startDate, end_date: endDate })
      .then(() => {
        toast.success('Periode dibuat')
        setOpen(false)
        setPeriod('')
        setStartDate('')
        setEndDate('')
      })
      .catch((err) => toast.error('Gagal membuat periode', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<AccountingPeriod>[] = [
    { key: 'period', header: 'Periode', render: (p) => p.period, sortValue: (p) => p.period, csvValue: (p) => p.period },
    { key: 'start', header: 'Mulai', render: (p) => formatDate(p.start_date), csvValue: (p) => p.start_date },
    { key: 'end', header: 'Selesai', render: (p) => formatDate(p.end_date), csvValue: (p) => p.end_date },
    { key: 'status', header: 'Status', render: (p) => <StatusBadge kind="period" value={p.status} />, csvValue: (p) => p.status },
    {
      key: 'actions',
      header: '',
      render: (p) =>
        p.status === 'OPEN' ? (
          <Button size="sm" variant="secondary" onClick={() => closePeriod.mutate(p.period)}>
            <Lock className="h-3.5 w-3.5" /> Tutup
          </Button>
        ) : (
          <Button size="sm" variant="secondary" onClick={() => reopenPeriod.mutate(p.period)}>
            <Unlock className="h-3.5 w-3.5" /> Buka Kembali
          </Button>
        ),
      csvValue: () => '',
      className: 'text-right',
    },
  ]

  return (
    <div>
      <PageHeader
        title="Periode Akuntansi"
        description="Menutup periode mengunci semua posting bertanggal di dalamnya dari perubahan lebih lanjut."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Periode Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && <DataTable data={data} columns={columns} rowKey={(p) => p.id} exportFilename="accounting-periods" pageSize={15} />}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Periode Akuntansi Baru</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label>Periode (YYYYMM)</Label>
              <Input value={period} onChange={(e) => setPeriod(e.target.value)} placeholder="202609" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Tanggal Mulai</Label>
                <Input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
              </div>
              <div className="space-y-1.5">
                <Label>Tanggal Selesai</Label>
                <Input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Batal
            </Button>
            <Button onClick={submit} loading={createPeriod.isPending}>
              Simpan Periode
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
