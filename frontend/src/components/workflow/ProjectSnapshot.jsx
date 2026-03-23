import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { formatDate } from '@/lib/format'

import { StatusBadge } from './StatusBadge'

export function ProjectSnapshot({ project }) {
  return (
    <Card className="bg-white/95">
      <CardHeader>
        <CardTitle className="text-2xl">Snapshot trạng thái</CardTitle>
        <CardDescription>Quan sát nhanh tiến độ của project và tài liệu đầu vào.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="rounded-lg bg-slate-50 p-4">
            <p className="text-sm text-muted-foreground">Project status</p>
            <div className="mt-2"><StatusBadge value={project?.status} /></div>
          </div>
          <div className="rounded-lg bg-slate-50 p-4">
            <p className="text-sm text-muted-foreground">Current step</p>
            <div className="mt-2"><StatusBadge kind="step" value={project?.current_step} /></div>
          </div>
        </div>
        <div className="rounded-lg border border-border bg-slate-50/80 p-4">
          <p className="text-sm text-muted-foreground">Documents attached</p>
          <p className="mt-2 text-3xl font-semibold">{project?.documents?.length ?? 0}</p>
        </div>
        <div className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
          {project ? (
            <div className="space-y-2">
              <p><span className="font-medium text-foreground">Project ID:</span> {project.id}</p>
              <p><span className="font-medium text-foreground">Tạo lúc:</span> {formatDate(project.created_at)}</p>
              <p><span className="font-medium text-foreground">Cập nhật:</span> {formatDate(project.updated_at)}</p>
            </div>
          ) : (
            'Chưa có project nào được tải. Hãy tạo project mới hoặc nhập UUID để xem dữ liệu.'
          )}
        </div>
      </CardContent>
    </Card>
  )
}
