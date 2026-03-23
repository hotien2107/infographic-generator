import { Badge } from '@/components/ui/badge'
import { badgeVariantForStatus, documentStatusLabels, projectStatusLabels, stepLabels } from '@/lib/status'

export function StatusBadge({ kind = 'project', value }) {
  const labels = kind === 'step' ? stepLabels : kind === 'document' ? documentStatusLabels : projectStatusLabels

  return <Badge variant={badgeVariantForStatus(value)}>{labels[value] ?? value ?? '—'}</Badge>
}
