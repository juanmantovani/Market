import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import Providers from './Providers'
import { routes } from './router'

const NOW = '2026-11-10T15:00:00Z'

function json(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
}

function stubApi() {
  const fetchMock = vi.fn(async (input: string) => {
    const url = new URL(input, 'http://localhost')
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
      return json({
        items: [
          { id: 'c1', name: 'Almacén', description: null, createdAt: NOW, updatedAt: NOW },
          { id: 'c2', name: 'Bebidas', description: null, createdAt: NOW, updatedAt: NOW },
        ],
        total: 2,
        page: 1,
        limit: 100,
      })
    }
    return json({
      items: [
        {
          id: 'i1', code: 'BEB-001', barcode: null, name: 'Agua mineral', description: null, categoryId: 'c2', supplierIds: ['s1'],
          unit: 'unit', salePrice: 110_000, cost: null, taxRate: 2100, tracksStock: true, minStock: 0,
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

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ItemsRoute', () => {
  it('shows items with the names of their categories and offers the suppliers in the form', async () => {
    const fetchMock = stubApi()
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <Providers queryClient={queryClient}>
        <RouterProvider router={createMemoryRouter(routes, { initialEntries: ['/articulos'] })} />
      </Providers>,
    )

    const table = within(await screen.findByRole('table', { name: 'Artículos' }))
    expect(await table.findByText('Bebidas')).toBeInTheDocument()
    expect(table.getByText('Agua mineral')).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'Almacén' })).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/categories?page=1&limit=100',
      expect.objectContaining({ method: 'GET' }),
    )

    await userEvent.setup().click(table.getByRole('button', { name: 'Editar Agua mineral' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    expect(form.getByRole('checkbox', { name: 'Distribuidora del Centro' })).toBeChecked()
  })
})
