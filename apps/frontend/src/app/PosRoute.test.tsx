import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Providers from './Providers'
import { routes } from './router'

const NOW = '2026-11-12T15:00:00Z'

function json(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
}

function item(id: string, code: string, name: string, categoryId: string, salePrice: number) {
  return {
    id, code, barcode: null, name, description: null, categoryId, supplierIds: [], unit: 'unit', salePrice,
    cost: null, taxRate: 2100, minStock: 0, createdAt: NOW, updatedAt: NOW,
  }
}

function stubApi() {
  const fetchMock = vi.fn(async (input: string) => {
    const url = new URL(input, 'http://localhost')
    if (url.pathname === '/api/v1/categories') {
      return json({ items: [{ id: 'c1', name: 'Bebidas', description: null, createdAt: NOW, updatedAt: NOW }], total: 1, page: 1, limit: 100 })
    }
    if (url.pathname === '/api/v1/customers') {
      const customer = {
        id: 'cu1', name: 'Ana Pérez', documentType: 'dni', documentNumber: '28456789', taxCondition: 'final_consumer',
        email: null, phone: null, address: null, notes: null, system: false, createdAt: NOW, updatedAt: NOW,
      }
      return json({ items: [customer], total: 1, page: 1, limit: 20 })
    }
    if (url.pathname === '/api/v1/sellers') {
      return json({ items: [{ id: 's1', name: 'Juan Pérez', code: 'JP', commissionRate: null, createdAt: NOW, updatedAt: NOW }], total: 1, page: 1, limit: 20 })
    }
    const items = [item('i1', 'BEB-001', 'Gaseosa cola', 'c1', 320_000), item('i2', 'ALM-001', 'Yerba mate', 'c2', 450_000)].filter(
      (current) => !url.searchParams.get('categoryId') || current.categoryId === url.searchParams.get('categoryId'),
    )
    return json({ items, total: items.length, page: 1, limit: Number(url.searchParams.get('limit')) })
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

beforeEach(() => {
  window.localStorage.clear()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('PosRoute', () => {
  it('sells from the items catalog with the category chips', async () => {
    const fetchMock = stubApi()
    const user = userEvent.setup()
    render(
      <Providers queryClient={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <RouterProvider router={createMemoryRouter(routes, { initialEntries: ['/'] })} />
      </Providers>,
    )

    const catalog = within(await screen.findByRole('list', { name: 'Artículos' }))
    expect(await catalog.findByRole('button', { name: /Yerba mate/ })).toBeInTheDocument()
    await waitFor(() => expect(document.title).toBe('Punto de venta · Market'))

    await user.click(await screen.findByRole('button', { name: 'Bebidas' }))
    await waitFor(() => expect(catalog.queryByRole('button', { name: /Yerba mate/ })).not.toBeInTheDocument())
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/api/v1/items?page=1&limit=24&categoryId=c1'), expect.anything())

    await user.click(catalog.getByRole('button', { name: /Gaseosa cola/ }))
    const sale = within(screen.getByRole('region', { name: 'Venta actual' }))
    expect(sale.getByText('Gaseosa cola')).toBeInTheDocument()

    await user.click(sale.getByRole('button', { name: /^Cliente:/ }))
    await user.click(await within(await screen.findByRole('dialog', { name: 'Elegir cliente' })).findByRole('button', { name: /Ana Pérez/ }))
    await user.click(sale.getByRole('button', { name: /^Vendedor:/ }))
    await user.click(await within(await screen.findByRole('dialog', { name: 'Elegir vendedor' })).findByRole('button', { name: /Juan Pérez/ }))
    expect(sale.getByRole('button', { name: /^Cliente: Ana Pérez/ })).toBeInTheDocument()
    expect(sale.getByRole('button', { name: /^Vendedor: Juan Pérez/ })).toBeInTheDocument()
  })
})
