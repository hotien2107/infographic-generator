import { render, screen } from '@testing-library/react'

import { StatusBadge } from './StatusBadge'

describe('StatusBadge', () => {
  it('renders user-friendly label for project status', () => {
    render(<StatusBadge status="extracting" />)

    expect(screen.getByText('Đang trích xuất')).toBeInTheDocument()
  })
})
