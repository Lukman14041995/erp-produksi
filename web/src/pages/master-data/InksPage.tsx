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
import { useCreateInk, useInks, useProductTypes, useUpdateInk, useUploadInkImage } from '@/hooks/useCatalog'
import { useMaterials } from '@/hooks/useMaterials'
import { formatCurrency, toUploadUrl } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Ink } from '@/types/catalog'

const schema = z.object({
  product_type_id: z.string().min(1, 'Wajib dipilih'),
  material_id: z.string().min(1, 'Wajib dipilih'),
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  price_addon: z.string().min(1, 'Wajib diisi'),
  sort_order: z.string().min(1, 'Wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function InksPage() {
  const { data, isLoading, error } = useInks()
  const { data: productTypes } = useProductTypes()
  const { data: materials } = useMaterials()
  const createInk = useCreateInk()
  const updateInk = useUpdateInk()
  const uploadImage = useUploadInkImage()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Ink | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { register, control, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ product_type_id: '', material_id: '', code: '', name: '', price_addon: '0', sort_order: '0' })
    setOpen(true)
  }

  function openEdit(i: Ink) {
    setEditing(i)
    reset({
      product_type_id: i.product_type_id, material_id: i.material_id, code: i.code, name: i.name,
      price_addon: i.price_addon, sort_order: String(i.sort_order),
    })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const input = { ...values, sort_order: Number(values.sort_order) }
    const action = editing ? updateInk.mutateAsync({ id: editing.id, input }) : createInk.mutateAsync(input)
    action
      .then(() => {
        toast.success(editing ? 'Tinta diperbarui' : 'Tinta dibuat')
        if (!editing) setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  function onPickImage(file: File | undefined) {
    if (!file || !editing) return
    uploadImage
      .mutateAsync({ id: editing.id, file })
      .then(() => toast.success('Foto sample tinta diunggah'))
      .catch((err) => toast.error('Gagal mengunggah gambar', err instanceof ApiError ? err.message : undefined))
      .finally(() => {
        if (fileInputRef.current) fileInputRef.current.value = ''
      })
  }

  const liveEditing = editing ? (data?.find((i) => i.id === editing.id) ?? editing) : null
  const typeName = (id: string) => productTypes?.find((t) => t.id === id)?.name ?? id.slice(0, 8)
  const material = (id: string) => materials?.find((m) => m.id === id)

  const columns: Column<Ink>[] = [
    {
      key: 'thumb',
      header: '',
      render: (i) =>
        i.image_url ? (
          <img src={toUploadUrl(i.image_url)} alt={i.name} className="h-10 w-10 rounded-md object-cover" />
        ) : (
          <div className="h-10 w-10 rounded-md bg-slate-100" />
        ),
      className: 'w-14',
    },
    { key: 'code', header: 'Kode', render: (i) => i.code, csvValue: (i) => i.code },
    { key: 'name', header: 'Nama Tinta', render: (i) => i.name, sortValue: (i) => i.name, csvValue: (i) => i.name },
    { key: 'jenis', header: 'Jenis', render: (i) => typeName(i.product_type_id), csvValue: (i) => typeName(i.product_type_id) },
    {
      key: 'material',
      header: 'Bahan Baku (Harga Modal)',
      render: (i) => {
        const m = material(i.material_id)
        return m ? `${m.name} · ${formatCurrency(m.unit_cost)}/${m.uom}` : '-'
      },
      csvValue: (i) => material(i.material_id)?.name ?? '',
    },
    {
      key: 'price_addon',
      header: 'Biaya Tambahan',
      render: (i) => formatCurrency(i.price_addon),
      sortValue: (i) => Number(i.price_addon),
      csvValue: (i) => i.price_addon,
    },
    {
      key: 'status',
      header: 'Status',
      render: (i) => <Badge tone={i.is_active ? 'success' : 'neutral'}>{i.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (i) => (i.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Tinta"
        description="Jenis tinta/sablon (Rubber, Plastisol, dst.), khusus per jenis pesanan, dengan foto sample hasil cetak."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Tinta Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(i) => i.id}
          exportFilename="inks"
          searchPlaceholder="Cari tinta..."
          searchFn={(i, q) => i.name.toLowerCase().includes(q) || i.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Tinta' : 'Tinta Baru'}</DialogTitle>
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
                <Input {...register('code')} disabled={!!editing} placeholder="RUBBER" />
              </div>
              <div className="space-y-1.5">
                <Label>Nama Tinta</Label>
                <Input {...register('name')} placeholder="Rubber" />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Bahan Baku (untuk harga modal per warna)</Label>
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
                          {m.code} &ndash; {m.name} ({formatCurrency(m.unit_cost)}/{m.uom})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
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
                <Label>Foto Sample Hasil Cetak</Label>
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
              <Button type="submit" loading={createInk.isPending || updateInk.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
