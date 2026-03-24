export function PageHeader({ eyebrow, title, description, actions }) {
  return (
    <div className="flex flex-col gap-4 border-b border-border/70 pb-6 md:flex-row md:items-start md:justify-between">
      <div className="space-y-2">
        {eyebrow ? <p className="text-sm font-medium uppercase tracking-[0.2em] text-primary/70">{eyebrow}</p> : null}
        <div className="space-y-1">
          <h2 className="text-3xl font-semibold tracking-tight text-slate-950">{title}</h2>
          {description ? <p className="max-w-2xl text-sm leading-6 text-muted-foreground">{description}</p> : null}
        </div>
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-3">{actions}</div> : null}
    </div>
  )
}
