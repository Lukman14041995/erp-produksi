import { useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useAccounts, useCreateAccount } from '@/hooks/useAccounts'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Account, AccountType } from '@/types/master'

const schema = z.object({
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  account_type: z.enum(['ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'COGS', 'EXPENSE']),
  normal_balance: z.enum(['DEBIT', 'CREDIT']),
})
type FormValues = z.infer<typeof schema>

const typeTone: Record<AccountType, BadgeTone> = {
  ASSET: 'info',
  LIABILITY: 'warning',
  EQUITY: 'neutral',
  REVENUE: 'success',
  COGS: 'danger',
  EXPENSE: 'danger',
}

const typeLabel: Record<AccountType, string> = {
  ASSET: 'Aset',
  LIABILITY: 'Liabilitas',
  EQUITY: 'Ekuitas',
  REVENUE: 'Pendapatan',
  COGS: 'HPP',
  EXPENSE: 'Beban',
}

const balanceLabel: Record<'DEBIT' | 'CREDIT', string> = {
  DEBIT: 'Debit',
  CREDIT: 'Kredit',
}

export function ChartOfAccountsPage() {
  const { data, isLoading, error } = useAccounts()
  const createAccount = useCreateAccount()
  const [open, setOpen] = useState(false)

  const { register, handleSubmit, control, reset } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { account_type: 'ASSET', normal_balance: 'DEBIT' },
  })

  function onSubmit(values: FormValues) {
    createAccount
      .mutateAsync(values)
      .then(() => {
        toast.success('Akun dibuat')
        setOpen(false)
        reset()
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<Account>[] = [
    { key: 'code', header: 'Kode', render: (a) => a.code, sortValue: (a) => a.code, csvValue: (a) => a.code },
    { key: 'name', header: 'Nama', render: (a) => a.name, sortValue: (a) => a.name, csvValue: (a) => a.name },
    {
      key: 'type',
      header: 'Jenis',
      render: (a) => <Badge tone={typeTone[a.account_type]}>{typeLabel[a.account_type]}</Badge>,
      sortValue: (a) => a.account_type,
      csvValue: (a) => a.account_type,
    },
    { key: 'normal', header: 'Saldo Normal', render: (a) => balanceLabel[a.normal_balance], csvValue: (a) => a.normal_balance },
    {
      key: 'status',
      header: 'Status',
      render: (a) => <Badge tone={a.is_active ? 'success' : 'neutral'}>{a.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (a) => (a.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Bagan Akun"
        description="Bagan akun standar 1-6 yang mendasari buku besar double-entry."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Akun Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(a) => a.id}
          exportFilename="chart-of-accounts"
          searchPlaceholder="Cari akun..."
          searchFn={(a, q) => a.name.toLowerCase().includes(q) || a.code.toLowerCase().includes(q)}
          pageSize={20}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Akun Baru</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kode</Label>
                <Input placeholder="cth. 6-2000" {...register('code')} />
              </div>
              <div className="space-y-1.5">
                <Label>Nama</Label>
                <Input {...register('name')} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Jenis Akun</Label>
                <Controller
                  control={control}
                  name="account_type"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {(['ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'COGS', 'EXPENSE'] as const).map((t) => (
                          <SelectItem key={t} value={t}>
                            {typeLabel[t]}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Saldo Normal</Label>
                <Controller
                  control={control}
                  name="normal_balance"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="DEBIT">Debit</SelectItem>
                        <SelectItem value="CREDIT">Kredit</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createAccount.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
