import { useMemo, useState } from 'react'
import { ArrowLeft, Check, Shirt, ShoppingCart, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Input, Textarea } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { useSubmitQuotation } from '@/hooks/useQuotations'
import { formatCurrency, toUploadUrl } from '@/lib/format'
import { uid } from '@/lib/uid'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { ItemInput, PublicCatalog } from '@/types/quotation'

// Unambiguous "selected" mark shown on top of an option card -- a border
// color change alone was too subtle against a photo or a dark card
// background, so every pickable option in this wizard gets this badge too.
function SelectedBadge() {
  return (
    <span className="absolute right-1.5 top-1.5 z-10 flex h-5 w-5 items-center justify-center rounded-full bg-[var(--color-primary)] text-white shadow-md">
      <Check className="h-3.5 w-3.5" strokeWidth={3} />
    </span>
  )
}

interface CartLine extends ItemInput {
  key: string
  jenisName: string
  modelName: string
  fabricName: string
  inkName: string
  sizeCode: string
  unitPrice: number
  lineTotal: number
}

export function NewOrderWizard({ token, catalog, onBack, onSubmitted }: {
  token: string
  catalog: PublicCatalog
  onBack: () => void
  onSubmitted: () => void
}) {
  const submitQuotation = useSubmitQuotation(token)

  const [productTypeId, setProductTypeId] = useState<string | null>(null)
  const [variantId, setVariantId] = useState<string | null>(null)
  const [fabricId, setFabricId] = useState<string | null>(null)
  const [inkId, setInkId] = useState<string | null>(null)
  const [qtyBySize, setQtyBySize] = useState<Record<string, string>>({})
  const [cart, setCart] = useState<CartLine[]>([])
  const [notes, setNotes] = useState('')
  const [submitted, setSubmitted] = useState<{ number: string; total: string } | null>(null)

  const fabricsForType = catalog.fabrics.filter((f) => f.product_type_id === productTypeId)
  const variantsForType = catalog.variants.filter((v) => v.product_type_id === productTypeId)
  const inksForType = catalog.inks.filter((i) => i.product_type_id === productTypeId)

  const selectedType = catalog.product_types.find((t) => t.id === productTypeId)
  const selectedVariant = catalog.variants.find((v) => v.id === variantId)
  const selectedFabric = catalog.fabrics.find((f) => f.id === fabricId)
  const selectedInk = catalog.inks.find((i) => i.id === inkId)

  function resetFrom(step: 'jenis' | 'model' | 'bahan' | 'tinta') {
    if (step === 'jenis') {
      setVariantId(null)
      setFabricId(null)
      setInkId(null)
    } else if (step === 'model') {
      setFabricId(null)
      setInkId(null)
    } else if (step === 'bahan') {
      setInkId(null)
    }
    setQtyBySize({})
  }

  function pickJenis(id: string) {
    setProductTypeId(id)
    resetFrom('jenis')
  }
  function pickModel(id: string) {
    setVariantId(id)
    resetFrom('model')
  }
  function pickBahan(id: string) {
    setFabricId(id)
    resetFrom('bahan')
  }
  function pickTinta(id: string) {
    setInkId(id)
    setQtyBySize({})
  }

  const builderLines = useMemo(() => {
    if (!selectedType || !selectedVariant || !selectedFabric || !selectedInk) return []
    const lines: CartLine[] = []
    for (const size of catalog.sizes) {
      const qty = Number(qtyBySize[size.id] ?? 0)
      if (!qty || qty <= 0) continue
      const unitPrice = Number(selectedFabric.sales_price) * Number(size.size_multiplier) + Number(selectedVariant.price_addon) + Number(selectedInk.price_addon)
      lines.push({
        key: uid(),
        product_type_id: productTypeId!,
        fabric_id: fabricId!,
        garment_size_id: size.id,
        variant_id: variantId!,
        ink_id: inkId!,
        qty: String(qty),
        jenisName: selectedType.name,
        modelName: selectedVariant.name,
        fabricName: selectedFabric.name,
        inkName: selectedInk.name,
        sizeCode: size.size_code,
        unitPrice,
        lineTotal: unitPrice * qty,
      })
    }
    return lines
  }, [catalog.sizes, selectedType, selectedVariant, selectedFabric, selectedInk, productTypeId, variantId, fabricId, inkId, qtyBySize])

  const builderSubtotal = builderLines.reduce((sum, l) => sum + l.lineTotal, 0)
  const cartSubtotal = cart.reduce((sum, l) => sum + l.lineTotal, 0)
  const grandTotal = cartSubtotal + builderSubtotal

  function addToCart() {
    if (builderLines.length === 0) return
    setCart((prev) => [...prev, ...builderLines])
    setProductTypeId(null)
    setVariantId(null)
    setFabricId(null)
    setInkId(null)
    setQtyBySize({})
  }

  function removeLine(key: string) {
    setCart((prev) => prev.filter((l) => l.key !== key))
  }

  function submit() {
    const allLines = [...cart, ...builderLines]
    if (allLines.length === 0) return
    submitQuotation
      .mutateAsync({
        notes,
        items: allLines.map(({ product_type_id, fabric_id, garment_size_id, variant_id, ink_id, qty }) => ({
          product_type_id, fabric_id, garment_size_id, variant_id, ink_id, qty,
        })),
      })
      .then((q) => setSubmitted({ number: q.quotation_number, total: q.estimated_subtotal }))
      .catch((err) => toast.error('Gagal mengirim pesanan', err instanceof ApiError ? err.message : undefined))
  }

  if (submitted) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-full bg-[var(--color-success-surface)] text-[var(--color-success)]">
          <ShoppingCart className="h-7 w-7" />
        </div>
        <p className="text-lg font-semibold text-slate-900">Pesanan Anda telah diterima</p>
        <p className="text-sm text-slate-500">
          Nomor referensi <span className="font-medium text-slate-700">{submitted.number}</span>
        </p>
        <p className="text-sm text-slate-500">Estimasi total {formatCurrency(submitted.total)}</p>
        <p className="mt-2 max-w-sm text-xs text-slate-400">Tim kami akan menghubungi Anda untuk konfirmasi harga final dan detail pembayaran.</p>
        <Button className="mt-3" onClick={onSubmitted}>
          Lihat Pesanan Saya
        </Button>
      </div>
    )
  }

  return (
    <div>
      <button onClick={onBack} className="mb-4 flex items-center gap-1 text-sm text-slate-400 hover:text-slate-200">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Pesanan Saya
      </button>

      {/* Step 1: Jenis */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle>1. Pilih Jenis Pesanan</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-3 sm:w-96">
            {catalog.product_types.map((t) => (
              <button
                key={t.id}
                onClick={() => pickJenis(t.id)}
                className={`rounded-lg border-2 p-4 text-center text-sm font-semibold transition-colors ${
                  productTypeId === t.id
                    ? 'border-[var(--color-primary)] bg-[var(--color-primary)] text-[var(--color-primary-foreground)] shadow-md'
                    : 'border-slate-200 text-slate-700 hover:border-slate-300'
                }`}
              >
                {t.name}
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Step 2: Model (icon) */}
      {productTypeId && (
        <Card className="mb-4">
          <CardHeader>
            <CardTitle>2. Pilih Model</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-3">
              {variantsForType.map((v) => (
                <button
                  key={v.id}
                  onClick={() => pickModel(v.id)}
                  className={`relative flex w-24 flex-col items-center gap-1.5 rounded-lg border-2 p-3 transition-colors ${
                    variantId === v.id ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/15 shadow-md' : 'border-slate-200 hover:border-slate-300'
                  }`}
                >
                  {variantId === v.id && <SelectedBadge />}
                  {v.icon_url ? (
                    <img src={toUploadUrl(v.icon_url)} alt={v.name} className="h-8 w-8 object-contain" />
                  ) : (
                    <Shirt className="h-8 w-8 text-slate-300" />
                  )}
                  <span className="text-xs font-medium text-slate-700">{v.name}</span>
                </button>
              ))}
              {variantsForType.length === 0 && <p className="text-sm text-slate-400">Belum ada pilihan model untuk jenis ini.</p>}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Step 3: Bahan (photo) */}
      {variantId && (
        <Card className="mb-4">
          <CardHeader>
            <CardTitle>3. Pilih Bahan</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              {fabricsForType.map((f) => (
                <button
                  key={f.id}
                  onClick={() => pickBahan(f.id)}
                  className={`relative overflow-hidden rounded-lg border-2 text-left transition-colors ${
                    fabricId === f.id ? 'border-[var(--color-primary)] shadow-md shadow-[var(--color-primary)]/20' : 'border-transparent hover:border-slate-200'
                  }`}
                >
                  {fabricId === f.id && <SelectedBadge />}
                  {f.image_url ? (
                    <img src={toUploadUrl(f.image_url)} alt={f.name} className="aspect-square w-full object-cover" />
                  ) : (
                    <div className="flex aspect-square w-full items-center justify-center bg-slate-100">
                      <Shirt className="h-8 w-8 text-slate-300" />
                    </div>
                  )}
                  <div className="px-2 py-1.5">
                    <p className="text-xs font-medium text-slate-700">{f.name}</p>
                    <p className="text-[11px] text-slate-400">{formatCurrency(f.sales_price)}/pcs</p>
                  </div>
                </button>
              ))}
              {fabricsForType.length === 0 && <p className="text-sm text-slate-400">Belum ada pilihan bahan untuk jenis ini.</p>}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Step 4: Tinta (photo) */}
      {fabricId && (
        <Card className="mb-4">
          <CardHeader>
            <CardTitle>4. Pilih Tinta</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              {inksForType.map((i) => (
                <button
                  key={i.id}
                  onClick={() => pickTinta(i.id)}
                  className={`relative overflow-hidden rounded-lg border-2 text-left transition-colors ${
                    inkId === i.id ? 'border-[var(--color-primary)] shadow-md shadow-[var(--color-primary)]/20' : 'border-transparent hover:border-slate-200'
                  }`}
                >
                  {inkId === i.id && <SelectedBadge />}
                  {i.image_url ? (
                    <img src={toUploadUrl(i.image_url)} alt={i.name} className="aspect-square w-full object-cover" />
                  ) : (
                    <div className="flex aspect-square w-full items-center justify-center bg-slate-100">
                      <Shirt className="h-8 w-8 text-slate-300" />
                    </div>
                  )}
                  <div className="px-2 py-1.5">
                    <p className="text-xs font-medium text-slate-700">{i.name}</p>
                    {Number(i.price_addon) > 0 && <p className="text-[11px] text-slate-400">+{formatCurrency(i.price_addon)}/pcs</p>}
                  </div>
                </button>
              ))}
              {inksForType.length === 0 && <p className="text-sm text-slate-400">Belum ada pilihan tinta untuk jenis ini.</p>}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Step 5: Ukuran + qty */}
      {inkId && (
        <Card className="mb-4">
          <CardHeader>
            <CardTitle>5. Jumlah per Ukuran</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              {catalog.sizes.map((s) => (
                <div key={s.id} className="space-y-1">
                  <Label className="text-xs">{s.size_code}</Label>
                  <Input
                    type="number"
                    min={0}
                    value={qtyBySize[s.id] ?? ''}
                    onChange={(e) => setQtyBySize((prev) => ({ ...prev, [s.id]: e.target.value }))}
                    placeholder="0"
                  />
                </div>
              ))}
            </div>

            <div className="mt-4 flex items-center justify-between rounded-md bg-slate-50 px-3 py-2">
              <span className="text-sm text-slate-500">Subtotal kombinasi ini</span>
              <span className="text-sm font-semibold">{formatCurrency(builderSubtotal)}</span>
            </div>
            <Button className="mt-3" variant="secondary" onClick={addToCart} disabled={builderLines.length === 0}>
              <ShoppingCart className="h-4 w-4" /> Tambahkan ke Pesanan
            </Button>
          </CardContent>
        </Card>
      )}

      {cart.length > 0 && (
        <Card className="mb-4">
          <CardHeader>
            <CardTitle>Ringkasan Pesanan</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {cart.map((l) => (
              <div key={l.key} className="flex items-center justify-between rounded-md border border-slate-100 px-3 py-2 text-sm">
                <div>
                  <p className="font-medium text-slate-800">
                    {l.jenisName} &middot; {l.modelName} &middot; {l.fabricName} &middot; {l.inkName}
                  </p>
                  <p className="text-xs text-slate-400">
                    {l.sizeCode} &middot; {l.qty} pcs &times; {formatCurrency(l.unitPrice)}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="font-medium">{formatCurrency(l.lineTotal)}</span>
                  <button onClick={() => removeLine(l.key)} className="text-slate-400 hover:text-[var(--color-danger)]">
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      <Card className="mb-4">
        <CardHeader>
          <CardTitle>Catatan (opsional)</CardTitle>
        </CardHeader>
        <CardContent>
          <Label className="sr-only">Catatan</Label>
          <Textarea rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="Contoh: nama punggung, nomor, tanggal butuh selesai" />
        </CardContent>
      </Card>

      <div className="sticky bottom-4 flex items-center justify-between rounded-lg border border-slate-200 bg-[var(--color-surface)] px-5 py-4 shadow-lg">
        <div>
          <p className="text-xs text-slate-400">Estimasi Total</p>
          <p className="text-lg font-semibold text-slate-900">{formatCurrency(grandTotal)}</p>
        </div>
        <Button size="lg" onClick={submit} loading={submitQuotation.isPending} disabled={grandTotal <= 0}>
          Kirim Pesanan
        </Button>
      </div>
    </div>
  )
}
