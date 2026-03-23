import { Card, CardContent } from '@/components/ui/card'

export function StatCard({ title, value, hint }) {
  return (
    <Card className="rounded-3xl border-white/70 bg-white/90 shadow-sm">
      <CardContent className="space-y-3 p-6">
        <p className="text-sm font-medium text-muted-foreground">{title}</p>
        <p className="text-4xl font-semibold tracking-tight text-slate-950">{value}</p>
        <p className="text-sm text-muted-foreground">{hint}</p>
      </CardContent>
    </Card>
  )
}
