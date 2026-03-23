import { Button } from '@/components/ui/button'

import { Modal } from '@/components/common/Modal'

export function ConfirmDialog({ open, title, description, confirmLabel = 'Xóa', cancelLabel = 'Hủy', isSubmitting = false, onConfirm, onClose }) {
  return (
    <Modal open={open} title={title} description={description} onClose={onClose}>
      <div className="flex justify-end gap-3">
        <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
          {cancelLabel}
        </Button>
        <Button variant="default" onClick={onConfirm} disabled={isSubmitting}>
          {isSubmitting ? 'Đang xử lý...' : confirmLabel}
        </Button>
      </div>
    </Modal>
  )
}
