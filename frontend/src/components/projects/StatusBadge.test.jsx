import { render, screen } from '@testing-library/react'

import { StatusBadge } from './StatusBadge'

describe('StatusBadge', () => {
  it('renders user-friendly label for project status', () => {
    render(<StatusBadge status="processing" />)

    expect(screen.getByText('Đang xử lý')).toBeInTheDocument()
  })
})
