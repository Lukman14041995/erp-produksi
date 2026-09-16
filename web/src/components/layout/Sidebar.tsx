import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { ChevronDown, ChevronLeft, ChevronRight, Shirt, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCurrentUser } from '@/hooks/useAuth'
import { isGroupVisible, navGroups } from './nav'

const CLOSED_GROUPS_KEY = 'sidebar_closed_groups'

function loadClosedGroups(): Set<string> {
  try {
    const raw = localStorage.getItem(CLOSED_GROUPS_KEY)
    return raw ? new Set(JSON.parse(raw)) : new Set()
  } catch {
    return new Set()
  }
}

interface SidebarProps {
  mobileOpen: boolean
  onCloseMobile: () => void
}

export function Sidebar({ mobileOpen, onCloseMobile }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false)
  const [closedGroups, setClosedGroups] = useState<Set<string>>(loadClosedGroups)
  const user = useCurrentUser()
  const visibleGroups = navGroups.filter((g) => (user ? isGroupVisible(g, user.role) : false))

  function toggleGroup(label: string) {
    setClosedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(label)) next.delete(label)
      else next.add(label)
      try {
        localStorage.setItem(CLOSED_GROUPS_KEY, JSON.stringify([...next]))
      } catch {
        // ignore (private browsing / storage disabled) -- state still works for this session
      }
      return next
    })
  }

  return (
    <>
      {mobileOpen && <div className="fixed inset-0 z-30 bg-black/60 md:hidden" onClick={onCloseMobile} />}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-40 flex h-screen w-64 flex-col border-r border-slate-200 bg-[var(--color-surface)] transition-transform duration-200',
          'md:relative md:translate-x-0',
          mobileOpen ? 'translate-x-0' : '-translate-x-full',
          collapsed && 'md:w-16',
        )}
      >
        <div className="flex h-14 shrink-0 items-center gap-2 border-b border-slate-100 px-4">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-[var(--color-primary)] text-white">
            <Shirt className="h-4.5 w-4.5" />
          </div>
          {!collapsed && <span className="truncate text-sm font-semibold text-slate-900">Clothing ERP</span>}
          <button onClick={onCloseMobile} className="ml-auto rounded-md p-1 text-slate-400 hover:bg-slate-100 md:hidden">
            <X className="h-4 w-4" />
          </button>
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto px-2 py-4">
          {visibleGroups.map((group) => {
            const isClosed = closedGroups.has(group.label)
            return (
              <div key={group.label}>
                {group.to ? (
                  <NavLink
                    to={group.to}
                    end
                    onClick={onCloseMobile}
                    className={({ isActive }) =>
                      cn(
                        'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                        isActive ? 'bg-red-50 text-[var(--color-primary)]' : 'text-slate-600 hover:bg-slate-50',
                      )
                    }
                  >
                    <group.icon className="h-4.5 w-4.5 shrink-0" />
                    {!collapsed && group.label}
                  </NavLink>
                ) : (
                  <>
                    {!collapsed && (
                      <button
                        onClick={() => toggleGroup(group.label)}
                        className="mb-1 flex w-full items-center gap-2 rounded-md px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-400 hover:text-slate-600"
                      >
                        <group.icon className="h-3.5 w-3.5" />
                        <span className="flex-1 text-left">{group.label}</span>
                        <ChevronDown className={cn('h-3.5 w-3.5 shrink-0 transition-transform', isClosed && '-rotate-90')} />
                      </button>
                    )}
                    {(collapsed || !isClosed) && (
                      <div className="space-y-0.5">
                        {group.children?.map((child) => (
                          <NavLink
                            key={child.to}
                            to={child.to}
                            onClick={onCloseMobile}
                            title={collapsed ? child.label : undefined}
                            className={({ isActive }) =>
                              cn(
                                'block truncate rounded-md py-1.5 text-sm transition-colors',
                                collapsed ? 'px-3 text-center' : 'pl-8 pr-3',
                                isActive ? 'bg-red-50 font-medium text-[var(--color-primary)]' : 'text-slate-600 hover:bg-slate-50',
                              )
                            }
                          >
                            {collapsed ? child.label.slice(0, 1) : child.label}
                          </NavLink>
                        ))}
                      </div>
                    )}
                  </>
                )}
              </div>
            )
          })}
        </nav>

        <button
          onClick={() => setCollapsed((c) => !c)}
          className="hidden h-10 shrink-0 items-center justify-center border-t border-slate-100 text-slate-400 hover:bg-slate-50 hover:text-slate-600 md:flex"
        >
          {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </button>
      </aside>
    </>
  )
}
