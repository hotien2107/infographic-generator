import { cn } from '@/lib/utils'

export function Alert({ className, ...props }) {
  return <div className={cn('relative w-full rounded-lg border border-border bg-white px-4 py-3 text-sm', className)} {...props} />
}

export function AlertTitle({ className, ...props }) {
  return <h5 className={cn('mb-1 font-medium leading-none tracking-tight', className)} {...props} />
}

export function AlertDescription({ className, ...props }) {
  return <div className={cn('text-sm text-muted-foreground', className)} {...props} />
}
