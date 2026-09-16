import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Undo2 } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent } from '@/components/ui/Card'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { useJournal, useReverseJournal } from '@/hooks/useAccounting'
import { formatCurrency, formatDateTime } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

export function JournalDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: journal, isLoading, error } = useJournal(id)
  const reverseJournal = useReverseJournal()
  const [reverseOpen, setReverseOpen] = useState(false)
  const [reason, setReason] = useState('')

  if (isLoading) return <LoadingState />
  if (error || !journal) return <ErrorState message={(error as Error)?.message ?? 'Jurnal tidak ditemukan'} />

  function submitReverse() {
    reverseJournal
      .mutateAsync({ id: id!, reason })
      .then(() => {
        toast.success('Jurnal dibalik')
        setReverseOpen(false)
      })
      .catch((err) => toast.error('Gagal membalik jurnal', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div>
      <button onClick={() => navigate('/accounting/journals')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Jurnal
      </button>

      <PageHeader
        title={journal.journal_number}
        description={`${journal.source_type} · ${formatDateTime(journal.created_at)} · oleh ${journal.created_by}`}
        actions={
          journal.status === 'POSTED' && (
            <Button variant="destructive" onClick={() => setReverseOpen(true)}>
              <Undo2 className="h-4 w-4" /> Balik Jurnal
            </Button>
          )
        }
      />

      <div className="mb-4">
        <StatusBadge kind="journal" value={journal.status} />
      </div>

      <Card>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Akun</th>
                <th className="px-4 py-2.5">Deskripsi</th>
                <th className="px-4 py-2.5 text-right">Debit</th>
                <th className="px-4 py-2.5 text-right">Kredit</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {journal.lines?.map((l) => (
                <tr key={l.id}>
                  <td className="px-4 py-2.5 font-medium">{l.account_code} · {l.account_name}</td>
                  <td className="px-4 py-2.5 text-slate-500">{l.description}</td>
                  <td className="px-4 py-2.5 text-right">{Number(l.debit) > 0 ? formatCurrency(l.debit) : ''}</td>
                  <td className="px-4 py-2.5 text-right">{Number(l.credit) > 0 ? formatCurrency(l.credit) : ''}</td>
                </tr>
              ))}
            </tbody>
            <tfoot className="border-t-2 border-slate-200 font-semibold">
              <tr>
                <td colSpan={2} className="px-4 py-2.5">
                  Jumlah Total
                </td>
                <td className="px-4 py-2.5 text-right">{formatCurrency(journal.total_debit)}</td>
                <td className="px-4 py-2.5 text-right">{formatCurrency(journal.total_credit)}</td>
              </tr>
            </tfoot>
          </table>
        </CardContent>
      </Card>

      <Dialog open={reverseOpen} onOpenChange={setReverseOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Balik Jurnal {journal.journal_number}</DialogTitle>
          </DialogHeader>
          <div className="space-y-1.5">
            <Label>Alasan</Label>
            <Input value={reason} onChange={(e) => setReason(e.target.value)} placeholder="mis. Diposting ke akun yang salah" />
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setReverseOpen(false)}>
              Batal
            </Button>
            <Button variant="destructive" onClick={submitReverse} loading={reverseJournal.isPending}>
              Balik Jurnal
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
