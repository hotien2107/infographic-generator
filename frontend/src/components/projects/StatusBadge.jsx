import { Badge } from '@/components/ui/badge'
import { badgeVariantForStatus, documentStatusLabels, projectStatusLabels } from '@/lib/status'

export function StatusBadge({ status, type = 'project' }) {
  const labelMap = type === 'document' ? documentStatusLabels : projectStatusLabels
  return <Badge variant={badgeVariantForStatus(status)}>{labelMap[status] ?? status}</Badge>
}
