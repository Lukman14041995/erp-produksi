import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useCurrentUser } from '@/hooks/useAuth'
import type { Role } from '@/types/auth'

export function ProtectedRoute() {
  const user = useCurrentUser()
  const location = useLocation()

  if (!user) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }
  return <Outlet />
}

export function RoleGate({ allow, children }: { allow: Role[]; children: React.ReactNode }) {
  const user = useCurrentUser()
  if (!user || !allow.includes(user.role)) {
    return (
      <div className="flex h-full flex-col items-center justify-center py-24 text-center">
        <p className="text-lg font-semibold text-slate-900">Akses dibatasi</p>
        <p className="mt-1 text-sm text-slate-500">Peran Anda ({user?.role}) tidak memiliki akses ke halaman ini.</p>
      </div>
    )
  }
  return <>{children}</>
}
