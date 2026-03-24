import { useEffect, useState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

import { Modal } from '@/components/common/Modal'

export function DocumentNameModal({ open, document, isSubmitting = false, onClose, onSubmit }) {
  const [filename, setFilename] = useState('')

  useEffect(() => {
    setFilename(document?.filename ?? '')
  }, [document])

  function handleSubmit(event) {
    event.preventDefault()
    onSubmit(filename.trim())
  }

  return (
    <Modal open={open} title="Cập nhật tên tài liệu" description="Đổi tên hiển thị để nhóm của bạn dễ nhận biết hơn." onClose={onClose}>
      <form className="space-y-5" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <Label htmlFor="document-name">Tên tài liệu</Label>
          <Input id="document-name" value={filename} onChange={(event) => setFilename(event.target.value)} required />
        </div>
        <div className="flex justify-end gap-3">
          <Button type="button" variant="outline" onClick={onClose} disabled={isSubmitting}>
            Hủy
          </Button>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Đang lưu...' : 'Lưu thay đổi'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
