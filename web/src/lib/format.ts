export function formatCurrency(value: string | number): string {
  const n = typeof value === 'string' ? Number(value) : value
  if (Number.isNaN(n)) return '-'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(n)
}

export function formatNumber(value: string | number, fractionDigits = 0): string {
  const n = typeof value === 'string' ? Number(value) : value
  if (Number.isNaN(n)) return '-'
  return new Intl.NumberFormat('id-ID', {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(n)
}

export function formatDate(value: string | Date | null | undefined): string {
  if (!value) return '-'
  const d = typeof value === 'string' ? new Date(value) : value
  if (Number.isNaN(d.getTime())) return '-'
  return new Intl.DateTimeFormat('id-ID', { day: '2-digit', month: 'short', year: 'numeric' }).format(d)
}

export function formatDateTime(value: string | Date | null | undefined): string {
  if (!value) return '-'
  const d = typeof value === 'string' ? new Date(value) : value
  if (Number.isNaN(d.getTime())) return '-'
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
  }).format(d)
}

export function toISODate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

// toUploadUrl resolves a server-relative upload path (e.g.
// "/uploads/designs/xyz.png") against the API's origin, since uploaded
// files are served by the backend itself (see e.Static in cmd/api/main.go
// and the matching nginx /uploads/ proxy) rather than by Vite/the SPA host.
export function toUploadUrl(path: string): string {
  const apiUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1'
  const origin = apiUrl.replace(/\/api\/v1\/?$/, '')
  return origin + path
}
