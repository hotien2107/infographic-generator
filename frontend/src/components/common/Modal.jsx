import { X } from 'lucide-react'

import { Button } from '@/components/ui/button'

export function Modal({ open, title, description, children, onClose }) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4 py-6 backdrop-blur-sm">
      <div className="w-full max-w-2xl rounded-[28px] bg-white p-6 shadow-soft sm:p-8">
        <div className="flex items-start justify-between gap-4 border-b border-border pb-5">
          <div>
            <h3 className="text-2xl font-semibold text-slate-950">{title}</h3>
            {description ? <p className="mt-2 text-sm leading-6 text-muted-foreground">{description}</p> : null}
          </div>
          <Button variant="ghost" onClick={onClose} aria-label="Đóng">
            <X className="h-4 w-4" />
          </Button>
        </div>
        <div className="pt-6">{children}</div>
      </div>
    </div>
  )
}
