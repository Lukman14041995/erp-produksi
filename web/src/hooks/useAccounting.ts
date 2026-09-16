import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type { AccountingPeriod, JournalEntry } from '@/types/accounting'

export function useJournals() {
  return useQuery({
    queryKey: ['journals'],
    queryFn: () => apiGet<JournalEntry[]>('/journals', { limit: 200 }),
  })
}

export function useJournal(id: string | undefined) {
  return useQuery({
    queryKey: ['journals', id],
    queryFn: () => apiGet<JournalEntry>(`/journals/${id}`),
    enabled: !!id,
  })
}

export function useAuditTrail(sourceIds: string[]) {
  const key = sourceIds.filter(Boolean).sort().join(',')
  return useQuery({
    queryKey: ['journals', 'audit-trail', key],
    queryFn: () => apiGet<JournalEntry[]>('/journals/audit-trail', { source_ids: key }),
    enabled: sourceIds.filter(Boolean).length > 0,
  })
}

export function useReverseJournal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => apiPost<JournalEntry>(`/journals/${id}/reverse`, { reason }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['journals'] }),
  })
}

export interface PostManualJournalInput {
  journal_date?: string
  description: string
  lines: { account_code: string; debit: string; credit: string; description?: string }[]
}

export function usePostManualJournal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: PostManualJournalInput) => apiPost<JournalEntry>('/journals', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['journals'] }),
  })
}

export function useAccountingPeriods() {
  return useQuery({
    queryKey: ['accounting-periods'],
    queryFn: () => apiGet<AccountingPeriod[]>('/accounting-periods'),
  })
}

export function useCreatePeriod() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { period: string; start_date: string; end_date: string }) =>
      apiPost<AccountingPeriod>('/accounting-periods', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['accounting-periods'] }),
  })
}

function usePeriodCommand(command: 'close' | 'reopen') {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (period: string) => apiPost<AccountingPeriod>(`/accounting-periods/${period}/${command}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['accounting-periods'] }),
  })
}

export const useClosePeriod = () => usePeriodCommand('close')
export const useReopenPeriod = () => usePeriodCommand('reopen')
