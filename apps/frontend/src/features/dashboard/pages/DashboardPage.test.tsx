import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import DashboardPage from './DashboardPage'

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

const lowItem = {
  itemId: 'i1',
  code: 'COCA',
  name: 'Gaseosa cola',
  categoryId: 'c1',
  unit: 'unit',
  minStock: 5000,
  quantity: 2000,
  lowStock: true,
  updatedAt: null,
}

function summary(overrides: Record<string, unknown> = {}) {
  return {
    from: '2026-11-12T03:00:00Z',
    to: '2026-11-13T03:00:00Z',
    sales: {
      count: 3,
      total: 330_000,
      byMethod: [
        { method: 'cash', amount: 180_000 },
        { method: 'qr', amount: 150_000 },
      ],
    },
    lowStock: { total: 1, items: [lowItem] },
    ...overrides,
  }
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('DashboardPage', () => {
  it('shows the sales of the day, the total by payment method and the low stock items', async () => {
    const fetchMock = vi.fn(async (_input: string) => json(summary()))
    vi.stubGlobal('fetch', fetchMock)

    renderWithProviders(<DashboardPage />)

    expect(screen.getByRole('status', { name: 'Cargando el resumen' })).toBeInTheDocument()
    const sales = (await screen.findByRole('heading', { name: 'Ventas de hoy' })).closest('section')!
    expect(within(sales).getByText('3')).toBeInTheDocument()
    const total = screen.getByRole('heading', { name: 'Total vendido' }).closest('section')!
    expect(total).toHaveTextContent(/3\.300,00/)

    const methods = screen.getByRole('heading', { name: 'Total por medio de pago' }).closest('section')!
    expect(within(methods).getByText('Efectivo')).toBeInTheDocument()
    expect(within(methods).getByText('QR')).toBeInTheDocument()
    expect(methods).toHaveTextContent(/1\.800,00/)

    const low = screen.getByRole('heading', { name: 'Artículos con stock bajo' }).closest('section')!
    expect(within(low).getByText('Gaseosa cola')).toBeInTheDocument()
    expect(within(low).getByText(/Stock bajo: quedan/)).toBeInTheDocument()
    expect(low).toHaveTextContent('Mínimo 5 u.')

    const url = new URL(fetchMock.mock.calls[0][0], 'http://localhost')
    expect(url.pathname).toBe('/api/v1/dashboard/summary')
    const from = new Date(url.searchParams.get('from')!)
    const to = new Date(url.searchParams.get('to')!)
    expect(from.getHours()).toBe(0)
    expect(to.getTime() - from.getTime()).toBeGreaterThanOrEqual(23 * 3_600_000)
  })

  it('says so when there are no sales and no low stock', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => json(summary({ sales: { count: 0, total: 0, byMethod: [] }, lowStock: { total: 0, items: [] } }))),
    )

    renderWithProviders(<DashboardPage />)

    expect(await screen.findByText('Todavía no hay ventas hoy.')).toBeInTheDocument()
    expect(screen.getByText('Ningún artículo llegó a su stock mínimo.')).toBeInTheDocument()
  })

  it('tells how many low stock items are not listed', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => json(summary({ lowStock: { total: 23, items: [lowItem] } }))))

    renderWithProviders(<DashboardPage />)

    expect(await screen.findByText(/y 22 más en Artículos/)).toBeInTheDocument()
  })

  it('shows the error and tries again', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(json({ code: 'storage_unavailable', message: 'El almacenamiento no responde' }, 503))
      .mockResolvedValue(json(summary()))
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()

    renderWithProviders(<DashboardPage />)

    expect(await screen.findByText('No se pudo cargar el resumen')).toBeInTheDocument()
    expect(screen.getByText('El almacenamiento no responde')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByRole('heading', { name: 'Ventas de hoy' })).toBeInTheDocument()
  })
})
