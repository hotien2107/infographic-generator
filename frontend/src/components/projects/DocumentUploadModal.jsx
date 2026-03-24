import { useEffect, useState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

import { Modal } from '@/components/common/Modal'

export function DocumentUploadModal({ open, isSubmitting = false, onClose, onSubmit }) {
  const [file, setFile] = useState(null)

  useEffect(() => {
    if (!open) {
      setFile(null)
    }
  }, [open])

  function handleSubmit(event) {
    event.preventDefault()
    if (!file) return
    onSubmit(file)
  }

  return (
    <Modal open={open} title="Thêm tài liệu" description="Chọn tệp để bổ sung nội dung cho dự án này." onClose={onClose}>
      <form className="space-y-5" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <Label htmlFor="document-upload">Tệp tài liệu</Label>
          <Input id="document-upload" type="file" accept=".pdf,.docx,.txt" onChange={(event) => setFile(event.target.files?.[0] ?? null)} />
          <p className="text-sm text-muted-foreground">Hỗ trợ PDF, DOCX và TXT.</p>
        </div>

        <div className="flex justify-end gap-3">
          <Button type="button" variant="outline" onClick={onClose} disabled={isSubmitting}>
            Hủy
          </Button>
          <Button type="submit" disabled={!file || isSubmitting}>
            {isSubmitting ? 'Đang tải lên...' : 'Thêm tài liệu'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
