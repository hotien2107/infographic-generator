import { render, screen } from '@testing-library/react'

import { StatusBadge } from './StatusBadge'

describe('StatusBadge', () => {
  it('renders processed document label', () => {
    render(<StatusBadge kind="document" value="processed" />)

    expect(screen.getByText('Processed')).toBeInTheDocument()
  })

  it('renders failed step label', () => {
    render(<StatusBadge kind="step" value="failed" />)

    expect(screen.getByText('Failed')).toBeInTheDocument()
  })
})
