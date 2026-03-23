import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function ApiChecklistCard() {
  return (
    <Card className="bg-white/95">
      <CardHeader>
        <CardTitle className="text-2xl">Checklist tích hợp</CardTitle>
        <CardDescription>Các API backend mà frontend này đang tiêu thụ theo Sprint 1 + Sprint 2.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4 text-sm text-slate-700">
        <div className="rounded-lg border border-border p-4">
          <p className="font-medium">POST /api/v1/projects</p>
          <p className="mt-1 text-muted-foreground">Tạo project mới với title và input_mode.</p>
        </div>
        <div className="rounded-lg border border-border p-4">
          <p className="font-medium">GET /api/v1/projects/:projectId</p>
          <p className="mt-1 text-muted-foreground">Tải chi tiết project, processing summary và danh sách document.</p>
        </div>
        <div className="rounded-lg border border-border p-4">
          <p className="font-medium">POST /api/v1/projects/:projectId/documents</p>
          <p className="mt-1 text-muted-foreground">Upload file đầu vào và auto-enqueue sang worker xử lý giả lập.</p>
        </div>
        <div className="rounded-lg border border-border p-4">
          <p className="font-medium">POST /api/v1/projects/:projectId/processing</p>
          <p className="mt-1 text-muted-foreground">Cho phép trigger processing thủ công cho document mới nhất.</p>
        </div>
      </CardContent>
    </Card>
  )
}
