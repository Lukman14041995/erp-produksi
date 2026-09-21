import { useState } from 'react'
import { Copy, Link2, ShieldOff } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { LoadingState } from '@/components/QueryState'
import { useCustomers } from '@/hooks/useCustomers'
import { useCreateOrderLink, useOrderLinks, useRevokeOrderLink } from '@/hooks/useQuotations'
import { formatDateTime } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

export function OrderLinksPage() {
  const { data: customers } = useCustomers()
  const [customerId, setCustomerId] = useState('')
  const { data: links, isLoading } = useOrderLinks(customerId || undefined)
  const createLink = useCreateOrderLink()
  const revokeLink = useRevokeOrderLink()
  const [freshLink, setFreshLink] = useState<string | null>(null)

  function generateLink() {
    if (!customerId) return
    createLink
      .mutateAsync({ customer_id: customerId })
      .then((out) => setFreshLink(`${window.location.origin}/order/${out.token}`))
      .catch((err) => toast.error('Gagal membuat link', err instanceof ApiError ? err.message : undefined))
  }

  function copyLink(url: string) {
    navigator.clipboard.writeText(url).then(() => toast.success('Link disalin'))
  }

  return (
    <div>
      <PageHeader
        title="Link Pesanan Customer"
        description="Buat link unik per pelanggan agar mereka bisa memilih desain, bahan, ukuran, dan varian sendiri."
      />

      <Card className="mb-4">
        <CardContent className="flex flex-wrap items-end gap-3 py-4">
          <div className="w-64">
            <Select
              value={customerId}
              onValueChange={(v) => {
                setCustomerId(v)
                setFreshLink(null)
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Pilih pelanggan" />
              </SelectTrigger>
              <SelectContent>
                {customers?.map((c) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.code} &ndash; {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button onClick={generateLink} disabled={!customerId} loading={createLink.isPending}>
            <Link2 className="h-4 w-4" /> Buat Link Baru
          </Button>
        </CardContent>
      </Card>

      {freshLink && (
        <Card className="mb-4 border-[var(--color-primary)]">
          <CardContent className="flex flex-wrap items-center justify-between gap-2 py-4">
            <div>
              <p className="text-sm font-medium text-slate-900">Link berhasil dibuat</p>
              <p className="text-xs text-slate-500">
                Kirim link ini ke pelanggan. Token hanya ditampilkan sekali &mdash; simpan sekarang.
              </p>
              <p className="mt-1 break-all text-sm text-[var(--color-primary)]">{freshLink}</p>
            </div>
            <Button variant="secondary" size="sm" onClick={() => copyLink(freshLink)}>
              <Copy className="h-4 w-4" /> Salin
            </Button>
          </CardContent>
        </Card>
      )}

      {!customerId && <p className="text-sm text-slate-400">Pilih pelanggan untuk melihat daftar link.</p>}
      {customerId && isLoading && <LoadingState />}
      {customerId && links && links.length === 0 && <p className="text-sm text-slate-400">Belum ada link untuk pelanggan ini.</p>}
      {customerId && links && links.length > 0 && (
        <div className="space-y-2">
          {links.map((link) => {
            const expired = new Date(link.expires_at).getTime() < Date.now()
            return (
              <Card key={link.id}>
                <CardContent className="flex items-center justify-between py-3">
                  <div className="text-sm">
                    <p className="text-slate-500">Dibuat {formatDateTime(link.created_at)}</p>
                    <p className="text-xs text-slate-400">Kedaluwarsa {formatDateTime(link.expires_at)}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    {link.status === 'REVOKED' ? (
                      <Badge tone="danger">Dicabut</Badge>
                    ) : expired ? (
                      <Badge tone="neutral">Kedaluwarsa</Badge>
                    ) : (
                      <Badge tone="success">Aktif</Badge>
                    )}
                    {link.status === 'ACTIVE' && !expired && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          revokeLink
                            .mutateAsync(link.id)
                            .then(() => toast.success('Link dicabut'))
                            .catch((err) => toast.error('Gagal mencabut link', err instanceof ApiError ? err.message : undefined))
                        }
                      >
                        <ShieldOff className="h-4 w-4" /> Cabut
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
