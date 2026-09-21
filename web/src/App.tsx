import { QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { queryClient } from '@/lib/queryClient'
import { Toaster } from '@/components/ui/Toaster'
import { ProtectedRoute, RoleGate } from '@/components/ProtectedRoute'
import { AppShell } from '@/components/layout/AppShell'
import { LoginPage } from '@/pages/LoginPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { CustomersPage } from '@/pages/master-data/CustomersPage'
import { SuppliersPage } from '@/pages/master-data/SuppliersPage'
import { ProductsPage } from '@/pages/master-data/ProductsPage'
import { MaterialsPage } from '@/pages/master-data/MaterialsPage'
import { ChartOfAccountsPage } from '@/pages/master-data/ChartOfAccountsPage'
import { ProductTypesPage } from '@/pages/master-data/ProductTypesPage'
import { FabricsPage } from '@/pages/master-data/FabricsPage'
import { GarmentVariantsPage } from '@/pages/master-data/GarmentVariantsPage'
import { GarmentSizesPage } from '@/pages/master-data/GarmentSizesPage'
import { InksPage } from '@/pages/master-data/InksPage'
import { SalesOrdersPage } from '@/pages/sales/SalesOrdersPage'
import { SalesOrderNewPage } from '@/pages/sales/SalesOrderNewPage'
import { SalesOrderDetailPage } from '@/pages/sales/SalesOrderDetailPage'
import { InvoicesPage } from '@/pages/sales/InvoicesPage'
import { OrderLinksPage } from '@/pages/sales/OrderLinksPage'
import { QuotationsPage } from '@/pages/sales/QuotationsPage'
import { QuotationDetailPage } from '@/pages/sales/QuotationDetailPage'
import { OrderConfiguratorPage } from '@/pages/public/OrderConfiguratorPage'
import { ProductionOrdersPage } from '@/pages/production/ProductionOrdersPage'
import { ProductionOrderNewPage } from '@/pages/production/ProductionOrderNewPage'
import { ProductionOrderDetailPage } from '@/pages/production/ProductionOrderDetailPage'
import { SpkListPage } from '@/pages/production/SpkListPage'
import { SpkDetailPage } from '@/pages/production/SpkDetailPage'
import { SpkKanbanPage } from '@/pages/production/SpkKanbanPage'
import { InventoryBalancesPage } from '@/pages/inventory/InventoryBalancesPage'
import { InventoryTransactionsPage } from '@/pages/inventory/InventoryTransactionsPage'
import { WarehousesPage } from '@/pages/inventory/WarehousesPage'
import { StockTransfersPage } from '@/pages/inventory/StockTransfersPage'
import { StockTransferNewPage } from '@/pages/inventory/StockTransferNewPage'
import { StockTransferDetailPage } from '@/pages/inventory/StockTransferDetailPage'
import { StockOpnamesPage } from '@/pages/inventory/StockOpnamesPage'
import { StockOpnameNewPage } from '@/pages/inventory/StockOpnameNewPage'
import { StockOpnameDetailPage } from '@/pages/inventory/StockOpnameDetailPage'
import { SupplierBillsPage } from '@/pages/purchasing/SupplierBillsPage'
import { SupplierBillDetailPage } from '@/pages/purchasing/SupplierBillDetailPage'
import { PurchaseOrdersPage } from '@/pages/purchasing/PurchaseOrdersPage'
import { PurchaseOrderNewPage } from '@/pages/purchasing/PurchaseOrderNewPage'
import { PurchaseOrderDetailPage } from '@/pages/purchasing/PurchaseOrderDetailPage'
import { PaymentsPage } from '@/pages/finance/PaymentsPage'
import { BankAccountsPage } from '@/pages/finance/BankAccountsPage'
import { ExpensesPage } from '@/pages/finance/ExpensesPage'
import { JournalsPage } from '@/pages/accounting/JournalsPage'
import { JournalDetailPage } from '@/pages/accounting/JournalDetailPage'
import { PeriodsPage } from '@/pages/accounting/PeriodsPage'
import { ProfitLossPage } from '@/pages/reports/ProfitLossPage'
import { BalanceSheetPage } from '@/pages/reports/BalanceSheetPage'
import { CashFlowPage } from '@/pages/reports/CashFlowPage'
import { ARAgingPage } from '@/pages/reports/ARAgingPage'
import { APAgingPage } from '@/pages/reports/APAgingPage'
import { OrderProfitabilityPage } from '@/pages/reports/OrderProfitabilityPage'
import { UsersPage } from '@/pages/settings/UsersPage'

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/order/:token" element={<OrderConfiguratorPage view="list" />} />
          <Route path="/order/:token/new" element={<OrderConfiguratorPage view="new" />} />
          <Route path="/order/:token/orders/:quotationId" element={<OrderConfiguratorPage view="detail" />} />

          <Route element={<ProtectedRoute />}>
            <Route element={<AppShell />}>
              <Route index element={<DashboardPage />} />

              <Route path="master-data/customers" element={<CustomersPage />} />
              <Route path="master-data/suppliers" element={<SuppliersPage />} />
              <Route path="master-data/products" element={<ProductsPage />} />
              <Route path="master-data/materials" element={<MaterialsPage />} />
              <Route path="master-data/accounts" element={<ChartOfAccountsPage />} />
              <Route path="master-data/product-types" element={<ProductTypesPage />} />
              <Route path="master-data/fabrics" element={<FabricsPage />} />
              <Route path="master-data/garment-variants" element={<GarmentVariantsPage />} />
              <Route path="master-data/garment-sizes" element={<GarmentSizesPage />} />
              <Route path="master-data/inks" element={<InksPage />} />

              <Route
                path="sales/orders"
                element={
                  <RoleGate allow={['ADMIN', 'SALES']}>
                    <SalesOrdersPage />
                  </RoleGate>
                }
              />
              <Route
                path="sales/orders/new"
                element={
                  <RoleGate allow={['ADMIN', 'SALES']}>
                    <SalesOrderNewPage />
                  </RoleGate>
                }
              />
              <Route
                path="sales/orders/:id"
                element={
                  <RoleGate allow={['ADMIN', 'SALES']}>
                    <SalesOrderDetailPage />
                  </RoleGate>
                }
              />
              <Route
                path="sales/invoices"
                element={
                  <RoleGate allow={['ADMIN', 'SALES']}>
                    <InvoicesPage />
                  </RoleGate>
                }
              />
              <Route
                path="sales/order-links"
                element={
                  <RoleGate allow={['ADMIN', 'SALES']}>
                    <OrderLinksPage />
                  </RoleGate>
                }
              />
              <Route
                path="sales/quotations"
                element={
                  <RoleGate allow={['ADMIN', 'SALES']}>
                    <QuotationsPage />
                  </RoleGate>
                }
              />
              <Route
                path="sales/quotations/:id"
                element={
                  <RoleGate allow={['ADMIN', 'SALES']}>
                    <QuotationDetailPage />
                  </RoleGate>
                }
              />

              <Route
                path="production/costing"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION']}>
                    <SpkKanbanPage />
                  </RoleGate>
                }
              />
              <Route
                path="production/orders/tshirt"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION']}>
                    <SpkListPage productTypeCode="TSHIRT" title="Pesanan Produksi T-Shirt" />
                  </RoleGate>
                }
              />
              <Route
                path="production/orders/jersey"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION']}>
                    <SpkListPage productTypeCode="JERSEY" title="Pesanan Produksi Jersey" />
                  </RoleGate>
                }
              />
              <Route
                path="production/orders/:id"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION']}>
                    <SpkDetailPage />
                  </RoleGate>
                }
              />
              <Route
                path="production/hpp"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION']}>
                    <ProductionOrdersPage />
                  </RoleGate>
                }
              />
              <Route
                path="production/hpp/new"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION']}>
                    <ProductionOrderNewPage />
                  </RoleGate>
                }
              />
              <Route
                path="production/hpp/:id"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION']}>
                    <ProductionOrderDetailPage />
                  </RoleGate>
                }
              />

              <Route
                path="inventory/balances"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <InventoryBalancesPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/transactions"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <InventoryTransactionsPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/warehouses"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <WarehousesPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/transfers"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <StockTransfersPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/transfers/new"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <StockTransferNewPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/transfers/:id"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <StockTransferDetailPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/opnames"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <StockOpnamesPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/opnames/new"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <StockOpnameNewPage />
                  </RoleGate>
                }
              />
              <Route
                path="inventory/opnames/:id"
                element={
                  <RoleGate allow={['ADMIN', 'PRODUCTION', 'FINANCE']}>
                    <StockOpnameDetailPage />
                  </RoleGate>
                }
              />

              <Route
                path="purchasing/orders"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE', 'PRODUCTION']}>
                    <PurchaseOrdersPage />
                  </RoleGate>
                }
              />
              <Route
                path="purchasing/orders/new"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE', 'PRODUCTION']}>
                    <PurchaseOrderNewPage />
                  </RoleGate>
                }
              />
              <Route
                path="purchasing/orders/:id"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE', 'PRODUCTION']}>
                    <PurchaseOrderDetailPage />
                  </RoleGate>
                }
              />
              <Route
                path="purchasing/bills"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE', 'PRODUCTION']}>
                    <SupplierBillsPage />
                  </RoleGate>
                }
              />
              <Route
                path="purchasing/bills/:id"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE', 'PRODUCTION']}>
                    <SupplierBillDetailPage />
                  </RoleGate>
                }
              />

              <Route
                path="finance/payments"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE']}>
                    <PaymentsPage />
                  </RoleGate>
                }
              />
              <Route
                path="finance/expenses"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE']}>
                    <ExpensesPage />
                  </RoleGate>
                }
              />
              <Route
                path="finance/bank-accounts"
                element={
                  <RoleGate allow={['ADMIN', 'FINANCE']}>
                    <BankAccountsPage />
                  </RoleGate>
                }
              />

              <Route
                path="accounting/journals"
                element={
                  <RoleGate allow={['ADMIN', 'ACCOUNTING']}>
                    <JournalsPage />
                  </RoleGate>
                }
              />
              <Route
                path="accounting/journals/:id"
                element={
                  <RoleGate allow={['ADMIN', 'ACCOUNTING']}>
                    <JournalDetailPage />
                  </RoleGate>
                }
              />
              <Route
                path="accounting/periods"
                element={
                  <RoleGate allow={['ADMIN', 'ACCOUNTING']}>
                    <PeriodsPage />
                  </RoleGate>
                }
              />

              <Route path="reports/profit-loss" element={<ProfitLossPage />} />
              <Route path="reports/balance-sheet" element={<BalanceSheetPage />} />
              <Route path="reports/cash-flow" element={<CashFlowPage />} />
              <Route path="reports/ar-aging" element={<ARAgingPage />} />
              <Route path="reports/ap-aging" element={<APAgingPage />} />
              <Route path="reports/order-profitability" element={<OrderProfitabilityPage />} />

              <Route
                path="settings/users"
                element={
                  <RoleGate allow={['ADMIN']}>
                    <UsersPage />
                  </RoleGate>
                }
              />

              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Route>
        </Routes>
      </BrowserRouter>
      <Toaster />
    </QueryClientProvider>
  )
}
