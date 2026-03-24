import { AlertCircle } from 'lucide-react'

import { Button } from '@/components/ui/button'

export function ErrorState({ title = 'Có lỗi xảy ra', description, onRetry }) {
  return (
    <div className="rounded-3xl border border-red-200 bg-red-50/80 px-6 py-8">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="flex items-start gap-3">
          <div className="rounded-full bg-white p-2 text-red-500 shadow-sm">
            <AlertCircle className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-red-900">{title}</h3>
            <p className="mt-1 text-sm leading-6 text-red-700">{description}</p>
          </div>
        </div>
        {onRetry ? (
          <Button variant="outline" onClick={onRetry}>
            Thử lại
          </Button>
        ) : null}
      </div>
    </div>
  )
}
