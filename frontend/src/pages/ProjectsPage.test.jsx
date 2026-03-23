import { render, screen, waitFor } from '@testing-library/react'

import { ProjectsPage } from './ProjectsPage'
import { AppRouter } from '@/router'

describe('ProjectsPage', () => {
  it('shows empty state when there are no projects', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], error: null, meta: {} }),
    })

    render(
      <AppRouter initialPath="/projects">
        <ProjectsPage />
      </AppRouter>,
    )

    await waitFor(() => expect(screen.getByText('Chưa có dự án nào')).toBeInTheDocument())
    expect(screen.getAllByText('Tạo dự án').length).toBeGreaterThan(0)

    global.fetch.mockRestore()
  })
})
