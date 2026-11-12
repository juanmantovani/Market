import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import DashboardPage from './DashboardPage'

describe('DashboardPage', () => {
  it('explains what the home page will show before there are sales', () => {
    render(<DashboardPage />)

    expect(screen.getByRole('heading', { level: 1, name: 'Inicio' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Todavía no hay ventas hoy' })).toBeInTheDocument()
  })
})
