import { Loader2, RefreshCw, Upload, Workflow } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'

export function ProjectActionsPanel({
  projectId,
  inputMode,
  canUpload,
  canTriggerProcessing,
  isLoading,
  isUploading,
  isTriggering,
  onProjectIdChange,
  onLoadProject,
  onFileChange,
  onUpload,
  onTriggerProcessing,
}) {
  return (
    <div className="space-y-5 rounded-xl border border-dashed border-border bg-slate-50/80 p-5">
      <div>
        <h2 className="text-lg font-semibold">Kết nối backend</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Frontend sẽ gọi API thông qua <code className="rounded bg-slate-100 px-1 py-0.5">VITE_API_BASE_URL</code> hoặc proxy của Vite tại <code className="rounded bg-slate-100 px-1 py-0.5">http://localhost:8080</code>.
        </p>
      </div>
      <Separator />
      <div className="space-y-3">
        <Label htmlFor="project-id">Tải project theo ID</Label>
        <div className="flex flex-col gap-3 sm:flex-row">
          <Input id="project-id" value={projectId} onChange={(event) => onProjectIdChange(event.target.value)} placeholder="Dán UUID project để reload" />
          <Button type="button" variant="outline" onClick={onLoadProject} disabled={isLoading}>
            {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
            Tải lại
          </Button>
        </div>
      </div>
      <div className="space-y-3">
        <Label htmlFor="document">Upload tài liệu</Label>
        <Input id="document" type="file" accept=".pdf,.docx,.txt" onChange={(event) => onFileChange(event.target.files?.[0] ?? null)} />
        <div className="flex flex-wrap gap-3">
          <Button type="button" onClick={onUpload} disabled={!canUpload || isUploading}>
            {isUploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Upload className="h-4 w-4" />}
            Upload file
          </Button>
          <Button type="button" variant="outline" onClick={onTriggerProcessing} disabled={!canTriggerProcessing || isTriggering}>
            {isTriggering ? <Loader2 className="h-4 w-4 animate-spin" /> : <Workflow className="h-4 w-4" />}
            Trigger processing
          </Button>
        </div>
        {inputMode === 'text' ? (
          <p className="text-xs text-amber-700">Project mode đang là <strong>text</strong>, nhưng backend hiện chưa nhận nội dung text trực tiếp. UI giữ mode này để tránh hiểu nhầm về phạm vi hiện tại.</p>
        ) : null}
      </div>
    </div>
  )
}
