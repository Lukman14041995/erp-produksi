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
import { useCreateMaterial, useMaterials, useUpdateMaterial } from '@/hooks/useMaterials'
import { formatCurrency } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Material } from '@/types/master'

const schema = z.object({
  code: z.string().min(1, 'Wajib diisi'),
  name: z.string().min(1, 'Wajib diisi'),
  uom: z.string().min(1, 'Wajib diisi'),
  unit_cost: z.string().min(1, 'Wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function MaterialsPage() {
  const { data, isLoading, error } = useMaterials()
  const createMaterial = useCreateMaterial()
  const updateMaterial = useUpdateMaterial()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Material | null>(null)

  const { register, handleSubmit, reset } = useForm<FormValues>({ resolver: zodResolver(schema) })

  function openCreate() {
    setEditing(null)
    reset({ code: '', name: '', uom: 'PCS', unit_cost: '0' })
    setOpen(true)
  }

  function openEdit(m: Material) {
    setEditing(m)
    reset({ code: m.code, name: m.name, uom: m.uom, unit_cost: m.unit_cost })
    setOpen(true)
  }

  function onSubmit(values: FormValues) {
    const action = editing ? updateMaterial.mutateAsync({ id: editing.id, input: values }) : createMaterial.mutateAsync(values)
    action
      .then(() => {
        toast.success(editing ? 'Bahan baku diperbarui' : 'Bahan baku dibuat')
        setOpen(false)
      })
      .catch((err) => toast.error('Gagal menyimpan', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<Material>[] = [
    { key: 'code', header: 'Kode', render: (m) => m.code, sortValue: (m) => m.code, csvValue: (m) => m.code },
    { key: 'name', header: 'Nama', render: (m) => m.name, sortValue: (m) => m.name, csvValue: (m) => m.name },
    { key: 'uom', header: 'Satuan', render: (m) => m.uom, csvValue: (m) => m.uom },
    {
      key: 'unit_cost',
      header: 'Biaya Satuan',
      render: (m) => formatCurrency(m.unit_cost),
      sortValue: (m) => Number(m.unit_cost),
      csvValue: (m) => m.unit_cost,
    },
    {
      key: 'status',
      header: 'Status',
      render: (m) => <Badge tone={m.is_active ? 'success' : 'neutral'}>{m.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (m) => (m.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Bahan Baku"
        description="Bahan baku yang digunakan dalam bill of materials dan persediaan."
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> Bahan Baku Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(m) => m.id}
          exportFilename="materials"
          searchPlaceholder="Cari bahan baku..."
          searchFn={(m, q) => m.name.toLowerCase().includes(q) || m.code.toLowerCase().includes(q)}
          onRowClick={openEdit}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Ubah Bahan Baku' : 'Bahan Baku Baru'}</DialogTitle>
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
                <Label>Satuan</Label>
                <Input {...register('uom')} />
              </div>
              <div className="space-y-1.5">
                <Label>Biaya Satuan</Label>
                <Input type="number" step="0.0001" {...register('unit_cost')} />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={createMaterial.isPending || updateMaterial.isPending}>
                Simpan
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
