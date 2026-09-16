import { Badge, type BadgeTone } from '@/components/ui/Badge'

type StatusMap = Record<string, { label: string; tone: BadgeTone }>

const orderStatusMap: StatusMap = {
  DRAFT: { label: 'Draf', tone: 'neutral' },
  CONFIRMED: { label: 'Dikonfirmasi', tone: 'info' },
  CANCELLED: { label: 'Dibatalkan', tone: 'danger' },
  CLOSED: { label: 'Ditutup', tone: 'success' },
}

const paymentStatusMap: StatusMap = {
  UNPAID: { label: 'Belum Lunas', tone: 'danger' },
  PARTIAL: { label: 'Sebagian', tone: 'warning' },
  PAID: { label: 'Lunas', tone: 'success' },
  OVERPAID: { label: 'Kelebihan Bayar', tone: 'info' },
}

const productionStatusMap: StatusMap = {
  NOT_STARTED: { label: 'Belum Dimulai', tone: 'neutral' },
  IN_PROGRESS: { label: 'Sedang Berjalan', tone: 'info' },
  COMPLETED: { label: 'Selesai', tone: 'success' },
  CANCELLED: { label: 'Dibatalkan', tone: 'danger' },
}

const deliveryStatusMap: StatusMap = {
  NOT_DELIVERED: { label: 'Belum Dikirim', tone: 'neutral' },
  PARTIAL: { label: 'Sebagian', tone: 'warning' },
  DELIVERED: { label: 'Terkirim', tone: 'success' },
}

const invoiceStatusMap: StatusMap = {
  DRAFT: { label: 'Draf', tone: 'neutral' },
  POSTED: { label: 'Diposting', tone: 'info' },
  PARTIALLY_PAID: { label: 'Dibayar Sebagian', tone: 'warning' },
  PAID: { label: 'Lunas', tone: 'success' },
  VOID: { label: 'Dibatalkan', tone: 'danger' },
}

const docStatusMap: StatusMap = {
  DRAFT: { label: 'Draf', tone: 'neutral' },
  POSTED: { label: 'Diposting', tone: 'success' },
  VOID: { label: 'Dibatalkan', tone: 'danger' },
}

const journalStatusMap: StatusMap = {
  POSTED: { label: 'Diposting', tone: 'success' },
  REVERSED: { label: 'Dibalik', tone: 'danger' },
}

const billStatusMap: StatusMap = {
  DRAFT: { label: 'Draf', tone: 'neutral' },
  POSTED: { label: 'Diposting', tone: 'info' },
  PARTIALLY_PAID: { label: 'Dibayar Sebagian', tone: 'warning' },
  PAID: { label: 'Lunas', tone: 'success' },
  VOID: { label: 'Dibatalkan', tone: 'danger' },
}

const periodStatusMap: StatusMap = {
  OPEN: { label: 'Terbuka', tone: 'success' },
  CLOSED: { label: 'Ditutup', tone: 'danger' },
}

const poStatusMap: StatusMap = {
  DRAFT: { label: 'Draf', tone: 'neutral' },
  APPROVED: { label: 'Disetujui', tone: 'info' },
  PARTIALLY_RECEIVED: { label: 'Diterima Sebagian', tone: 'warning' },
  FULLY_RECEIVED: { label: 'Diterima Penuh', tone: 'success' },
  CLOSED: { label: 'Ditutup', tone: 'success' },
  CANCELLED: { label: 'Dibatalkan', tone: 'danger' },
}

const poBillingStatusMap: StatusMap = {
  UNBILLED: { label: 'Belum Ditagih', tone: 'neutral' },
  PARTIALLY_BILLED: { label: 'Ditagih Sebagian', tone: 'warning' },
  FULLY_BILLED: { label: 'Ditagih Penuh', tone: 'success' },
}

const transferStatusMap: StatusMap = {
  DRAFT: { label: 'Draf', tone: 'neutral' },
  DISPATCHED: { label: 'Dikirim', tone: 'warning' },
  RECEIVED: { label: 'Diterima', tone: 'success' },
  CANCELLED: { label: 'Dibatalkan', tone: 'danger' },
}

const opnameStatusMap: StatusMap = {
  DRAFT: { label: 'Draf', tone: 'neutral' },
  POSTED: { label: 'Diposting', tone: 'success' },
  CANCELLED: { label: 'Dibatalkan', tone: 'danger' },
}

const kinds = {
  order: orderStatusMap,
  payment: paymentStatusMap,
  production: productionStatusMap,
  delivery: deliveryStatusMap,
  invoice: invoiceStatusMap,
  doc: docStatusMap,
  journal: journalStatusMap,
  bill: billStatusMap,
  period: periodStatusMap,
  po: poStatusMap,
  poBilling: poBillingStatusMap,
  transfer: transferStatusMap,
  opname: opnameStatusMap,
}

export function StatusBadge({ kind, value }: { kind: keyof typeof kinds; value: string }) {
  const entry = kinds[kind][value] ?? { label: value, tone: 'neutral' as BadgeTone }
  return <Badge tone={entry.tone}>{entry.label}</Badge>
}
