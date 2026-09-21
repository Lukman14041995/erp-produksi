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
import { useCreateProductType, useProductTypes, useUpdateProductType } from '@/hooks/useCatalog'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { ProductType } from '@/types/catalog'

const schema = z.object({
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  sort_order: z.string().min(1, 'Wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function ProductTypesPage() {
  const { data, isLoading, error } = useProductTypes()
  const createType = useCreateProductType()
  const updateType = useUpdateProductType()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<ProductType | null>(null)

  const { register, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ code: '', name: '', sort_order: '0' })
    setOpen(true)
  }

  function openEdit(p: ProductType) {
    setEditing(p)
    reset({ code: p.code, name: p.name, sort_order: String(p.sort_order) })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const input = { ...values, sort_order: Number(values.sort_order) }
    const action = editing ? updateType.mutateAsync({ id: editing.id, input }) : createType.mutateAsync(input)
    action
      .then(() => {
        toast.success(editing ? 'Jenis pesanan diperbarui' : 'Jenis pesanan dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<ProductType>[] = [
    { key: 'code', header: 'Kode', render: (p) => p.code, csvValue: (p) => p.code },
    { key: 'name', header: 'Nama', render: (p) => p.name, sortValue: (p) => p.name, csvValue: (p) => p.name },
    {
      key: 'status',
      header: 'Status',
      render: (p) => <Badge tone={p.is_active ? 'success' : 'neutral'}>{p.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (p) => (p.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Jenis Pesanan"
        description="Kategori utama yang dipilih customer pertama kali: Jersey atau T-Shirt. Model, Bahan, dan Tinta masing-masing punya daftar pilihan sendiri per jenis."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Jenis Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(p) => p.id}
          exportFilename="product-types"
          searchPlaceholder="Cari jenis pesanan..."
          searchFn={(p, q) => p.name.toLowerCase().includes(q) || p.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Jenis Pesanan' : 'Jenis Pesanan Baru'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kode</Label>
                <Input {...register('code')} disabled={!!editing} placeholder="JERSEY" />
              </div>
              <div className="space-y-1.5">
                <Label>Nama</Label>
                <Input {...register('name')} placeholder="Jersey" />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Urutan</Label>
              <Input type="number" step="1" {...register('sort_order')} />
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createType.isPending || updateType.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
