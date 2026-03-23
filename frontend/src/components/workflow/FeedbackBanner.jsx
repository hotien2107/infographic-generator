import { AlertCircle, CheckCircle2 } from 'lucide-react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

export function FeedbackBanner({ errorMessage, successMessage }) {
  if (!errorMessage && !successMessage) return null

  return (
    <Alert className={errorMessage ? 'border-red-200 bg-red-50' : 'border-emerald-200 bg-emerald-50'}>
      <div className="flex items-start gap-3">
        {errorMessage ? <AlertCircle className="mt-0.5 h-4 w-4 text-red-600" /> : <CheckCircle2 className="mt-0.5 h-4 w-4 text-emerald-600" />}
        <div>
          <AlertTitle>{errorMessage ? 'Có lỗi xảy ra' : 'Thành công'}</AlertTitle>
          <AlertDescription>{errorMessage ?? successMessage}</AlertDescription>
        </div>
      </div>
    </Alert>
  )
}
