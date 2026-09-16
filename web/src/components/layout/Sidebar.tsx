import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Shirt } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCurrentUser } from '@/hooks/useAuth'
import { isGroupVisible, navGroups } from './nav'

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const user = useCurrentUser()
  const visibleGroups = navGroups.filter((g) => (user ? isGroupVisible(g, user.role) : false))

  return (
    <aside
      className={cn(
        'flex h-screen flex-col border-r border-slate-200 bg-white transition-[width] duration-200',
        collapsed ? 'w-16' : 'w-64',
      )}
    >
      <div className="flex h-14 items-center gap-2 border-b border-slate-100 px-4">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-[var(--color-primary)] text-white">
          <Shirt className="h-4.5 w-4.5" />
        </div>
        {!collapsed && <span className="truncate text-sm font-semibold text-slate-900">Clothing ERP</span>}
      </div>

      <nav className="flex-1 space-y-4 overflow-y-auto px-2 py-4">
        {visibleGroups.map((group) => (
          <div key={group.label}>
            {group.to ? (
              <NavLink
                to={group.to}
                end
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
                  <div className="mb-1 flex items-center gap-2 px-3 text-xs font-semibold uppercase tracking-wide text-slate-400">
                    <group.icon className="h-3.5 w-3.5" />
                    {group.label}
                  </div>
                )}
                <div className="space-y-0.5">
                  {group.children?.map((child) => (
                    <NavLink
                      key={child.to}
                      to={child.to}
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
              </>
            )}
          </div>
        ))}
      </nav>

      <button
        onClick={() => setCollapsed((c) => !c)}
        className="flex h-10 items-center justify-center border-t border-slate-100 text-slate-400 hover:bg-slate-50 hover:text-slate-600"
      >
        {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
      </button>
    </aside>
  )
}
