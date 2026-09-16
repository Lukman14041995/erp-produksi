import { Component, type ErrorInfo, type ReactNode } from 'react'
import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/Button'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Unhandled UI error:', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex flex-col items-center justify-center gap-3 py-24 text-center">
          <AlertTriangle className="h-8 w-8 text-[var(--color-danger)]" />
          <p className="text-lg font-semibold text-slate-900">Terjadi kesalahan saat menampilkan halaman ini</p>
          <p className="max-w-md text-sm text-slate-500">{this.state.error.message}</p>
          <Button variant="secondary" onClick={() => window.location.reload()}>
            Muat Ulang Halaman
          </Button>
        </div>
      )
    }
    return this.props.children
  }
}
