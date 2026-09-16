import { useQuery } from '@tanstack/react-query'
import { apiGet } from '@/lib/api'
import type {
  APAgingReport,
  ARAgingReport,
  BalanceSheet,
  CashFlow,
  OrderProfitability,
  ProfitLoss,
  TrialBalance,
} from '@/types/reporting'

export function useTrialBalance(asOf: string) {
  return useQuery({
    queryKey: ['reports', 'trial-balance', asOf],
    queryFn: () => apiGet<TrialBalance>('/reports/trial-balance', { as_of: asOf }),
  })
}

export function useProfitLoss(from: string, to: string) {
  return useQuery({
    queryKey: ['reports', 'profit-loss', from, to],
    queryFn: () => apiGet<ProfitLoss>('/reports/profit-loss', { from, to }),
  })
}

export function useBalanceSheet(asOf: string) {
  return useQuery({
    queryKey: ['reports', 'balance-sheet', asOf],
    queryFn: () => apiGet<BalanceSheet>('/reports/balance-sheet', { as_of: asOf }),
  })
}

export function useCashFlow(from: string, to: string) {
  return useQuery({
    queryKey: ['reports', 'cash-flow', from, to],
    queryFn: () => apiGet<CashFlow>('/reports/cash-flow', { from, to }),
  })
}

export function useARAging(asOf: string) {
  return useQuery({
    queryKey: ['reports', 'ar-aging', asOf],
    queryFn: () => apiGet<ARAgingReport>('/reports/ar-aging', { as_of: asOf }),
  })
}

export function useAPAging(asOf: string) {
  return useQuery({
    queryKey: ['reports', 'ap-aging', asOf],
    queryFn: () => apiGet<APAgingReport>('/reports/ap-aging', { as_of: asOf }),
  })
}

export function useOrderProfitability() {
  return useQuery({
    queryKey: ['reports', 'order-profitability'],
    queryFn: () => apiGet<OrderProfitability[]>('/reports/order-profitability'),
  })
}
