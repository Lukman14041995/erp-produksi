import { useNavigate, useParams } from 'react-router-dom'
import { Shirt } from 'lucide-react'
import { LoadingState } from '@/components/QueryState'
import { usePublicCatalog } from '@/hooks/useQuotations'
import { NewOrderWizard } from '@/pages/public/NewOrderWizard'
import { OrderDetailView } from '@/pages/public/OrderDetailView'
import { OrderListView } from '@/pages/public/OrderListView'

// The view (list/new/detail) lives in the URL, not component state, so a
// reload or a shared/bookmarked link lands the customer back where they
// were -- e.g. checking on one order's payment status days later -- instead
// of always resetting to the order list.
export function OrderConfiguratorPage({ view }: { view: 'list' | 'new' | 'detail' }) {
  const { token, quotationId } = useParams<{ token: string; quotationId: string }>()
  const navigate = useNavigate()
  const { data: catalog, isLoading, error } = usePublicCatalog(token)

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <LoadingState label="Memuat..." />
      </div>
    )
  }

  if (error || !catalog || !token) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-2 bg-slate-50 px-4 text-center">
        <Shirt className="h-8 w-8 text-slate-300" />
        <p className="text-lg font-semibold text-slate-900">Link tidak valid</p>
        <p className="text-sm text-slate-500">Link pesanan ini sudah kedaluwarsa, dicabut, atau salah. Silakan minta link baru.</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-50 px-4 py-8">
      <div className="mx-auto max-w-4xl">
        <div className="mb-6 flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-[var(--color-primary)] text-white">
            <Shirt className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-lg font-semibold text-slate-900">Pesan Jersey &amp; T-Shirt Custom</h1>
            <p className="text-sm text-slate-500">Halo {catalog.customer_name}.</p>
          </div>
        </div>

        {view === 'list' && (
          <OrderListView
            token={token}
            onSelect={(id) => navigate(`/order/${token}/orders/${id}`)}
            onNewOrder={() => navigate(`/order/${token}/new`)}
          />
        )}
        {view === 'new' && (
          <NewOrderWizard
            token={token}
            catalog={catalog}
            onBack={() => navigate(`/order/${token}`)}
            onSubmitted={() => navigate(`/order/${token}`)}
          />
        )}
        {view === 'detail' && quotationId && (
          <OrderDetailView token={token} quotationId={quotationId} onBack={() => navigate(`/order/${token}`)} />
        )}
      </div>
    </div>
  )
}
