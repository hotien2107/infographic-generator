export function LoadingState({ label = 'Đang tải dữ liệu...' }) {
  return (
    <div className="flex min-h-[240px] items-center justify-center rounded-3xl border border-border bg-slate-50/70">
      <div className="flex items-center gap-3 text-sm text-muted-foreground">
        <span className="h-3 w-3 animate-pulse rounded-full bg-primary" />
        {label}
      </div>
    </div>
  )
}
