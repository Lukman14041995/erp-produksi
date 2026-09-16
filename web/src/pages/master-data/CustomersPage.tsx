import { useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
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
import { useCreateCustomer, useCustomers, useUpdateCustomer } from '@/hooks/useCustomers'
import { formatCurrency, formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Customer } from '@/types/master'

const schema = z.object({
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  contact_person: z.string().optional(),
  phone: z.string().optional(),
  email: z.union([z.string().email(), z.literal('')]).optional(),
  address: z.string().optional(),
  tax_id: z.string().optional(),
  credit_limit: z.string().optional(),
})
type FormValues = z.infer<typeof schema>

export function CustomersPage() {
  const { data, isLoading, error } = useCustomers()
  const createCustomer = useCreateCustomer()
  const updateCustomer = useUpdateCustomer()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Customer | null>(null)

  const { register, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ code: '', name: '', contact_person: '', phone: '', email: '', address: '', tax_id: '', credit_limit: '0' })
    setOpen(true)
  }

  function openEdit(c: Customer) {
    setEditing(c)
    reset({ ...c })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const action = editing
      ? updateCustomer.mutateAsync({ id: editing.id, input: values })
      : createCustomer.mutateAsync(values)
    action
      .then(() => {
        toast.success(editing ? 'Pelanggan diperbarui' : 'Pelanggan dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<Customer>[] = [
    { key: 'code', header: 'Kode', render: (c) => c.code, sortValue: (c) => c.code, csvValue: (c) => c.code },
    { key: 'name', header: 'Nama', render: (c) => c.name, sortValue: (c) => c.name, csvValue: (c) => c.name },
    { key: 'contact', header: 'Kontak', render: (c) => c.contact_person || '-', csvValue: (c) => c.contact_person },
    { key: 'phone', header: 'Telepon', render: (c) => c.phone || '-', csvValue: (c) => c.phone },
    {
      key: 'credit_limit',
      header: 'Batas Kredit',
      render: (c) => formatCurrency(c.credit_limit),
      sortValue: (c) => Number(c.credit_limit),
      csvValue: (c) => c.credit_limit,
    },
    {
      key: 'status',
      header: 'Status',
      render: (c) => <Badge tone={c.is_active ? 'success' : 'neutral'}>{c.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (c) => (c.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
    { key: 'created_at', header: 'Dibuat', render: (c) => formatDate(c.created_at), sortValue: (c) => c.created_at, csvValue: (c) => c.created_at },
  ]

  return (
    <div>
      <PageHeader
        title="Pelanggan"
        description="Kelola data master pelanggan yang digunakan di pesanan penjualan dan faktur."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Pelanggan Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(c) => c.id}
          exportFilename="customers"
          searchPlaceholder="Cari pelanggan..."
          searchFn={(c, q) => c.name.toLowerCase().includes(q) || c.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Pelanggan' : 'Pelanggan Baru'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kode</Label>
                <Input {...register('code')} disabled={!!editing} />
              </div>
              <div className="space-y-1.5">
                <Label>Nama</Label>
                <Input {...register('name')} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kontak</Label>
                <Input {...register('contact_person')} />
              </div>
              <div className="space-y-1.5">
                <Label>Telepon</Label>
                <Input {...register('phone')} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Email</Label>
                <Input type="email" {...register('email')} />
              </div>
              <div className="space-y-1.5">
                <Label>NPWP</Label>
                <Input {...register('tax_id')} />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Alamat</Label>
              <Input {...register('address')} />
            </div>
            <div className="space-y-1.5">
              <Label>Batas Kredit</Label>
              <Input type="number" step="0.01" {...register('credit_limit')} />
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createCustomer.isPending || updateCustomer.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
