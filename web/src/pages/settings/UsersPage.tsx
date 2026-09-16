import { useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Badge } from '@/components/ui/Badge'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useRegisterUser, useUsers } from '@/hooks/useUsers'
import { formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Role, User } from '@/types/auth'

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8, 'Minimal 8 karakter'),
  name: z.string().min(1, 'Wajib diisi'),
  role: z.enum(['ADMIN', 'SALES', 'PRODUCTION', 'FINANCE', 'ACCOUNTING']),
})
type FormValues = z.infer<typeof schema>

const roleTone: Record<Role, 'info' | 'success' | 'warning' | 'neutral'> = {
  ADMIN: 'info',
  SALES: 'success',
  PRODUCTION: 'warning',
  FINANCE: 'neutral',
  ACCOUNTING: 'neutral',
}

export function UsersPage() {
  const { data, isLoading, error } = useUsers()
  const registerUser = useRegisterUser()
  const [open, setOpen] = useState(false)

  const { register, handleSubmit, control, reset } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { role: 'SALES' },
  })

  function onSubmit(values: FormValues) {
    registerUser
      .mutateAsync(values)
      .then(() => {
        toast.success('Pengguna berhasil dibuat')
        setOpen(false)
        reset({ role: 'SALES', email: '', password: '', name: '' })
      })
      .catch((err) => toast.error('Gagal membuat pengguna', err instanceof ApiError ? err.message : undefined))
  }

  const columns: Column<User>[] = [
    { key: 'name', header: 'Nama', render: (u) => u.name, sortValue: (u) => u.name, csvValue: (u) => u.name },
    { key: 'email', header: 'Email', render: (u) => u.email, csvValue: (u) => u.email },
    { key: 'role', header: 'Peran', render: (u) => <Badge tone={roleTone[u.role]}>{u.role}</Badge>, csvValue: (u) => u.role },
    {
      key: 'status',
      header: 'Status',
      render: (u) => <Badge tone={u.is_active ? 'success' : 'neutral'}>{u.is_active ? 'Aktif' : 'Tidak Aktif'}</Badge>,
      csvValue: (u) => (u.is_active ? 'Aktif' : 'Tidak Aktif'),
    },
    { key: 'created', header: 'Dibuat', render: (u) => formatDate(u.created_at), csvValue: (u) => u.created_at },
  ]

  return (
    <div>
      <PageHeader
        title="Pengguna"
        description="Kelola akun ERP dan akses berbasis peran (role) mereka."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Pengguna Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(u) => u.id}
          exportFilename="users"
          searchPlaceholder="Cari pengguna..."
          searchFn={(u, q) => u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q)}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Pengguna Baru</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="space-y-1.5">
              <Label>Nama</Label>
              <Input {...register('name')} />
            </div>
            <div className="space-y-1.5">
              <Label>Email</Label>
              <Input type="email" {...register('email')} />
            </div>
            <div className="space-y-1.5">
              <Label>Kata Sandi</Label>
              <Input type="password" {...register('password')} />
            </div>
            <div className="space-y-1.5">
              <Label>Peran</Label>
              <Controller
                control={control}
                name="role"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {(['ADMIN', 'SALES', 'PRODUCTION', 'FINANCE', 'ACCOUNTING'] as const).map((r) => (
                        <SelectItem key={r} value={r}>
                          {r}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={registerUser.isPending}>
                Buat Pengguna
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
