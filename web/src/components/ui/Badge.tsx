import { cn } from '@/lib/utils'

export type BadgeTone = 'success' | 'warning' | 'danger' | 'info' | 'neutral'

const toneStyles: Record<BadgeTone, string> = {
  success: 'bg-[var(--color-success-surface)] text-[var(--color-success)]',
  warning: 'bg-[var(--color-warning-surface)] text-[var(--color-warning)]',
  danger: 'bg-[var(--color-danger-surface)] text-[var(--color-danger)]',
  info: 'bg-[var(--color-info-surface)] text-[var(--color-info)]',
  neutral: 'bg-[var(--color-neutral-surface)] text-[var(--color-neutral)]',
}

export function Badge({ tone = 'neutral', className, children }: { tone?: BadgeTone; className?: string; children: React.ReactNode }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium whitespace-nowrap',
        toneStyles[tone],
        className,
      )}
    >
      {children}
    </span>
  )
}
