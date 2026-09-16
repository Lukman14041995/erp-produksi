import { useEffect, useState } from 'react'
import { Command } from 'cmdk'
import { useNavigate } from 'react-router-dom'
import { Search } from 'lucide-react'
import { useCurrentUser } from '@/hooks/useAuth'
import { isGroupVisible, navGroups } from './nav'

export function CommandMenu() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const user = useCurrentUser()

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((o) => !o)
      }
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [])

  if (!user) return null
  const visibleGroups = navGroups.filter((g) => isGroupVisible(g, user.role))

  function go(to: string) {
    setOpen(false)
    navigate(to)
  }

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="flex h-9 w-9 items-center justify-center gap-2 rounded-md border border-slate-200 bg-slate-50 text-sm text-slate-400 hover:bg-slate-100 md:w-64 md:justify-start md:px-3 md:py-1.5"
      >
        <Search className="h-4 w-4 shrink-0" />
        <span className="hidden md:inline">Pencarian cepat</span>
        <kbd className="ml-auto hidden rounded border border-slate-300 bg-[var(--color-surface)] px-1.5 py-0.5 text-[10px] font-medium text-slate-500 md:inline-flex">
          &#8984;K
        </kbd>
      </button>

      {open && (
        <div className="fixed inset-0 z-100 flex items-start justify-center bg-black/40 pt-24" onClick={() => setOpen(false)}>
          <div onClick={(e) => e.stopPropagation()} className="w-full max-w-lg overflow-hidden rounded-lg bg-[var(--color-surface)] shadow-xl">
            <Command label="Pencarian cepat">
              <Command.Input
                autoFocus
                placeholder="Lompat ke halaman..."
                className="w-full border-b border-slate-100 px-4 py-3 text-sm outline-none"
              />
              <Command.List className="max-h-80 overflow-y-auto p-2">
                <Command.Empty className="px-3 py-6 text-center text-sm text-slate-400">Tidak ada hasil ditemukan.</Command.Empty>
                {visibleGroups.map((group) => (
                  <Command.Group key={group.label} heading={group.label} className="mb-1 [&_[cmdk-group-heading]]:px-3 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-semibold [&_[cmdk-group-heading]]:text-slate-400">
                    {group.to && (
                      <Command.Item
                        onSelect={() => go(group.to!)}
                        className="cursor-pointer rounded-md px-3 py-2 text-sm text-slate-700 data-[selected=true]:bg-slate-100"
                      >
                        {group.label}
                      </Command.Item>
                    )}
                    {group.children?.map((child) => (
                      <Command.Item
                        key={child.to}
                        onSelect={() => go(child.to)}
                        className="cursor-pointer rounded-md px-3 py-2 text-sm text-slate-700 data-[selected=true]:bg-slate-100"
                      >
                        {child.label}
                      </Command.Item>
                    ))}
                  </Command.Group>
                ))}
              </Command.List>
            </Command>
          </div>
        </div>
      )}
    </>
  )
}
