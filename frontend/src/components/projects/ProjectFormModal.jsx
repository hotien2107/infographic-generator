import { useEffect, useState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

import { Modal } from '@/components/common/Modal'

const initialValues = {
  title: '',
  description: '',
  input_mode: 'file',
}

function normalizeInputMode(inputMode) {
  return inputMode === 'file' ? inputMode : 'file'
}

export function ProjectFormModal({ open, mode = 'create', project, isSubmitting = false, onClose, onSubmit }) {
  const [values, setValues] = useState(initialValues)

  useEffect(() => {
    if (!open) return
    if (project) {
      setValues({
        title: project.title ?? '',
        description: project.description ?? '',
        input_mode: normalizeInputMode(project.input_mode),
      })
      return
    }
    setValues(initialValues)
  }, [open, project])

  function updateField(field, value) {
    setValues((current) => ({ ...current, [field]: value }))
  }

  function handleSubmit(event) {
    event.preventDefault()
    onSubmit({
      title: values.title.trim(),
      description: values.description.trim(),
      input_mode: normalizeInputMode(values.input_mode),
    })
  }

  return (
    <Modal
      open={open}
      title={mode === 'create' ? 'Tạo dự án mới' : 'Chỉnh sửa dự án'}
      description="Điền những thông tin cơ bản để nhóm của bạn dễ theo dõi và tiếp tục xử lý nội dung."
      onClose={onClose}
    >
      <form className="space-y-5" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <Label htmlFor="project-title">Tên dự án</Label>
          <Input id="project-title" value={values.title} onChange={(event) => updateField('title', event.target.value)} placeholder="Ví dụ: Báo cáo chiến dịch tháng 3" required />
        </div>

        <div className="space-y-2">
          <Label htmlFor="project-description">Mô tả ngắn</Label>
          <Textarea
            id="project-description"
            value={values.description}
            onChange={(event) => updateField('description', event.target.value)}
            placeholder="Tóm tắt nội dung chính hoặc mục tiêu của dự án."
            rows={4}
          />
        </div>

        <div className="space-y-3">
          <Label>Loại đầu vào</Label>
          <div className="grid gap-3 sm:grid-cols-1">
            {[
              { value: 'file', label: 'Tài liệu tải lên', description: 'Dùng khi bạn có file PDF, DOCX hoặc TXT.' },
            ].map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => updateField('input_mode', option.value)}
                className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                  values.input_mode === option.value ? 'border-primary bg-primary/5' : 'border-border hover:bg-slate-50'
                }`}
              >
                <p className="font-medium text-slate-900">{option.label}</p>
                <p className="mt-1 text-sm leading-6 text-muted-foreground">{option.description}</p>
              </button>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-3">
          <Button type="button" variant="outline" onClick={onClose} disabled={isSubmitting}>
            Hủy
          </Button>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Đang lưu...' : mode === 'create' ? 'Tạo dự án' : 'Lưu thay đổi'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
