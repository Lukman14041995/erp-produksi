import { Loader2 } from 'lucide-react'

export function LoadingState({ label = 'Memuat...' }: { label?: string }) {
  return (
    <div className="flex items-center justify-center gap-2 py-16 text-sm text-slate-400">
      <Loader2 className="h-4 w-4 animate-spin" /> {label}
    </div>
  )
}

export function ErrorState({ message }: { message: string }) {
  return (
    <div className="rounded-md bg-[var(--color-danger-surface)] px-4 py-3 text-sm text-[var(--color-danger)]">
      {message}
    </div>
  )
}
