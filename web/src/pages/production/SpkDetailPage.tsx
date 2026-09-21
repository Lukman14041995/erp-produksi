import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Check, ClipboardList, PackageMinus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { useAddSpkLabor, useAddSpkOverhead, useIssueSpkMaterial, useLogSpkStage, useMarkSpkMilestone, useSpkOrder } from '@/hooks/useSpk'
import { useMaterials } from '@/hooks/useMaterials'
import { formatCurrency, formatDateTime, formatNumber } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { SpkOrder, SpkStage, SpkStageStatus } from '@/types/spk'

const stageStatusTone: Record<SpkStageStatus, BadgeTone> = {
  PENDING: 'neutral',
  IN_PROGRESS: 'info',
  DONE: 'success',
}

const stageStatusLabel: Record<SpkStageStatus, string> = {
  PENDING: 'Belum Mulai',
  IN_PROGRESS: 'Berjalan',
  DONE: 'Selesai',
}

function StageCard({ stage, spkOrderId }: { stage: SpkStage; spkOrderId: string }) {
  const [qty, setQty] = useState('')
  const [note, setNote] = useState('')
  const logStage = useLogSpkStage()
  const markMilestone = useMarkSpkMilestone()

  const remaining = Number(stage.planned_qty) - Number(stage.completed_qty)
  const pct = stage.requires_qty && Number(stage.planned_qty) > 0 ? Math.min(100, Math.round((Number(stage.completed_qty) / Number(stage.planned_qty)) * 100)) : stage.status === 'DONE' ? 100 : 0

  function submitLog() {
    if (!qty || Number(qty) <= 0) return toast.error('Masukkan jumlah yang selesai')
    logStage
      .mutateAsync({ stageId: stage.id, spkOrderId, qty, note })
      .then(() => {
        toast.success(`Progress ${stage.stage_name} dicatat`)
        setQty('')
        setNote('')
      })
      .catch((err) => toast.error('Gagal mencatat progress', err instanceof ApiError ? err.message : undefined))
  }

  function submitMilestone() {
    markMilestone
      .mutateAsync({ stageId: stage.id, spkOrderId, note })
      .then(() => {
        toast.success(`${stage.stage_name} ditandai selesai`)
        setNote('')
      })
      .catch((err) => toast.error('Gagal menandai selesai', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <Card>
      <CardContent className="py-4">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-sm font-semibold text-slate-900">
              {stage.stage_seq}. {stage.stage_name}
            </p>
            {stage.requires_qty ? (
              <p className="mt-0.5 text-xs text-slate-500">
                {formatNumber(stage.completed_qty)} / {formatNumber(stage.planned_qty)} selesai &middot; sisa {formatNumber(String(remaining))}
              </p>
            ) : (
              <p className="mt-0.5 text-xs text-slate-500">Tahap tanpa qty (milestone)</p>
            )}
          </div>
          <Badge tone={stageStatusTone[stage.status]}>{stageStatusLabel[stage.status]}</Badge>
        </div>

        {stage.requires_qty && (
          <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
            <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${pct}%` }} />
          </div>
        )}

        {stage.status !== 'DONE' && (
          <div className="mt-3 space-y-2 border-t border-slate-100 pt-3">
            {stage.requires_qty ? (
              <div className="flex items-end gap-2">
                <div className="flex-1">
                  <Label className="text-xs">Jumlah selesai hari ini</Label>
                  <Input type="number" min="0" max={remaining} value={qty} onChange={(e) => setQty(e.target.value)} />
                </div>
                <div className="flex-1">
                  <Label className="text-xs">Catatan (opsional)</Label>
                  <Input value={note} onChange={(e) => setNote(e.target.value)} />
                </div>
                <Button size="sm" onClick={submitLog} loading={logStage.isPending}>
                  <ClipboardList className="h-3.5 w-3.5" /> Catat
                </Button>
              </div>
            ) : (
              <div className="flex items-end gap-2">
                <div className="flex-1">
                  <Label className="text-xs">Catatan (opsional)</Label>
                  <Input value={note} onChange={(e) => setNote(e.target.value)} />
                </div>
                <Button size="sm" onClick={submitMilestone} loading={markMilestone.isPending}>
                  <Check className="h-3.5 w-3.5" /> Tandai Selesai
                </Button>
              </div>
            )}
          </div>
        )}

        {!!stage.logs?.length && (
          <div className="mt-3 space-y-1 border-t border-slate-100 pt-3">
            {stage.logs.map((l) => (
              <div key={l.id} className="flex items-center justify-between text-xs text-slate-400">
                <span>
                  {l.logged_by_name || 'Tim Produksi'}
                  {Number(l.qty) > 0 && <> &middot; +{formatNumber(l.qty)}</>}
                  {l.note && <> &middot; {l.note}</>}
                </span>
                <span>{formatDateTime(l.logged_at)}</span>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

const NO_STAGE = '__none__'

function MaterialUsageSection({ order }: { order: SpkOrder }) {
  const { data: materials } = useMaterials()
  const issueMaterial = useIssueSpkMaterial()
  const [materialId, setMaterialId] = useState('')
  const [stageId, setStageId] = useState(NO_STAGE)
  const [qty, setQty] = useState('')
  const [note, setNote] = useState('')

  const stageName = (id?: string) => order.stages?.find((s) => s.id === id)?.stage_name

  // Bahan yang sudah dipilih customer (fabric/ink dari konfigurator) --
  // muncul duluan sebagai rekomendasi, tapi bahan lain (plastik packing,
  // benang, dst.) tetap bisa dipilih karena masih dibutuhkan produksi.
  const recommendedIds = new Set(
    (order.items ?? []).flatMap((it) => [it.fabric_material_id, it.ink_material_id]).filter((id): id is string => !!id),
  )
  const recommended = materials?.filter((m) => recommendedIds.has(m.id)) ?? []
  const others = materials?.filter((m) => !recommendedIds.has(m.id)) ?? []

  function submit() {
    if (!materialId) return toast.error('Pilih bahan baku')
    if (!qty || Number(qty) <= 0) return toast.error('Masukkan jumlah yang dikeluarkan')
    issueMaterial
      .mutateAsync({ spkOrderId: order.id, material_id: materialId, qty, note, spk_stage_id: stageId === NO_STAGE ? undefined : stageId })
      .then(() => {
        toast.success('Bahan baku dikeluarkan dari gudang')
        setMaterialId('')
        setQty('')
        setNote('')
        setStageId(NO_STAGE)
      })
      .catch((err) => toast.error('Gagal mengeluarkan bahan', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <Card className="mt-6">
      <CardHeader>
        <CardTitle>Bahan Baku Terpakai</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="mb-4 grid grid-cols-1 gap-2 md:grid-cols-4">
          <div className="md:col-span-1">
            <Label className="text-xs">Bahan Baku</Label>
            <Select value={materialId} onValueChange={setMaterialId}>
              <SelectTrigger>
                <SelectValue placeholder="Pilih bahan" />
              </SelectTrigger>
              <SelectContent>
                {recommended.length > 0 && (
                  <>
                    <div className="px-2 pb-1 pt-1.5 text-[11px] font-semibold uppercase text-slate-400">Dipilih Customer</div>
                    {recommended.map((m) => (
                      <SelectItem key={m.id} value={m.id}>
                        {m.code} &ndash; {m.name} ({m.uom})
                      </SelectItem>
                    ))}
                    <div className="px-2 pb-1 pt-2 text-[11px] font-semibold uppercase text-slate-400">Bahan Lainnya</div>
                  </>
                )}
                {others.map((m) => (
                  <SelectItem key={m.id} value={m.id}>
                    {m.code} &ndash; {m.name} ({m.uom})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label className="text-xs">Tahap Terkait (opsional)</Label>
            <Select
              value={stageId}
              onValueChange={(v) => {
                setStageId(v)
                // Auto-pilih bahan yang dipilih customer sesuai tahapnya --
                // Cutting -> kain, Sablon -> tinta -- tim produksi tinggal cek, bukan cari manual.
                const stage = order.stages?.find((s) => s.id === v)
                const firstItem = order.items?.[0]
                if (stage?.stage_code === 'CUTTING' && firstItem?.fabric_material_id) setMaterialId(firstItem.fabric_material_id)
                else if (stage?.stage_code === 'PRINTING' && firstItem?.ink_material_id) setMaterialId(firstItem.ink_material_id)
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Umum" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={NO_STAGE}>Umum (tidak terkait tahap)</SelectItem>
                {order.stages?.map((s) => (
                  <SelectItem key={s.id} value={s.id}>
                    {s.stage_seq}. {s.stage_name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label className="text-xs">Jumlah</Label>
            <Input type="number" min="0" step="0.0001" value={qty} onChange={(e) => setQty(e.target.value)} />
          </div>
          <div className="flex items-end gap-2">
            <div className="flex-1">
              <Label className="text-xs">Catatan (opsional)</Label>
              <Input value={note} onChange={(e) => setNote(e.target.value)} />
            </div>
            <Button size="sm" onClick={submit} loading={issueMaterial.isPending}>
              <PackageMinus className="h-3.5 w-3.5" /> Keluarkan
            </Button>
          </div>
        </div>

        {order.material_usages?.length ? (
          <table className="w-full text-sm">
            <thead className="text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="py-1.5">Bahan</th>
                <th className="py-1.5">Tahap</th>
                <th className="py-1.5">Jumlah</th>
                <th className="py-1.5">Harga Modal</th>
                <th className="py-1.5">Total Biaya</th>
                <th className="py-1.5">Oleh</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {order.material_usages.map((u) => (
                <tr key={u.id}>
                  <td className="py-1.5">{u.material_name}</td>
                  <td className="py-1.5 text-slate-500">{stageName(u.spk_stage_id) ?? 'Umum'}</td>
                  <td className="py-1.5">
                    {formatNumber(u.qty, 4)} {u.material_uom}
                  </td>
                  <td className="py-1.5">{formatCurrency(u.unit_cost)}</td>
                  <td className="py-1.5 font-medium">{formatCurrency(u.total_cost)}</td>
                  <td className="py-1.5 text-xs text-slate-400">
                    {u.logged_by_name || 'Tim Produksi'} &middot; {formatDateTime(u.logged_at)}
                  </td>
                </tr>
              ))}
              <tr className="border-t-2 border-slate-200">
                <td colSpan={4} className="py-1.5 text-right font-medium text-slate-500">
                  Total Biaya Bahan
                </td>
                <td className="py-1.5 font-semibold">{formatCurrency(order.material_cost_total)}</td>
                <td />
              </tr>
            </tbody>
          </table>
        ) : (
          <p className="text-sm text-slate-400">Belum ada bahan baku yang dikeluarkan untuk SPK ini.</p>
        )}
      </CardContent>
    </Card>
  )
}

function HppSummaryCard({ order }: { order: SpkOrder }) {
  const total = Number(order.material_cost_total) + Number(order.labor_cost_total) + Number(order.overhead_cost_total)
  const perPcs = Number(order.total_qty) > 0 ? total / Number(order.total_qty) : 0
  return (
    <Card className="mb-6">
      <CardHeader>
        <CardTitle>Ringkasan HPP</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-2 gap-4 lg:grid-cols-5">
        <div>
          <p className="text-xs uppercase text-slate-400">Biaya Bahan</p>
          <p className="mt-1 text-lg font-semibold">{formatCurrency(order.material_cost_total)}</p>
        </div>
        <div>
          <p className="text-xs uppercase text-slate-400">Tenaga Kerja</p>
          <p className="mt-1 text-lg font-semibold">{formatCurrency(order.labor_cost_total)}</p>
        </div>
        <div>
          <p className="text-xs uppercase text-slate-400">Overhead</p>
          <p className="mt-1 text-lg font-semibold">{formatCurrency(order.overhead_cost_total)}</p>
        </div>
        <div>
          <p className="text-xs uppercase text-slate-400">Total HPP</p>
          <p className="mt-1 text-lg font-semibold">{formatCurrency(total)}</p>
        </div>
        <div>
          <p className="text-xs uppercase text-slate-400">HPP / pcs</p>
          <p className="mt-1 text-lg font-semibold text-[var(--color-primary)]">{formatCurrency(perPcs)}</p>
        </div>
      </CardContent>
    </Card>
  )
}

function LaborOverheadSection({ order }: { order: SpkOrder }) {
  const addLabor = useAddSpkLabor()
  const addOverhead = useAddSpkOverhead()
  const [laborDesc, setLaborDesc] = useState('')
  const [laborHours, setLaborHours] = useState('')
  const [laborRate, setLaborRate] = useState('')
  const [laborStageId, setLaborStageId] = useState(NO_STAGE)
  const [overheadDesc, setOverheadDesc] = useState('')
  const [overheadBasis, setOverheadBasis] = useState('')
  const [overheadAmount, setOverheadAmount] = useState('')

  const stageName = (id?: string) => order.stages?.find((s) => s.id === id)?.stage_name

  function submitLabor() {
    if (!laborDesc || !laborHours || !laborRate) return toast.error('Isi semua kolom tenaga kerja')
    addLabor
      .mutateAsync({
        spkOrderId: order.id, description: laborDesc, hours: laborHours, rate: laborRate,
        spk_stage_id: laborStageId === NO_STAGE ? undefined : laborStageId,
      })
      .then(() => {
        toast.success('Tenaga kerja dicatat')
        setLaborDesc('')
        setLaborHours('')
        setLaborRate('')
        setLaborStageId(NO_STAGE)
      })
      .catch((err) => toast.error('Gagal mencatat tenaga kerja', err instanceof ApiError ? err.message : undefined))
  }

  function submitOverhead() {
    if (!overheadDesc || !overheadAmount) return toast.error('Isi semua kolom overhead')
    addOverhead
      .mutateAsync({ spkOrderId: order.id, description: overheadDesc, allocation_basis: overheadBasis, amount: overheadAmount })
      .then(() => {
        toast.success('Overhead dicatat')
        setOverheadDesc('')
        setOverheadBasis('')
        setOverheadAmount('')
      })
      .catch((err) => toast.error('Gagal mencatat overhead', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Tenaga Kerja</CardTitle>
        </CardHeader>
        <CardContent>
          <table className="mb-4 w-full text-sm">
            <tbody className="divide-y divide-slate-100">
              {order.labor?.map((l) => (
                <tr key={l.id}>
                  <td className="py-1.5">
                    {l.description}
                    {l.spk_stage_id && stageName(l.spk_stage_id) && <span className="text-slate-400"> &middot; {stageName(l.spk_stage_id)}</span>}
                  </td>
                  <td className="py-1.5 text-right text-slate-500">
                    {l.hours} jam &times; {formatCurrency(l.rate)}
                  </td>
                  <td className="py-1.5 text-right font-medium">{formatCurrency(l.total_cost)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="space-y-2">
            <Input placeholder="Deskripsi (mis. Jahit lengan)" value={laborDesc} onChange={(e) => setLaborDesc(e.target.value)} />
            <Select value={laborStageId} onValueChange={setLaborStageId}>
              <SelectTrigger>
                <SelectValue placeholder="Umum" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={NO_STAGE}>Umum (tidak terkait tahap)</SelectItem>
                {order.stages?.map((s) => (
                  <SelectItem key={s.id} value={s.id}>
                    {s.stage_seq}. {s.stage_name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <div className="flex gap-2">
              <Input type="number" step="0.01" placeholder="Jam" value={laborHours} onChange={(e) => setLaborHours(e.target.value)} />
              <Input type="number" step="1" placeholder="Tarif/jam" value={laborRate} onChange={(e) => setLaborRate(e.target.value)} />
            </div>
            <Button size="sm" onClick={submitLabor} loading={addLabor.isPending}>
              Tambah Tenaga Kerja
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Overhead Pabrik</CardTitle>
        </CardHeader>
        <CardContent>
          <table className="mb-4 w-full text-sm">
            <tbody className="divide-y divide-slate-100">
              {order.overheads?.map((o) => (
                <tr key={o.id}>
                  <td className="py-1.5">
                    {o.description}
                    {o.allocation_basis && <span className="text-slate-400"> &middot; {o.allocation_basis}</span>}
                  </td>
                  <td className="py-1.5 text-right font-medium">{formatCurrency(o.amount)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="space-y-2">
            <Input placeholder="Deskripsi (mis. Listrik mesin jahit)" value={overheadDesc} onChange={(e) => setOverheadDesc(e.target.value)} />
            <Input placeholder="Dasar alokasi (opsional)" value={overheadBasis} onChange={(e) => setOverheadBasis(e.target.value)} />
            <Input type="number" step="1" placeholder="Jumlah" value={overheadAmount} onChange={(e) => setOverheadAmount(e.target.value)} />
            <Button size="sm" onClick={submitOverhead} loading={addOverhead.isPending}>
              Tambah Overhead
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export function SpkDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: order, isLoading, error } = useSpkOrder(id)

  if (isLoading) return <LoadingState />
  if (error || !order) return <ErrorState message={(error as Error)?.message ?? 'SPK tidak ditemukan'} />

  const backTo = order.product_type_code === 'JERSEY' ? '/production/orders/jersey' : '/production/orders/tshirt'
  const pct = Math.min(100, Math.round(Number(order.progress_pct)))

  return (
    <div>
      <button onClick={() => navigate(backTo)} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke {order.product_type_name}
      </button>

      <PageHeader
        title={order.spk_number}
        description={`${order.quotation_number ? order.quotation_number + ' · ' : ''}${order.so_number} · ${order.customer_name} · ${order.product_type_name} · ${formatNumber(order.total_qty)} pcs`}
      />

      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Spesifikasi Pesanan</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Bahan</th>
                <th className="px-4 py-2.5">Model</th>
                <th className="px-4 py-2.5">Tinta</th>
                <th className="px-4 py-2.5">Ukuran</th>
                <th className="px-4 py-2.5">Qty</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {order.items?.map((it) => (
                <tr key={it.id}>
                  <td className="px-4 py-2.5">{it.fabric_name || '-'}</td>
                  <td className="px-4 py-2.5">{it.variant_name || '-'}</td>
                  <td className="px-4 py-2.5">{it.ink_name || '-'}</td>
                  <td className="px-4 py-2.5">{it.size_code || '-'}</td>
                  <td className="px-4 py-2.5 font-medium">{formatNumber(it.qty)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      <HppSummaryCard order={order} />

      <Card className="mb-6">
        <CardContent className="py-4">
          <div className="flex items-center justify-between text-sm">
            <span className="font-medium text-slate-700">Progress Keseluruhan</span>
            <span className="text-slate-500">{pct}%</span>
          </div>
          <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-slate-100">
            <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${pct}%` }} />
          </div>
        </CardContent>
      </Card>

      <div className="space-y-3">
        {order.stages?.map((stage) => (
          <StageCard key={stage.id} stage={stage} spkOrderId={order.id} />
        ))}
      </div>

      <MaterialUsageSection order={order} />
      <LaborOverheadSection order={order} />
    </div>
  )
}
