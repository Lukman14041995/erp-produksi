import { LogOut, User as UserIcon } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useCurrentUser, useLogout } from '@/hooks/useAuth'
import { Badge } from '@/components/ui/Badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/DropdownMenu'
import { CommandMenu } from './CommandMenu'

const roleTone: Record<string, 'info' | 'success' | 'warning' | 'neutral'> = {
  ADMIN: 'info',
  SALES: 'success',
  PRODUCTION: 'warning',
  FINANCE: 'neutral',
  ACCOUNTING: 'neutral',
}

export function Header() {
  const user = useCurrentUser()
  const logout = useLogout()
  const navigate = useNavigate()

  return (
    <header className="flex h-14 items-center justify-between border-b border-slate-200 bg-white px-6">
      <CommandMenu />

      <div className="flex items-center gap-4">
        {user && <Badge tone={roleTone[user.role] ?? 'neutral'}>{user.role}</Badge>}

        <DropdownMenu>
          <DropdownMenuTrigger className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-slate-50">
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-slate-200 text-slate-600">
              <UserIcon className="h-4 w-4" />
            </div>
            <span className="text-slate-700">{user?.name}</span>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>{user?.email}</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onSelect={() => {
                logout()
                navigate('/login')
              }}
            >
              <LogOut className="h-4 w-4" /> Keluar
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
