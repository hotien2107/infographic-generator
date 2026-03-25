export function PageHeader({ eyebrow, title, description, actions }) {
  return (
    <div className="flex flex-col gap-4 rounded-3xl border border-primary/10 bg-gradient-to-r from-white via-indigo-50/70 to-pink-50/80 p-4 sm:p-5 md:flex-row md:items-start md:justify-between">
      <div className="space-y-2">
        {eyebrow ? <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary/75">{eyebrow}</p> : null}
        <div className="space-y-1">
          <h2 className="text-2xl font-semibold tracking-tight text-slate-950 sm:text-3xl">{title}</h2>
          {description ? <p className="max-w-2xl text-sm leading-6 text-muted-foreground">{description}</p> : null}
        </div>
      </div>
      {actions ? <div className="flex w-full flex-wrap items-center gap-2 sm:gap-3 md:w-auto md:justify-end">{actions}</div> : null}
    </div>
  )
}
