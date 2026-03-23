import { AppShell } from '@/components/layout/AppShell'
import { DashboardPage } from '@/pages/DashboardPage'
import { ProjectDetailPage } from '@/pages/ProjectDetailPage'
import { ProjectsPage } from '@/pages/ProjectsPage'
import { AppRouter, usePathname } from '@/router'

function AppContent() {
  const pathname = usePathname()

  if (pathname === '/') {
    if (typeof window !== 'undefined') {
      window.history.replaceState({}, '', '/dashboard')
    }
    return <DashboardPage />
  }

  if (pathname === '/dashboard') return <DashboardPage />
  if (pathname === '/projects') return <ProjectsPage />
  if (/^\/projects\/[^/]+$/.test(pathname)) return <ProjectDetailPage />

  return <ProjectsPage />
}

function App() {
  return (
    <AppRouter>
      <AppShell>
        <AppContent />
      </AppShell>
    </AppRouter>
  )
}

export default App
