import { render, screen } from '@testing-library/react'

import { ProcessingSummaryCard } from './ProcessingSummaryCard'

describe('ProcessingSummaryCard', () => {
  it('renders main processing counters', () => {
    render(
      <ProcessingSummaryCard
        project={{
          processing_summary: {
            queued_documents: 1,
            processing_documents: 2,
            processed_documents: 3,
            failed_documents: 1,
            last_processed_at: '2026-03-23T10:00:00Z',
            last_error: 'worker failed',
          },
        }}
      />,
    )

    expect(screen.getByText('Queued')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('worker failed')).toBeInTheDocument()
  })
})
