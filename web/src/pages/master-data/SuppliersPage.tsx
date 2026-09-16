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
import { useCreateSupplier, useSuppliers, useUpdateSupplier } from '@/hooks/useSuppliers'
import { formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Supplier } from '@/types/master'

const schema = z.object({
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  contact_person: z.string().optional(),
  phone: z.string().optional(),
  email: z.union([z.string().email(), z.literal('')]).optional(),
  address: z.string().optional(),
  tax_id: z.string().optional(),
})
type FormValues = z.infer<typeof schema>

export function SuppliersPage() {
  const { data, isLoading, error } = useSuppliers()
  const createSupplier = useCreateSupplier()
  const updateSupplier = useUpdateSupplier()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Supplier | null>(null)

  const { register, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ code: '', name: '', contact_person: '', phone: '', email: '', address: '', tax_id: '' })
    setOpen(true)
  }

  function openEdit(s: Supplier) {
    setEditing(s)
    reset({ ...s })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const action = editing ? updateSupplier.mutateAsync({ id: editing.id, input: values }) : createSupplier.mutateAsync(values)
    action
      .then(() => {
        toast.success(editing ? 'Pemasok diperbarui' : 'Pemasok dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<Supplier>[] = [
    { key: 'code', header: 'Kode', render: (s) => s.code, sortValue: (s) => s.code, csvValue: (s) => s.code },
    { key: 'name', header: 'Nama', render: (s) => s.name, sortValue: (s) => s.name, csvValue: (s) => s.name },
    { key: 'contact', header: 'Kontak', render: (s) => s.contact_person || '-', csvValue: (s) => s.contact_person },
    { key: 'phone', header: 'Telepon', render: (s) => s.phone || '-', csvValue: (s) => s.phone },
    {
      key: 'status',
      header: 'Status',
      render: (s) => <Badge tone={s.is_active ? 'success' : 'neutral'}>{s.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (s) => (s.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
    { key: 'created_at', header: 'Dibuat', render: (s) => formatDate(s.created_at), sortValue: (s) => s.created_at, csvValue: (s) => s.created_at },
  ]

  return (
    <div>
      <PageHeader
        title="Pemasok"
        description="Kelola data master pemasok yang digunakan untuk pembelian dan utang usaha."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Pemasok Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(s) => s.id}
          exportFilename="suppliers"
          searchPlaceholder="Cari pemasok..."
          searchFn={(s, q) => s.name.toLowerCase().includes(q) || s.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Pemasok' : 'Pemasok Baru'}</DialogTitle>
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
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createSupplier.isPending || updateSupplier.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
