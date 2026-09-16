import { useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Plus, Ruler } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Badge } from '@/components/ui/Badge'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { useAddProductSize, useCreateProduct, useProduct, useProducts, useUpdateProduct } from '@/hooks/useProducts'
import { formatCurrency } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Product } from '@/types/master'

const productSchema = z.object({
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  category: z.string().optional(),
  uom: z.string().min(1, 'Wajib diisi'),
  base_price: z.string().min(1, 'Wajib diisi'),
})
type ProductFormValues = z.infer<typeof productSchema>

const sizeSchema = z.object({
  size_code: z.string().min(1, 'Wajib diisi'),
  size_multiplier: z.string().min(1, 'Wajib diisi'),
  sort_order: z.string().optional(),
})
type SizeFormValues = z.infer<typeof sizeSchema>

export function ProductsPage() {
  const { data, isLoading, error } = useProducts()
  const createProduct = useCreateProduct()
  const updateProduct = useUpdateProduct()
  const addSize = useAddProductSize()

  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Product | null>(null)
  const [sizeDialogProductId, setSizeDialogProductId] = useState<string | null>(null)

  const { register, handleSubmit, reset } = useForm<ProductFormValues>({ resolver: zodResolver(productSchema) })
  const sizeForm = useForm<SizeFormValues>({ resolver: zodResolver(sizeSchema) })

  const { data: sizeProduct } = useProduct(sizeDialogProductId ?? undefined)

  function openCreate() {
    setEditing(null)
    reset({ code: '', name: '', category: '', uom: 'PCS', base_price: '0' })
    setOpen(true)
  }

  function openEdit(p: Product) {
    setEditing(p)
    reset({ code: p.code, name: p.name, category: p.category, uom: p.uom, base_price: p.base_price })
    setOpen(true)
  }

  function onSubmit(values: ProductFormValues) {
    const action = editing ? updateProduct.mutateAsync({ id: editing.id, input: values }) : createProduct.mutateAsync(values)
    action
      .then(() => {
        toast.success(editing ? 'Produk diperbarui' : 'Produk dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  function onAddSize(values: SizeFormValues) {
    if (!sizeDialogProductId) return
    addSize
      .mutateAsync({
        productId: sizeDialogProductId,
        input: { size_code: values.size_code.toUpperCase(), size_multiplier: values.size_multiplier, sort_order: Number(values.sort_order ?? 0) },
      })
      .then(() => {
        toast.success('Ukuran ditambahkan')
        sizeForm.reset({ size_code: '', size_multiplier: '1.0', sort_order: '' })
      })
      .catch((err) => toast.error('Gagal menambahkan ukuran', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<Product>[] = [
    { key: 'code', header: 'Kode', render: (p) => p.code, sortValue: (p) => p.code, csvValue: (p) => p.code },
    { key: 'name', header: 'Nama', render: (p) => p.name, sortValue: (p) => p.name, csvValue: (p) => p.name },
    { key: 'category', header: 'Kategori', render: (p) => p.category || '-', csvValue: (p) => p.category },
    {
      key: 'base_price',
      header: 'Harga Dasar',
      render: (p) => formatCurrency(p.base_price),
      sortValue: (p) => Number(p.base_price),
      csvValue: (p) => p.base_price,
    },
    {
      key: 'status',
      header: 'Status',
      render: (p) => <Badge tone={p.is_active ? 'success' : 'neutral'}>{p.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (p) => (p.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
    {
      key: 'sizes',
      header: 'Ukuran',
      render: (p) => (
        <Button
          variant="secondary"
          size="sm"
          onClick={(e) => {
            e.stopPropagation()
            setSizeDialogProductId(p.id)
          }}
        >
          <Ruler className="h-3.5 w-3.5" /> Kelola
        </Button>
      ),
      csvValue: () => '',
    },
  ]

  return (
    <div>
      <PageHeader
        title="Produk"
        description="Produk jersey dan pakaian, dengan rincian ukuran dan pengali biaya."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Produk Baru
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
          exportFilename="products"
          searchPlaceholder="Cari produk..."
          searchFn={(p, q) => p.name.toLowerCase().includes(q) || p.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Produk' : 'Produk Baru'}</DialogTitle>
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
                <Label>Kategori</Label>
                <Input {...register('category')} />
              </div>
              <div className="space-y-1.5">
                <Label>Satuan</Label>
                <Input {...register('uom')} />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Harga Dasar</Label>
              <Input type="number" step="0.01" {...register('base_price')} />
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createProduct.isPending || updateProduct.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!sizeDialogProductId} onOpenChange={(o) => !o && setSizeDialogProductId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Ukuran &middot; {sizeProduct?.name}</DialogTitle>
          </DialogHeader>

          <div className="mb-4 overflow-hidden rounded-md border border-slate-200">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
                <tr>
                  <th className="px-3 py-2">Ukuran</th>
                  <th className="px-3 py-2">Pengali</th>
                  <th className="px-3 py-2">Urutan</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {sizeProduct?.sizes?.length ? (
                  sizeProduct.sizes.map((s) => (
                    <tr key={s.id}>
                      <td className="px-3 py-2 font-medium">{s.size_code}</td>
                      <td className="px-3 py-2">{s.size_multiplier}&times;</td>
                      <td className="px-3 py-2">{s.sort_order}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={3} className="px-3 py-4 text-center text-slate-400">
                      Belum ada ukuran yang ditentukan.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <form onSubmit={sizeForm.handleSubmit(onAddSize)} className="grid grid-cols-3 gap-2">
            <div className="space-y-1.5">
              <Label>Kode Ukuran</Label>
              <Input placeholder="cth. XL" {...sizeForm.register('size_code')} />
            </div>
            <div className="space-y-1.5">
              <Label>Pengali</Label>
              <Input type="number" step="0.01" placeholder="1.2" {...sizeForm.register('size_multiplier')} />
            </div>
            <div className="space-y-1.5">
              <Label>Urutan</Label>
              <Input type="number" placeholder="0" {...sizeForm.register('sort_order')} />
            </div>
            <div className="col-span-3 flex justify-end">
              <Button type="submit" size="sm" loading={addSize.isPending}>
                <Plus className="h-3.5 w-3.5" /> Tambah Ukuran
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
