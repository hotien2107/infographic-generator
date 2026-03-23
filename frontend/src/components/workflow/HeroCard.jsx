import { RefreshCw, Sparkles, Upload } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

import { CreateProjectForm } from './CreateProjectForm'
import { ProjectActionsPanel } from './ProjectActionsPanel'

export function HeroCard(props) {
  const { title, inputMode, notes, isCreating, onTitleChange, onInputModeChange, onNotesChange, onSubmit } = props

  return (
    <Card className="overflow-hidden bg-white/95 lg:col-span-2">
      <CardHeader className="space-y-4 bg-slate-950 text-slate-50">
        <Badge variant="secondary" className="w-fit bg-white/10 text-white">
          Sprint 1 hardening + Sprint 2 processing foundation
        </Badge>
        <div className="space-y-3">
          <CardTitle className="text-4xl font-bold leading-tight">AI Infographic Generator</CardTitle>
          <CardDescription className="max-w-2xl text-slate-200">
            Giao diện frontend bám response envelope chuẩn, hiển thị rõ validation/error và mở rộng thêm document ingestion lifecycle cho Sprint 2.
          </CardDescription>
        </div>
        <div className="grid gap-3 sm:grid-cols-3">
          <div className="rounded-lg border border-white/10 bg-white/5 p-4">
            <div className="mb-2 flex items-center gap-2 text-sm font-medium"><Sparkles className="h-4 w-4" />Create project</div>
            <p className="text-sm text-slate-300">Khởi tạo luồng file/text và giữ contract rõ ràng cho UI state.</p>
          </div>
          <div className="rounded-lg border border-white/10 bg-white/5 p-4">
            <div className="mb-2 flex items-center gap-2 text-sm font-medium"><Upload className="h-4 w-4" />Document ingestion</div>
            <p className="text-sm text-slate-300">Upload file, nhận trạng thái processing và theo dõi extraction giả lập.</p>
          </div>
          <div className="rounded-lg border border-white/10 bg-white/5 p-4">
            <div className="mb-2 flex items-center gap-2 text-sm font-medium"><RefreshCw className="h-4 w-4" />Refresh detail</div>
            <p className="text-sm text-slate-300">Reload project để quan sát worker chuyển state trong thời gian thực.</p>
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-6 p-6 lg:grid-cols-[1.1fr_0.9fr]">
        <CreateProjectForm
          title={title}
          inputMode={inputMode}
          notes={notes}
          isCreating={isCreating}
          onTitleChange={onTitleChange}
          onInputModeChange={onInputModeChange}
          onNotesChange={onNotesChange}
          onSubmit={onSubmit}
        />
        <ProjectActionsPanel {...props} />
      </CardContent>
    </Card>
  )
}
