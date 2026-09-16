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
import { useCreateWarehouse, useUpdateWarehouse, useWarehouses } from '@/hooks/useWarehouse'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Warehouse } from '@/types/warehouse'

const schema = z.object({
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  address: z.string().optional(),
})
type FormValues = z.infer<typeof schema>

export function WarehousesPage() {
  const { data, isLoading, error } = useWarehouses()
  const createWarehouse = useCreateWarehouse()
  const updateWarehouse = useUpdateWarehouse()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Warehouse | null>(null)

  const { register, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ code: '', name: '', address: '' })
    setOpen(true)
  }

  function openEdit(w: Warehouse) {
    setEditing(w)
    reset({ code: w.code, name: w.name, address: w.address })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const action = editing
      ? updateWarehouse.mutateAsync({ id: editing.id, input: { name: values.name, address: values.address } })
      : createWarehouse.mutateAsync(values)
    action
      .then(() => {
        toast.success(editing ? 'Gudang diperbarui' : 'Gudang dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<Warehouse>[] = [
    { key: 'code', header: 'Kode', render: (w) => w.code, sortValue: (w) => w.code, csvValue: (w) => w.code },
    { key: 'name', header: 'Nama', render: (w) => w.name, sortValue: (w) => w.name, csvValue: (w) => w.name },
    { key: 'address', header: 'Alamat', render: (w) => w.address || '-', csvValue: (w) => w.address },
    {
      key: 'status',
      header: 'Status',
      render: (w) => <Badge tone={w.is_active ? 'success' : 'neutral'}>{w.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (w) => (w.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Gudang"
        description="Lokasi penyimpanan yang digunakan untuk pelacakan persediaan multi-gudang."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Gudang Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(w) => w.id}
          exportFilename="warehouses"
          searchPlaceholder="Cari gudang..."
          searchFn={(w, q) => w.name.toLowerCase().includes(q) || w.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Gudang' : 'Gudang Baru'}</DialogTitle>
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
            <div className="space-y-1.5">
              <Label>Alamat</Label>
              <Input {...register('address')} />
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createWarehouse.isPending || updateWarehouse.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
