import { Card, CardContent } from '@/components/ui/card'

const tones = [
  'from-indigo-500/20 via-violet-500/15 to-sky-400/20',
  'from-sky-500/20 via-cyan-400/15 to-emerald-400/20',
  'from-fuchsia-500/20 via-rose-400/15 to-amber-300/20',
  'from-emerald-500/20 via-lime-400/15 to-cyan-300/20',
]

export function StatCard({ title, value, hint, tone = 0 }) {
  return (
    <Card className={`rounded-3xl border-white/70 bg-gradient-to-br ${tones[tone % tones.length]} shadow-sm`}>
      <CardContent className="space-y-3 p-6">
        <p className="text-sm font-medium text-muted-foreground">{title}</p>
        <p className="text-4xl font-semibold tracking-tight text-slate-950">{value}</p>
        <p className="text-sm text-slate-700/80">{hint}</p>
      </CardContent>
    </Card>
  )
}
