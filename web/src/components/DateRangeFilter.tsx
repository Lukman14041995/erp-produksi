import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'

export function DateRangeFilter({
  from,
  to,
  onFromChange,
  onToChange,
}: {
  from: string
  to: string
  onFromChange: (v: string) => void
  onToChange: (v: string) => void
}) {
  return (
    <div className="flex items-end gap-3 print:hidden">
      <div className="space-y-1.5">
        <Label>Dari</Label>
        <Input type="date" value={from} onChange={(e) => onFromChange(e.target.value)} />
      </div>
      <div className="space-y-1.5">
        <Label>Sampai</Label>
        <Input type="date" value={to} onChange={(e) => onToChange(e.target.value)} />
      </div>
    </div>
  )
}

export function AsOfFilter({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <div className="space-y-1.5 print:hidden">
      <Label>Per Tanggal</Label>
      <Input type="date" value={value} onChange={(e) => onChange(e.target.value)} />
    </div>
  )
}
