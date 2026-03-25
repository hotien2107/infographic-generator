import { X } from 'lucide-react'

import { Button } from '@/components/ui/button'

export function Modal({ open, title, description, children, onClose }) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-slate-950/45 px-3 py-3 backdrop-blur-sm sm:items-center sm:px-4 sm:py-6">
      <div className="max-h-[92vh] w-full max-w-2xl overflow-hidden rounded-[24px] border border-white/50 bg-white shadow-soft sm:rounded-[28px]">
        <div className="flex items-start justify-between gap-4 border-b border-primary/10 bg-gradient-to-r from-indigo-50 via-sky-50 to-pink-50 px-5 py-4 sm:px-7 sm:py-6">
          <div>
            <h3 className="text-xl font-semibold text-slate-950 sm:text-2xl">{title}</h3>
            {description ? <p className="mt-2 text-sm leading-6 text-muted-foreground">{description}</p> : null}
          </div>
          <Button variant="ghost" size="sm" onClick={onClose} aria-label="Đóng">
            <X className="h-4 w-4" />
          </Button>
        </div>
        <div className="max-h-[calc(92vh-7rem)] overflow-y-auto px-5 py-5 sm:px-7 sm:py-6">{children}</div>
      </div>
    </div>
  )
}
