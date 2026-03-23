import { AlertTriangle, CheckCircle2, Clock3, Loader2 } from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { formatDate } from '@/lib/format'

function Metric({ icon: Icon, label, value }) {
  return (
    <div className="rounded-lg border border-border bg-slate-50 p-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Icon className="h-4 w-4" />
        {label}
      </div>
      <p className="mt-2 text-2xl font-semibold">{value}</p>
    </div>
  )
}

export function ProcessingSummaryCard({ project }) {
  const summary = project?.processing_summary

  return (
    <Card className="bg-white/95">
      <CardHeader>
        <CardTitle className="text-2xl">Processing lifecycle</CardTitle>
        <CardDescription>Theo dõi queue, xử lý và kết quả extraction giả lập cho Sprint 2.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <Metric icon={Clock3} label="Queued" value={summary?.queued_documents ?? 0} />
          <Metric icon={Loader2} label="Processing" value={summary?.processing_documents ?? 0} />
          <Metric icon={CheckCircle2} label="Processed" value={summary?.processed_documents ?? 0} />
          <Metric icon={AlertTriangle} label="Failed" value={summary?.failed_documents ?? 0} />
        </div>
        <div className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
          <p><span className="font-medium text-foreground">Last processed at:</span> {formatDate(summary?.last_processed_at)}</p>
          <p className="mt-2"><span className="font-medium text-foreground">Last error:</span> {summary?.last_error ?? 'Không có lỗi gần nhất.'}</p>
        </div>
      </CardContent>
    </Card>
  )
}
