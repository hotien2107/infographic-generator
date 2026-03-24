import { LayoutGrid, FolderOpen } from 'lucide-react'
import { NavLink } from '@/router'

import { navigationItems } from '@/constants/navigation'
import { cn } from '@/lib/utils'

const icons = {
  '/dashboard': LayoutGrid,
  '/projects': FolderOpen,
}

export function AppShell({ children }) {
  return (
    <div className="min-h-screen px-4 py-6 text-foreground sm:px-6 lg:px-8">
      <div className="mx-auto flex min-h-[calc(100vh-3rem)] max-w-7xl gap-6 lg:flex-row">
        <aside className="w-full rounded-[28px] border border-white/70 bg-white/85 p-5 shadow-soft backdrop-blur lg:sticky lg:top-6 lg:h-fit lg:w-72">
          <div className="space-y-2">
            <p className="text-sm font-medium uppercase tracking-[0.2em] text-primary/70">Workspace</p>
            <h1 className="text-2xl font-semibold text-slate-950">Trung tâm dự án</h1>
            <p className="text-sm leading-6 text-muted-foreground">
              Theo dõi tiến độ tạo infographic, quản lý tài liệu và cập nhật dự án trong một nơi duy nhất.
            </p>
          </div>

          <nav className="mt-8 space-y-2">
            {navigationItems.map((item) => {
              const Icon = icons[item.href]
              return (
                <NavLink
                  key={item.href}
                  to={item.href}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-medium transition-colors',
                      isActive ? 'bg-primary text-white shadow-lg shadow-primary/20' : 'text-slate-700 hover:bg-slate-100',
                    )
                  }
                >
                  <Icon className="h-4 w-4" />
                  {item.label}
                </NavLink>
              )
            })}
          </nav>
        </aside>

        <div className="min-w-0 flex-1 rounded-[28px] border border-white/70 bg-white/80 p-5 shadow-soft backdrop-blur sm:p-8">
          {children}
        </div>
      </div>
    </div>
  )
}
