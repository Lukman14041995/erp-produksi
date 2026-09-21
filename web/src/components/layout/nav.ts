import {
  BarChart3,
  BookOpen,
  Boxes,
  Factory,
  LayoutDashboard,
  Database,
  Settings,
  ShoppingCart,
  Truck,
  Wallet,
} from 'lucide-react'
import type { Role } from '@/types/auth'

export interface NavChild {
  label: string
  to: string
}

export interface NavGroup {
  label: string
  icon: typeof LayoutDashboard
  roles: Role[] | 'all'
  to?: string
  children?: NavChild[]
}

export const navGroups: NavGroup[] = [
  { label: 'Dasbor', icon: LayoutDashboard, roles: 'all', to: '/' },
  {
    label: 'Data Master',
    icon: Database,
    roles: 'all',
    children: [
      { label: 'Pelanggan', to: '/master-data/customers' },
      { label: 'Pemasok', to: '/master-data/suppliers' },
      { label: 'Produk', to: '/master-data/products' },
      { label: 'Bahan Baku', to: '/master-data/materials' },
      { label: 'Bagan Akun', to: '/master-data/accounts' },
      { label: 'Jenis Pesanan', to: '/master-data/product-types' },
      { label: 'Model Potongan', to: '/master-data/garment-variants' },
      { label: 'Bahan Jual', to: '/master-data/fabrics' },
      { label: 'Tinta', to: '/master-data/inks' },
      { label: 'Ukuran Pesanan', to: '/master-data/garment-sizes' },
    ],
  },
  {
    label: 'Penjualan',
    icon: ShoppingCart,
    roles: ['ADMIN', 'SALES'],
    children: [
      { label: 'Pesanan Penjualan', to: '/sales/orders' },
      { label: 'Faktur', to: '/sales/invoices' },
      { label: 'Link Pesanan Customer', to: '/sales/order-links' },
      { label: 'Quotation Masuk', to: '/sales/quotations' },
    ],
  },
  {
    label: 'Produksi',
    icon: Factory,
    roles: ['ADMIN', 'PRODUCTION'],
    children: [
      { label: 'Papan Produksi', to: '/production/costing' },
      { label: 'Pesanan Produksi T-Shirt', to: '/production/orders/tshirt' },
      { label: 'Pesanan Produksi Jersey', to: '/production/orders/jersey' },
      { label: 'HPP Produksi (BOM)', to: '/production/hpp' },
    ],
  },
  {
    label: 'Persediaan',
    icon: Boxes,
    roles: ['ADMIN', 'PRODUCTION', 'FINANCE'],
    children: [
      { label: 'Saldo Persediaan', to: '/inventory/balances' },
      { label: 'Transaksi', to: '/inventory/transactions' },
      { label: 'Gudang', to: '/inventory/warehouses' },
      { label: 'Transfer Stok', to: '/inventory/transfers' },
      { label: 'Stock Opname', to: '/inventory/opnames' },
    ],
  },
  {
    label: 'Pembelian',
    icon: Truck,
    roles: ['ADMIN', 'FINANCE', 'PRODUCTION'],
    children: [
      { label: 'Pesanan Pembelian', to: '/purchasing/orders' },
      { label: 'Tagihan Pemasok', to: '/purchasing/bills' },
    ],
  },
  {
    label: 'Keuangan',
    icon: Wallet,
    roles: ['ADMIN', 'FINANCE'],
    children: [
      { label: 'Pembayaran', to: '/finance/payments' },
      { label: 'Pengeluaran', to: '/finance/expenses' },
      { label: 'Rekening Bank', to: '/finance/bank-accounts' },
    ],
  },
  {
    label: 'Akuntansi',
    icon: BookOpen,
    roles: ['ADMIN', 'ACCOUNTING'],
    children: [
      { label: 'Jurnal', to: '/accounting/journals' },
      { label: 'Periode', to: '/accounting/periods' },
    ],
  },
  {
    label: 'Laporan',
    icon: BarChart3,
    roles: 'all',
    children: [
      { label: 'Laba Rugi', to: '/reports/profit-loss' },
      { label: 'Neraca', to: '/reports/balance-sheet' },
      { label: 'Arus Kas', to: '/reports/cash-flow' },
      { label: 'Umur Piutang', to: '/reports/ar-aging' },
      { label: 'Umur Utang', to: '/reports/ap-aging' },
      { label: 'Profitabilitas Pesanan', to: '/reports/order-profitability' },
    ],
  },
  {
    label: 'Pengaturan',
    icon: Settings,
    roles: ['ADMIN'],
    children: [{ label: 'Pengguna', to: '/settings/users' }],
  },
]

export function isGroupVisible(group: NavGroup, role: Role): boolean {
  return group.roles === 'all' || group.roles.includes(role)
}
