import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import Providers from './Providers'
import { routes } from './router'

const NOW = '2026-11-10T15:00:00Z'

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function stubApi({ stockFails = false, withoutMovements = false, categoryCount = 2 } = {}) {
  const fetchMock = vi.fn(async (input: string) => {
    const url = new URL(input, 'http://localhost')
    if (url.pathname === '/api/v1/stock') {
      if (stockFails) return json({ code: 'storage_unavailable', message: 'No se pudo acceder al almacenamiento' }, 503)
      return json({
        items: [
          {
            itemId: 'i1', code: 'BEB-001', name: 'Agua mineral', categoryId: 'c2', unit: 'unit', minStock: 24_000,
            quantity: 12_000, lowStock: true, updatedAt: withoutMovements ? null : NOW,
          },
        ],
        total: 1,
        page: 1,
        limit: 1,
      })
    }
    if (url.pathname === '/api/v1/suppliers') {
      return json({
        items: [
          {
            id: 's1', businessName: 'Distribuidora del Centro', cuit: null, contactName: null, phone: null, email: null,
            address: null, notes: null, createdAt: NOW, updatedAt: NOW,
          },
        ],
        total: 1,
        page: 1,
        limit: 100,
      })
    }
    if (url.pathname === '/api/v1/categories') {
      const page = Number(url.searchParams.get('page') ?? 1)
      const names = ['Almacén', 'Bebidas', ...Array.from({ length: Math.max(categoryCount - 2, 0) }, (_, i) => 'Categoría ' + (i + 3))]
      const all = names.map((name, i) => ({ id: 'c' + (i + 1), name, description: null, createdAt: NOW, updatedAt: NOW }))
      return json({ items: all.slice((page - 1) * 100, page * 100), total: all.length, page, limit: 100 })
    }
    return json({
      items: [
        {
          id: 'i1', code: 'BEB-001', barcode: null, name: 'Agua mineral', description: null, categoryId: 'c2', supplierIds: ['s1'],
          unit: 'unit', salePrice: 110_000, cost: null, taxRate: 2100, minStock: 24_000,
          createdAt: NOW, updatedAt: NOW,
        },
      ],
      total: 1,
      page: 1,
      limit: 25,
    })
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

function renderRoute() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <Providers queryClient={queryClient}>
      <RouterProvider router={createMemoryRouter(routes, { initialEntries: ['/articulos'] })} />
    </Providers>,
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ItemsRoute', () => {
  it('shows the stock of the items on the page, asking for just those', async () => {
    const fetchMock = stubApi()
    renderRoute()

    const table = within(await screen.findByRole('table', { name: 'Artículos' }))
    const row = within(await table.findByRole('row', { name: /Agua mineral/ }))
    expect(await row.findByText('12 u.')).toBeInTheDocument()
    expect(row.getByText('Stock bajo')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/stock?page=1&limit=1&itemId=i1', expect.objectContaining({ method: 'GET' }))
  })

  it('offers to adjust the stock, see its movements and receive goods from the items screen', async () => {
    stubApi()
    renderRoute()

    const table = within(await screen.findByRole('table', { name: 'Artículos' }))
    const row = within(await table.findByRole('row', { name: /Agua mineral/ }))
    expect(await row.findByRole('button', { name: 'Ajustar stock de Agua mineral' })).toBeEnabled()
    expect(row.getByRole('button', { name: 'Movimientos de Agua mineral' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Ingresar mercadería' })).toBeInTheDocument()

    const user = userEvent.setup()
    await user.click(row.getByRole('button', { name: 'Ajustar stock de Agua mineral' }))
    expect(await screen.findByRole('dialog', { name: 'Ajustar stock' })).toBeInTheDocument()
  })

  it('offers the first count instead of an adjustment for an item without movements', async () => {
    stubApi({ withoutMovements: true })
    renderRoute()

    const table = within(await screen.findByRole('table', { name: 'Artículos' }))
    const row = within(await table.findByRole('row', { name: /Agua mineral/ }))
    expect(await row.findByRole('button', { name: 'Cargar stock inicial de Agua mineral' })).toBeEnabled()
    expect(row.queryByRole('button', { name: /Ajustar stock/ })).not.toBeInTheDocument()
  })

  it('offers every category in the filter, even when there are more than a page of them', async () => {
    stubApi({ categoryCount: 230 })
    renderRoute()

    await screen.findByRole('table', { name: 'Artículos' })
    expect(await screen.findByRole('option', { name: 'Categoría 230' })).toBeInTheDocument()
    expect(screen.getAllByRole('option', { name: /^Categoría / })).toHaveLength(228)
  })

  it('keeps listing the items when their stock cannot be loaded', async () => {
    stubApi({ stockFails: true })
    renderRoute()

    const table = within(await screen.findByRole('table', { name: 'Artículos' }))
    const row = within(await table.findByRole('row', { name: /Agua mineral/ }))
    expect(await row.findByText('Stock no disponible')).toBeInTheDocument()
  })

  it('shows items with the names of their categories and offers the suppliers in the form', async () => {
    const fetchMock = stubApi()
    renderRoute()

    const table = within(await screen.findByRole('table', { name: 'Artículos' }))
    expect(await table.findByText('Bebidas')).toBeInTheDocument()
    expect(table.getByText('Agua mineral')).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'Almacén' })).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/categories?page=1&limit=100',
      expect.objectContaining({ method: 'GET' }),
    )

    const user = userEvent.setup()
    await user.click(table.getByRole('button', { name: 'Editar Agua mineral' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    await user.click(form.getByRole('button', { name: 'Proveedores · 1 elegido' }))
    expect(form.getByRole('checkbox', { name: 'Distribuidora del Centro' })).toBeChecked()
  })
})
