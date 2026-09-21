import { useRef, useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import { Plus, Upload } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Badge } from '@/components/ui/Badge'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useCreateFabric, useFabrics, useProductTypes, useUpdateFabric, useUploadFabricImage } from '@/hooks/useCatalog'
import { useMaterials } from '@/hooks/useMaterials'
import { formatCurrency, toUploadUrl } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Fabric } from '@/types/catalog'

const schema = z.object({
  product_type_id: z.string().min(1, 'Wajib dipilih'),
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  material_id: z.string().min(1, 'Wajib dipilih'),
  sales_price: z.string().min(1, 'Wajib diisi'),
  sort_order: z.string().min(1, 'Wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function FabricsPage() {
  const { data, isLoading, error } = useFabrics()
  const { data: materials } = useMaterials()
  const { data: productTypes } = useProductTypes()
  const createFabric = useCreateFabric()
  const updateFabric = useUpdateFabric()
  const uploadImage = useUploadFabricImage()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Fabric | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { register, control, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ product_type_id: '', code: '', name: '', material_id: '', sales_price: '0', sort_order: '0' })
    setOpen(true)
  }

  function openEdit(f: Fabric) {
    setEditing(f)
    reset({
      product_type_id: f.product_type_id, code: f.code, name: f.name,
      material_id: f.material_id, sales_price: f.sales_price, sort_order: String(f.sort_order),
    })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const input = { ...values, sort_order: Number(values.sort_order) }
    const action = editing ? updateFabric.mutateAsync({ id: editing.id, input }) : createFabric.mutateAsync(input)
    action
      .then(() => {
        toast.success(editing ? 'Bahan diperbarui' : 'Bahan dibuat')
        if (!editing) setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  function onPickImage(file: File | undefined) {
    if (!file || !editing) return
    uploadImage
      .mutateAsync({ id: editing.id, file })
      .then(() => toast.success('Foto sample bahan diunggah'))
      .catch((err) => toast.error('Gagal mengunggah gambar', err instanceof ApiError ? err.message : undefined))
      .finally(() => {
        if (fileInputRef.current) fileInputRef.current.value = ''
      })
  }

  const liveEditing = editing ? (data?.find((f) => f.id === editing.id) ?? editing) : null
  const materialName = (id: string) => materials?.find((m) => m.id === id)?.name ?? id.slice(0, 8)
  const typeName = (id: string) => productTypes?.find((t) => t.id === id)?.name ?? id.slice(0, 8)

  const columns: Column<Fabric>[] = [
    {
      key: 'thumb',
      header: '',
      render: (f) =>
        f.image_url ? (
          <img src={toUploadUrl(f.image_url)} alt={f.name} className="h-10 w-10 rounded-md object-cover" />
        ) : (
          <div className="h-10 w-10 rounded-md bg-slate-100" />
        ),
      className: 'w-14',
    },
    { key: 'code', header: 'Kode', render: (f) => f.code, csvValue: (f) => f.code },
    { key: 'name', header: 'Nama Bahan', render: (f) => f.name, sortValue: (f) => f.name, csvValue: (f) => f.name },
    { key: 'jenis', header: 'Jenis', render: (f) => typeName(f.product_type_id), csvValue: (f) => typeName(f.product_type_id) },
    { key: 'material', header: 'Bahan Baku (BOM)', render: (f) => materialName(f.material_id), csvValue: (f) => materialName(f.material_id) },
    {
      key: 'sales_price',
      header: 'Harga Jual Dasar',
      render: (f) => formatCurrency(f.sales_price),
      sortValue: (f) => Number(f.sales_price),
      csvValue: (f) => f.sales_price,
    },
    {
      key: 'status',
      header: 'Status',
      render: (f) => <Badge tone={f.is_active ? 'success' : 'neutral'}>{f.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (f) => (f.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Bahan"
        description="Bahan yang bisa dipilih customer per jenis pesanan, dengan foto sample dan harga jual dasar per unit ukuran normal (×1.0)."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Bahan Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(f) => f.id}
          exportFilename="fabrics"
          searchPlaceholder="Cari bahan..."
          searchFn={(f, q) => f.name.toLowerCase().includes(q) || f.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Bahan' : 'Bahan Baru'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="space-y-1.5">
              <Label>Jenis Pesanan</Label>
              <Controller
                control={control}
                name="product_type_id"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger>
                      <SelectValue placeholder="Pilih jenis pesanan" />
                    </SelectTrigger>
                    <SelectContent>
                      {productTypes?.map((t) => (
                        <SelectItem key={t.id} value={t.id}>
                          {t.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Kode</Label>
                <Input {...register('code')} disabled={!!editing} placeholder="BHN-A" />
              </div>
              <div className="space-y-1.5">
                <Label>Nama Bahan</Label>
                <Input {...register('name')} placeholder="Jersey A" />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Bahan Baku (untuk BOM &amp; biaya produksi)</Label>
              <Controller
                control={control}
                name="material_id"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger>
                      <SelectValue placeholder="Pilih bahan baku" />
                    </SelectTrigger>
                    <SelectContent>
                      {materials?.map((m) => (
                        <SelectItem key={m.id} value={m.id}>
                          {m.code} &ndash; {m.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Harga Jual Dasar</Label>
                <Input type="number" step="1" {...register('sales_price')} />
              </div>
              <div className="space-y-1.5">
                <Label>Urutan</Label>
                <Input type="number" step="1" {...register('sort_order')} />
              </div>
            </div>

            {editing && (
              <div className="space-y-1.5">
                <Label>Foto Sample Bahan</Label>
                <div className="flex items-center gap-3">
                  {liveEditing?.image_url ? (
                    <img src={toUploadUrl(liveEditing.image_url)} alt="" className="h-16 w-16 rounded-md border border-slate-200 object-cover" />
                  ) : (
                    <div className="h-16 w-16 rounded-md bg-slate-100" />
                  )}
                  <Button type="button" variant="secondary" size="sm" onClick={() => fileInputRef.current?.click()} loading={uploadImage.isPending}>
                    <Upload className="h-4 w-4" /> {liveEditing?.image_url ? 'Ganti Foto' : 'Unggah Foto'}
                  </Button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    className="hidden"
                    onChange={(e) => onPickImage(e.target.files?.[0])}
                  />
                </div>
              </div>
            )}

            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                {editing ? 'Tutup' : 'Batal'}
              </Button>
              <Button type="submit" loading={createFabric.isPending || updateFabric.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
