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
    <div className="min-h-screen px-3 py-4 text-foreground sm:px-5 sm:py-6 lg:px-8">
      <div className="mx-auto flex min-h-[calc(100vh-2rem)] max-w-7xl flex-col gap-4 lg:min-h-[calc(100vh-3rem)] lg:flex-row lg:gap-6">
        <aside className="w-full rounded-[28px] border border-white/70 bg-gradient-to-br from-indigo-600/95 via-violet-600/90 to-sky-500/90 p-5 text-white shadow-soft backdrop-blur lg:sticky lg:top-6 lg:h-fit lg:w-72">
          <div className="space-y-2">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-white/80">Workspace</p>
            <h1 className="text-2xl font-semibold">Trung tâm dự án</h1>
            <p className="text-sm leading-6 text-white/85">Theo dõi tiến độ tạo infographic, quản lý tài liệu và cập nhật dự án trong một nơi duy nhất.</p>
          </div>

          <nav className="mt-8 grid gap-2 sm:grid-cols-2 lg:grid-cols-1">
            {navigationItems.map((item) => {
              const Icon = icons[item.href]
              return (
                <NavLink
                  key={item.href}
                  to={item.href}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-medium transition-all',
                      isActive
                        ? 'bg-white text-indigo-700 shadow-lg shadow-indigo-950/20'
                        : 'bg-white/10 text-white hover:bg-white/20 hover:shadow-md hover:shadow-indigo-900/20',
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

        <div className="min-w-0 flex-1 rounded-[28px] border border-white/70 bg-white/85 p-4 shadow-soft backdrop-blur sm:p-6 lg:p-8">{children}</div>
      </div>
    </div>
  )
}
