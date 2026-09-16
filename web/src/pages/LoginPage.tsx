import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { Shirt } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { useCurrentUser, useLogin } from '@/hooks/useAuth'
import { ApiError } from '@/types/api'

const schema = z.object({
  email: z.string().email('Masukkan email yang valid'),
  password: z.string().min(1, 'Kata sandi wajib diisi'),
})
type FormValues = z.infer<typeof schema>

export function LoginPage() {
  const user = useCurrentUser()
  const navigate = useNavigate()
  const location = useLocation()
  const login = useLogin()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  if (user) {
    const from = (location.state as { from?: Location })?.from?.pathname ?? '/'
    return <Navigate to={from} replace />
  }

  function onSubmit(values: FormValues) {
    login.mutate(values, {
      onSuccess: () => navigate('/'),
    })
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
      <div className="w-full max-w-sm rounded-xl border border-slate-200 bg-[var(--color-surface)] p-8 shadow-sm">
        <div className="mb-6 flex flex-col items-center">
          <div className="mb-3 flex h-11 w-11 items-center justify-center rounded-lg bg-[var(--color-primary)] text-white">
            <Shirt className="h-6 w-6" />
          </div>
          <h1 className="text-lg font-semibold text-slate-900">Clothing ERP</h1>
          <p className="mt-1 text-sm text-slate-500">Masuk ke akun Anda</p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" autoComplete="email" {...register('email')} />
            {errors.email && <p className="text-xs text-[var(--color-danger)]">{errors.email.message}</p>}
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="password">Kata Sandi</Label>
            <Input id="password" type="password" autoComplete="current-password" {...register('password')} />
            {errors.password && <p className="text-xs text-[var(--color-danger)]">{errors.password.message}</p>}
          </div>

          {login.isError && (
            <p className="rounded-md bg-[var(--color-danger-surface)] px-3 py-2 text-xs text-[var(--color-danger)]">
              {login.error instanceof ApiError ? login.error.message : 'Login gagal'}
            </p>
          )}

          <Button type="submit" className="w-full" loading={login.isPending}>
            Masuk
          </Button>
        </form>
      </div>
    </div>
  )
}
