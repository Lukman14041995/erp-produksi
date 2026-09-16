import { CheckCircle2, Info, X, XCircle } from 'lucide-react'
import { useToastStore } from '@/store/toastStore'
import { cn } from '@/lib/utils'

const icons = {
  success: CheckCircle2,
  error: XCircle,
  info: Info,
}

const variantStyles = {
  success: 'border-l-4 border-l-[var(--color-success)]',
  error: 'border-l-4 border-l-[var(--color-danger)]',
  info: 'border-l-4 border-l-[var(--color-info)]',
}

const iconColor = {
  success: 'text-[var(--color-success)]',
  error: 'text-[var(--color-danger)]',
  info: 'text-[var(--color-info)]',
}

export function Toaster() {
  const { toasts, dismiss } = useToastStore()

  return (
    <div className="fixed bottom-4 right-4 z-100 flex w-full max-w-sm flex-col gap-2">
      {toasts.map((t) => {
        const Icon = icons[t.variant]
        return (
          <div
            key={t.id}
            role="alert"
            className={cn(
              'flex items-start gap-3 rounded-md bg-white p-4 shadow-lg ring-1 ring-black/5',
              variantStyles[t.variant],
            )}
          >
            <Icon className={cn('mt-0.5 h-5 w-5 shrink-0', iconColor[t.variant])} />
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-slate-900">{t.title}</p>
              {t.description && <p className="mt-0.5 text-sm text-slate-500 break-words">{t.description}</p>}
            </div>
            <button onClick={() => dismiss(t.id)} className="text-slate-400 hover:text-slate-600">
              <X className="h-4 w-4" />
            </button>
          </div>
        )
      })}
    </div>
  )
}
