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
import { useCreateGarmentSize, useGarmentSizes, useUpdateGarmentSize } from '@/hooks/useCatalog'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { GarmentSize } from '@/types/catalog'

const schema = z.object({
  size_code: z.string().min(1, 'Wajib diisi'),
  size_multiplier: z.string().min(1, 'Wajib diisi'),
  sort_order: z.string().min(1, 'Wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function GarmentSizesPage() {
  const { data, isLoading, error } = useGarmentSizes()
  const createSize = useCreateGarmentSize()
  const updateSize = useUpdateGarmentSize()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<GarmentSize | null>(null)

  const { register, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ size_code: '', size_multiplier: '1', sort_order: '0' })
    setOpen(true)
  }

  function openEdit(s: GarmentSize) {
    setEditing(s)
    reset({ size_code: s.size_code, size_multiplier: s.size_multiplier, sort_order: String(s.sort_order) })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const input = { ...values, sort_order: Number(values.sort_order) }
    const action = editing ? updateSize.mutateAsync({ id: editing.id, input }) : createSize.mutateAsync(input)
    action
      .then(() => {
        toast.success(editing ? 'Ukuran diperbarui' : 'Ukuran dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<GarmentSize>[] = [
    { key: 'size_code', header: 'Kode Ukuran', render: (s) => s.size_code, csvValue: (s) => s.size_code },
    {
      key: 'size_multiplier',
      header: 'Pengali Harga',
      render: (s) => `×${s.size_multiplier}`,
      sortValue: (s) => Number(s.size_multiplier),
      csvValue: (s) => s.size_multiplier,
    },
    { key: 'sort_order', header: 'Urutan', render: (s) => s.sort_order, sortValue: (s) => s.sort_order, csvValue: (s) => s.sort_order },
    {
      key: 'status',
      header: 'Status',
      render: (s) => <Badge tone={s.is_active ? 'success' : 'neutral'}>{s.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (s) => (s.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Ukuran Pesanan"
        description="Ukuran generik (S/M/L/XL) untuk pesanan konfigurasi customer, terpisah dari ukuran per produk."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Ukuran Baru
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
          exportFilename="garment-sizes"
          searchPlaceholder="Cari ukuran..."
          searchFn={(s, q) => s.size_code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Ukuran' : 'Ukuran Baru'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kode Ukuran</Label>
                <Input {...register('size_code')} placeholder="XL" />
              </div>
              <div className="space-y-1.5">
                <Label>Pengali Harga</Label>
                <Input type="number" step="0.0001" {...register('size_multiplier')} />
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
              <Button type="submit" loading={createSize.isPending || updateSize.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
