import type { ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { Download, Printer } from 'lucide-react'
import { Dialog, DialogContent } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'

interface PrintPreviewModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  onDownloadPdf?: () => void
  downloadPending?: boolean
  children: ReactNode
}

// Generic print-preview shell: shows an A4-sized page preview on screen,
// with "Print" (window.print()) and "Download PDF" (server-rendered PDF)
// actions.
//
// The interactive on-screen dialog (Radix Dialog: fixed + transformed +
// height-clamped) is NOT what gets printed -- print CSS can't reliably
// strip away a `fixed`/`transform` ancestor's containing-block side effects
// (verified: it silently mispositions/clips content instead of erroring).
// Instead, a second, plain copy of the same content is portaled straight
// onto <body>, completely outside the dialog's DOM, and is the only thing
// left visible under `@media print` (see index.css: #root and the dialog
// itself are hidden for print; .print-portal-root is shown).
export function PrintPreviewModal({ open, onOpenChange, title, onDownloadPdf, downloadPending, children }: PrintPreviewModalProps) {
  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-[880px] bg-slate-100 p-0">
          <div className="flex items-center justify-between border-b border-slate-200 bg-white py-3 pl-5 pr-14">
            <p className="text-sm font-medium text-slate-700">{title}</p>
            <div className="flex gap-2">
              {onDownloadPdf && (
                <Button variant="secondary" size="sm" onClick={onDownloadPdf} loading={downloadPending}>
                  <Download className="h-3.5 w-3.5" /> Unduh PDF
                </Button>
              )}
              <Button size="sm" onClick={() => window.print()}>
                <Printer className="h-3.5 w-3.5" /> Cetak
              </Button>
            </div>
          </div>
          <div className="max-h-[78vh] overflow-y-auto p-6">
            <div className="mx-auto bg-white p-10 text-sm text-slate-800 shadow-md" style={{ width: '210mm', minHeight: '297mm' }}>
              {children}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {open &&
        createPortal(
          <div className="print-portal-root hidden bg-white p-10 text-sm text-slate-800 print:block" style={{ width: '210mm' }}>
            {children}
          </div>,
          document.body,
        )}
    </>
  )
}
