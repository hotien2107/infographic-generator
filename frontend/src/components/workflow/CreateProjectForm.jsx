import { FileText, Loader2, Sparkles, Upload } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

export function CreateProjectForm({ title, inputMode, notes, isCreating, onTitleChange, onInputModeChange, onNotesChange, onSubmit }) {
  return (
    <form className="space-y-5" onSubmit={onSubmit}>
      <div className="space-y-2">
        <Label htmlFor="title">Tên project</Label>
        <Input id="title" value={title} onChange={(event) => onTitleChange(event.target.value)} placeholder="Ví dụ: Báo cáo tăng trưởng Q2" />
      </div>
      <div className="space-y-3">
        <Label>Input mode</Label>
        <div className="grid gap-3 sm:grid-cols-2">
          {['file', 'text'].map((mode) => (
            <button
              key={mode}
              type="button"
              onClick={() => onInputModeChange(mode)}
              className={`rounded-lg border p-4 text-left transition ${
                inputMode === mode ? 'border-blue-600 bg-blue-50 shadow-sm' : 'border-border bg-white hover:border-blue-200'
              }`}
            >
              <div className="mb-2 flex items-center gap-2 font-medium capitalize">
                {mode === 'file' ? <Upload className="h-4 w-4" /> : <FileText className="h-4 w-4" />}
                {mode}
              </div>
              <p className="text-sm text-muted-foreground">
                {mode === 'file'
                  ? 'Dành cho tài liệu gốc như PDF, DOCX hoặc TXT để frontend upload trực tiếp.'
                  : 'Luồng nhập text trực tiếp chưa được triển khai ở Sprint 2. Mode này được giữ lại để frontend/backend cùng bám contract.'}
              </p>
            </button>
          ))}
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="notes">Ghi chú cho đội nội dung</Label>
        <Textarea id="notes" value={notes} onChange={(event) => onNotesChange(event.target.value)} placeholder="Mô tả scope infographic mong muốn..." />
        <p className="text-xs text-muted-foreground">Trường này hiện chỉ phục vụ UI note, không gửi sang backend ở Sprint 1/2.</p>
      </div>
      <Button type="submit" className="w-full sm:w-auto" disabled={isCreating}>
        {isCreating ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
        Tạo project
      </Button>
    </form>
  )
}
