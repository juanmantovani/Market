import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import Providers from './Providers'
import { routes } from './router'

const NOW = '2026-11-11T15:00:00Z'

function json(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
}

function page(items: unknown[], limit = 25) {
  return json({ items, total: items.length, page: 1, limit })
}

function stubApi() {
  const fetchMock = vi.fn(async (input: string) => {
    const url = new URL(input, 'http://localhost')
    switch (url.pathname) {
      case '/api/v1/suppliers':
        return page(
          [
            {
              id: 's1', businessName: 'Mayorista Norte', cuit: null, contactName: null, phone: null, email: null,
              address: null, notes: null, createdAt: NOW, updatedAt: NOW,
            },
          ],
          100,
        )
      case '/api/v1/categories':
        return page([{ id: 'c2', name: 'Bebidas', description: null, createdAt: NOW, updatedAt: NOW }], 100)
      case '/api/v1/stock/movements':
        return page([
          {
            id: 'm1', itemId: 'i1', type: 'entry', quantity: 12_000, balanceAfter: 30_000, reason: null,
            reference: 'Remito 4512', supplierId: 's1', unitCost: 220_000, actorId: 'dev', createdAt: NOW,
          },
        ])
      default:
        return page([
          {
            itemId: 'i1', code: 'BEB-001', name: 'Gaseosa cola', categoryId: 'c2', unit: 'unit', minStock: 12_000,
            quantity: 30_000, lowStock: false, updatedAt: NOW,
          },
        ])
    }
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('StockRoute', () => {
  it('shows the stock with category names and the suppliers in entries and movements', async () => {
    stubApi()
    const user = userEvent.setup()
    render(
      <Providers queryClient={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <RouterProvider router={createMemoryRouter(routes, { initialEntries: ['/stock'] })} />
      </Providers>,
    )

    const table = within(await screen.findByRole('table', { name: 'Stock' }))
    expect(await table.findByText('Bebidas')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Stock', level: 1 })).toBeInTheDocument()

    await user.click(table.getByRole('button', { name: 'Movimientos de Gaseosa cola' }))
    const history = within(await screen.findByRole('dialog', { name: 'Movimientos' }))
    expect(await history.findByText(/Mayorista Norte · Remito 4512/)).toBeInTheDocument()
    await user.click(history.getByRole('button', { name: 'Cerrar' }))

    await user.click(screen.getByRole('button', { name: 'Ingresar mercadería' }))
    const entry = within(await screen.findByRole('dialog', { name: 'Ingresar mercadería' }))
    expect(entry.getByRole('option', { name: 'Mayorista Norte' })).toBeInTheDocument()
  })
})
