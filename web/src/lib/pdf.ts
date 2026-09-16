import { apiClient } from '@/lib/api'

// Fetches a server-rendered PDF (auth header attached by the apiClient
// interceptor) and opens it in a new tab as a blob URL. Falls back to a
// forced download if the popup is blocked.
export async function openServerPdf(url: string, filename: string): Promise<void> {
  const res = await apiClient.get(url, { responseType: 'blob' })
  const blobUrl = URL.createObjectURL(new Blob([res.data], { type: 'application/pdf' }))
  const win = window.open(blobUrl, '_blank')
  if (!win) {
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    a.remove()
  }
  setTimeout(() => URL.revokeObjectURL(blobUrl), 60_000)
}
