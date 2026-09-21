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
import { Badge } from '@/components/ui/Badge'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useBankAccounts, useCreateBankAccount, useUpdateBankAccount } from '@/hooks/useBilling'
import { useAccounts } from '@/hooks/useAccounts'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { BankAccount } from '@/types/billing'

const schema = z.object({
  bank_name: z.string().min(1, 'Wajib diisi'),
  account_number: z.string().min(1, 'Wajib diisi'),
  account_holder: z.string().min(1, 'Wajib diisi'),
  coa_account_code: z.string().min(1, 'Wajib dipilih'),
  sort_order: z.string().min(1, 'Wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function BankAccountsPage() {
  const { data, isLoading, error } = useBankAccounts()
  const { data: accounts } = useAccounts()
  const createBank = useCreateBankAccount()
  const updateBank = useUpdateBankAccount()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<BankAccount | null>(null)

  const { register, control, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ bank_name: '', account_number: '', account_holder: '', coa_account_code: '', sort_order: '0' })
    setOpen(true)
  }

  function openEdit(b: BankAccount) {
    setEditing(b)
    reset({
      bank_name: b.bank_name, account_number: b.account_number, account_holder: b.account_holder,
      coa_account_code: b.coa_account_code, sort_order: String(b.sort_order),
    })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const input = { ...values, sort_order: Number(values.sort_order) }
    const action = editing ? updateBank.mutateAsync({ id: editing.id, input }) : createBank.mutateAsync(input)
    action
      .then(() => {
        toast.success(editing ? 'Rekening diperbarui' : 'Rekening dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<BankAccount>[] = [
    { key: 'bank_name', header: 'Bank', render: (b) => b.bank_name, sortValue: (b) => b.bank_name, csvValue: (b) => b.bank_name },
    { key: 'account_number', header: 'No. Rekening', render: (b) => b.account_number, csvValue: (b) => b.account_number },
    { key: 'account_holder', header: 'Atas Nama', render: (b) => b.account_holder, csvValue: (b) => b.account_holder },
    { key: 'coa', header: 'Akun COA', render: (b) => b.coa_account_code, csvValue: (b) => b.coa_account_code },
    {
      key: 'status',
      header: 'Status',
      render: (b) => <Badge tone={b.is_active ? 'success' : 'neutral'}>{b.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (b) => (b.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Rekening Bank"
        description="Rekening tujuan transfer yang ditampilkan ke customer saat membayar. Rekening aktif dengan urutan terkecil dipakai otomatis untuk pesanan baru."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Rekening Baru
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
          exportFilename="bank-accounts"
          searchPlaceholder="Cari rekening..."
          searchFn={(b, q) => b.bank_name.toLowerCase().includes(q) || b.account_number.includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Rekening' : 'Rekening Baru'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Nama Bank</Label>
                <Input {...register('bank_name')} placeholder="BCA" />
              </div>
              <div className="space-y-1.5">
                <Label>Nomor Rekening</Label>
                <Input {...register('account_number')} placeholder="1234567890" />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Atas Nama</Label>
              <Input {...register('account_holder')} placeholder="Lukman" />
            </div>
            <div className="space-y-1.5">
              <Label>Akun COA (untuk posting jurnal saat pembayaran dikonfirmasi)</Label>
              <Controller
                control={control}
                name="coa_account_code"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger>
                      <SelectValue placeholder="Pilih akun" />
                    </SelectTrigger>
                    <SelectContent>
                      {accounts?.map((a) => (
                        <SelectItem key={a.id} value={a.code}>
                          {a.code} &ndash; {a.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
            </div>
            <div className="space-y-1.5">
              <Label>Urutan</Label>
              <Input type="number" step="1" {...register('sort_order')} />
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createBank.isPending || updateBank.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
