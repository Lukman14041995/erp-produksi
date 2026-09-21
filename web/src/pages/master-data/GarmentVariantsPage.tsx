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
import { useCreateGarmentVariant, useGarmentVariants, useProductTypes, useUpdateGarmentVariant, useUploadVariantIcon } from '@/hooks/useCatalog'
import { formatCurrency, toUploadUrl } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { GarmentVariant } from '@/types/catalog'

const schema = z.object({
  product_type_id: z.string().min(1, 'Wajib dipilih'),
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  price_addon: z.string().min(1, 'Wajib diisi'),
  sort_order: z.string().min(1, 'Wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function GarmentVariantsPage() {
  const { data, isLoading, error } = useGarmentVariants()
  const { data: productTypes } = useProductTypes()
  const createVariant = useCreateGarmentVariant()
  const updateVariant = useUpdateGarmentVariant()
  const uploadIcon = useUploadVariantIcon()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<GarmentVariant | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { register, control, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ product_type_id: '', code: '', name: '', price_addon: '0', sort_order: '0' })
    setOpen(true)
  }

  function openEdit(v: GarmentVariant) {
    setEditing(v)
    reset({ product_type_id: v.product_type_id, code: v.code, name: v.name, price_addon: v.price_addon, sort_order: String(v.sort_order) })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const input = { ...values, sort_order: Number(values.sort_order) }
    const action = editing ? updateVariant.mutateAsync({ id: editing.id, input }) : createVariant.mutateAsync(input)
    action
      .then(() => {
        toast.success(editing ? 'Model diperbarui' : 'Model dibuat')
        if (!editing) setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  function onPickIcon(file: File | undefined) {
    if (!file || !editing) return
    uploadIcon
      .mutateAsync({ id: editing.id, file })
      .then(() => toast.success('Ikon model diunggah'))
      .catch((err) => toast.error('Gagal mengunggah ikon', err instanceof ApiError ? err.message : undefined))
      .finally(() => {
        if (fileInputRef.current) fileInputRef.current.value = ''
      })
  }

  const liveEditing = editing ? (data?.find((v) => v.id === editing.id) ?? editing) : null
  const typeName = (id: string) => productTypes?.find((t) => t.id === id)?.name ?? id.slice(0, 8)

  const columns: Column<GarmentVariant>[] = [
    {
      key: 'icon',
      header: '',
      render: (v) =>
        v.icon_url ? (
          <img src={toUploadUrl(v.icon_url)} alt={v.name} className="h-8 w-8 rounded object-contain" />
        ) : (
          <div className="h-8 w-8 rounded bg-slate-100" />
        ),
      className: 'w-14',
    },
    { key: 'code', header: 'Kode', render: (v) => v.code, csvValue: (v) => v.code },
    { key: 'name', header: 'Nama Model', render: (v) => v.name, sortValue: (v) => v.name, csvValue: (v) => v.name },
    { key: 'jenis', header: 'Jenis', render: (v) => typeName(v.product_type_id), csvValue: (v) => typeName(v.product_type_id) },
    {
      key: 'price_addon',
      header: 'Biaya Tambahan',
      render: (v) => formatCurrency(v.price_addon),
      sortValue: (v) => Number(v.price_addon),
      csvValue: (v) => v.price_addon,
    },
    {
      key: 'status',
      header: 'Status',
      render: (v) => <Badge tone={v.is_active ? 'success' : 'neutral'}>{v.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (v) => (v.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Model Potongan"
        description="Model seperti Normal/Long Sleeve/Raglan per jenis pesanan, ditampilkan sebagai ikon di configurator (bukan foto sample)."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Model Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(v) => v.id}
          exportFilename="garment-variants"
          searchPlaceholder="Cari model..."
          searchFn={(v, q) => v.name.toLowerCase().includes(q) || v.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Model' : 'Model Baru'}</DialogTitle>
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
                <Input {...register('code')} disabled={!!editing} placeholder="LONG_SLEEVE" />
              </div>
              <div className="space-y-1.5">
                <Label>Nama</Label>
                <Input {...register('name')} placeholder="Lengan Panjang" />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Biaya Tambahan (per pcs)</Label>
                <Input type="number" step="1" {...register('price_addon')} />
              </div>
              <div className="space-y-1.5">
                <Label>Urutan</Label>
                <Input type="number" step="1" {...register('sort_order')} />
              </div>
            </div>

            {editing && (
              <div className="space-y-1.5">
                <Label>Ikon</Label>
                <div className="flex items-center gap-3">
                  {liveEditing?.icon_url ? (
                    <img src={toUploadUrl(liveEditing.icon_url)} alt="" className="h-12 w-12 rounded-md border border-slate-200 object-contain" />
                  ) : (
                    <div className="h-12 w-12 rounded-md bg-slate-100" />
                  )}
                  <Button type="button" variant="secondary" size="sm" onClick={() => fileInputRef.current?.click()} loading={uploadIcon.isPending}>
                    <Upload className="h-4 w-4" /> {liveEditing?.icon_url ? 'Ganti Ikon' : 'Unggah Ikon'}
                  </Button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    className="hidden"
                    onChange={(e) => onPickIcon(e.target.files?.[0])}
                  />
                </div>
              </div>
            )}

            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                {editing ? 'Tutup' : 'Batal'}
              </Button>
              <Button type="submit" loading={createVariant.isPending || updateVariant.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
