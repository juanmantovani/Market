import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import type { SaleRecord } from '../types'
import SalesPage from './SalesPage'

type Failure = { status: number; code: string; message: string }

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function sale(number: number, overrides: Partial<SaleRecord> = {}): SaleRecord {
  return {
    id: `sa${number}`,
    number,
    status: 'completed',
    customer: { id: 'cu1', name: 'Consumidor Final' },
    seller: null,
    lines: [
      {
        itemId: 'it1',
        code: 'A1',
        name: 'Yerba 1 kg',
        unit: 'unit',
        quantity: 2000,
        unitPrice: 500000,
        discount: null,
        discountAmount: 0,
        cartDiscountAmount: 0,
        lineTotal: 1000000,
      },
    ],
    subtotal: 1000000,
    cartDiscount: null,
    cartDiscountAmount: 0,
    discountTotal: 0,
    total: 1000000,
    payments: [{ method: 'cash', amount: 1000000, reference: null }],
    change: 0,
    createdAt: '2026-11-10T15:00:00Z',
    voidedAt: null,
    voidReason: null,
    ...overrides,
  }
}

// An in-memory stand-in for the sales endpoints: filters by status and
// dates like the backend and voids a completed sale.
function fakeSalesApi(initial: SaleRecord[]) {
  const sales = [...initial]
  const requests: Array<{ method: string; path: string; body: unknown }> = []
  const failures: Failure[] = []

  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const url = new URL(input, 'http://localhost')
      const method = init.method ?? 'GET'
      const body = init.body ? JSON.parse(String(init.body)) : undefined
      requests.push({ method, path: `${url.pathname}${url.search}`, body })

      const failure = failures.shift()
      if (failure) return json({ code: failure.code, message: failure.message }, failure.status)

      if (url.pathname === '/api/v1/sales' && method === 'GET') {
        const status = url.searchParams.get('status')
        const page = Number(url.searchParams.get('page') ?? 1)
        const limit = Number(url.searchParams.get('limit') ?? 25)
        const matches = sales.filter((entry) => !status || entry.status === status)
        return json({ items: matches.slice((page - 1) * limit, page * limit), total: matches.length, page, limit })
      }

      const voidMatch = /^\/api\/v1\/sales\/([^/]+)\/void$/.exec(url.pathname)
      if (voidMatch && method === 'POST') {
        const target = sales.find((entry) => entry.id === voidMatch[1])
        if (!target) return json({ code: 'sale_not_found', message: 'La venta no existe' }, 404)
        Object.assign(target, { status: 'voided', voidedAt: '2026-11-11T10:00:00Z', voidReason: body.reason })
        return json(target)
      }

      return json({ code: 'not_found', message: 'No existe' }, 404)
    }),
  )

  return { sales, requests, failures }
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('SalesPage', () => {
  it('lists the sales with number, customer, seller, total and a written status', async () => {
    fakeSalesApi([
      sale(2, { status: 'voided', voidedAt: '2026-11-11T10:00:00Z', voidReason: 'Error de carga', seller: { id: 'se1', name: 'Ana', code: 'A' } }),
      sale(1),
    ])
    renderWithProviders(<SalesPage />)

    const table = await screen.findByRole('table', { name: 'Ventas' })
    const rows = within(table).getAllByRole('row').slice(1)
    expect(rows).toHaveLength(2)
    expect(within(rows[0]).getByText('N.º 2')).toBeInTheDocument()
    expect(within(rows[0]).getByText('Ana')).toBeInTheDocument()
    expect(within(rows[0]).getByText('Anulada')).toBeInTheDocument()
    expect(within(rows[1]).getByText('Consumidor Final')).toBeInTheDocument()
    expect(within(rows[1]).getByText('Sin vendedor')).toBeInTheDocument()
    expect(within(rows[1]).getByText('Completada')).toBeInTheDocument()
    expect(within(rows[1]).getByText(/10\.000,00/)).toBeInTheDocument()
  })

  it('shows an empty state when there are no sales', async () => {
    fakeSalesApi([])
    renderWithProviders(<SalesPage />)

    expect(await screen.findByRole('heading', { name: 'Todavía no hay ventas' })).toBeInTheDocument()
  })

  it('filters by status and by days, sending the day range as local instants', async () => {
    const api = fakeSalesApi([sale(2, { status: 'voided' }), sale(1)])
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)
    await screen.findByRole('table', { name: 'Ventas' })

    await user.selectOptions(screen.getByLabelText('Estado'), 'completed')
    await waitFor(() => expect(within(screen.getByRole('table', { name: 'Ventas' })).getAllByRole('row')).toHaveLength(2))
    expect(within(screen.getByRole('table', { name: 'Ventas' })).queryByText('Anulada')).not.toBeInTheDocument()

    await user.type(screen.getByLabelText('Desde'), '2026-11-10')
    await user.type(screen.getByLabelText('Hasta'), '2026-11-12')
    await waitFor(() => {
      const last = new URL(api.requests.at(-1)!.path, 'http://localhost').searchParams
      expect(last.get('status')).toBe('completed')
      expect(last.get('from')).toBe(new Date(2026, 10, 10).toISOString())
      expect(last.get('to')).toBe(new Date(2026, 10, 13).toISOString())
    })
  })

  it('explains when the filters leave nothing and warns about an inverted range', async () => {
    fakeSalesApi([sale(1)])
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)
    await screen.findByRole('table', { name: 'Ventas' })

    await user.selectOptions(screen.getByLabelText('Estado'), 'voided')
    expect(await screen.findByRole('heading', { name: 'No hay ventas con estos filtros' })).toBeInTheDocument()

    await user.type(screen.getByLabelText('Desde'), '2026-11-12')
    await user.type(screen.getByLabelText('Hasta'), '2026-11-10')
    expect(screen.getByText(/no puede ser posterior/)).toBeInTheDocument()
  })

  it('goes to the next page', async () => {
    const many = Array.from({ length: 30 }, (_, index) => sale(30 - index))
    fakeSalesApi(many)
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)

    await screen.findByText('1–25 de 30')
    await user.click(screen.getByRole('button', { name: /Siguiente/ }))

    expect(await screen.findByText('26–30 de 30')).toBeInTheDocument()
  })

  it('shows the error and retries', async () => {
    const api = fakeSalesApi([sale(1)])
    api.failures.push({ status: 500, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)

    expect(await screen.findByRole('heading', { name: 'No se pudieron cargar las ventas' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByRole('table', { name: 'Ventas' })).toBeInTheDocument()
  })

  it('opens the detail with lines, discounts, payments and change', async () => {
    fakeSalesApi([
      sale(7, {
        seller: { id: 'se1', name: 'Ana', code: 'A' },
        lines: [
          {
            itemId: 'it1',
            code: 'A1',
            name: 'Yerba 1 kg',
            unit: 'unit',
            quantity: 2000,
            unitPrice: 500000,
            discount: { type: 'percent', value: 1000 },
            discountAmount: 100000,
            cartDiscountAmount: 50000,
            lineTotal: 850000,
          },
        ],
        subtotal: 1000000,
        cartDiscount: { type: 'amount', value: 50000 },
        cartDiscountAmount: 50000,
        discountTotal: 150000,
        total: 850000,
        payments: [
          { method: 'cash', amount: 500000, reference: null },
          { method: 'transfer', amount: 400000, reference: 'TRF-9' },
        ],
        change: 50000,
      }),
    ])
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)

    await openSale(user, 7)

    const dialog = within(await screen.findByRole('dialog', { name: 'Venta N.º 7' }))
    expect(dialog.getByText('Yerba 1 kg')).toBeInTheDocument()
    expect(dialog.getByText(/Descuento 10 %/)).toBeInTheDocument()
    expect(dialog.getByText(/Descuento general/)).toBeInTheDocument()
    expect(dialog.getByText('Efectivo')).toBeInTheDocument()
    expect(dialog.getByText(/TRF-9/)).toBeInTheDocument()
    expect(dialog.getByText('Vuelto')).toBeInTheDocument()
  })

  it('shows who voided and why in a voided sale, without the void form', async () => {
    fakeSalesApi([sale(3, { status: 'voided', voidedAt: '2026-11-11T10:00:00Z', voidReason: 'Cobro duplicado' })])
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)

    await openSale(user, 3)

    const dialog = within(await screen.findByRole('dialog', { name: 'Venta N.º 3' }))
    expect(dialog.getByText(/Cobro duplicado/)).toBeInTheDocument()
    expect(dialog.queryByRole('button', { name: 'Anular venta' })).not.toBeInTheDocument()
  })

  it('needs a reason, asks for confirmation and voids the sale', async () => {
    const api = fakeSalesApi([sale(5)])
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)

    await openSale(user, 5)
    const dialog = within(await screen.findByRole('dialog', { name: 'Venta N.º 5' }))
    expect(dialog.getByRole('button', { name: 'Anular venta' })).toBeDisabled()

    await user.type(dialog.getByLabelText('Motivo de la anulación'), '  Cobro duplicado ')
    await user.click(dialog.getByRole('button', { name: 'Anular venta' }))

    const confirm = within(await screen.findByRole('alertdialog', { name: '¿Anular la venta N.º 5?' }))
    expect(api.requests.some((request) => request.method === 'POST')).toBe(false)
    await user.click(confirm.getByRole('button', { name: 'Confirmar anulación' }))

    expect((await screen.findAllByText('La venta N.º 5 quedó anulada y el stock se repuso.'))[0]).toBeInTheDocument()
    expect(api.requests.find((request) => request.method === 'POST')).toEqual({
      method: 'POST',
      path: '/api/v1/sales/sa5/void',
      body: { reason: 'Cobro duplicado' },
    })
    expect(await within(screen.getByRole('dialog', { name: 'Venta N.º 5' })).findByText(/Motivo: Cobro duplicado/)).toBeInTheDocument()
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument())
  })

  it('cancelling the confirmation voids nothing', async () => {
    const api = fakeSalesApi([sale(5)])
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)

    await openSale(user, 5)
    const dialog = within(await screen.findByRole('dialog', { name: 'Venta N.º 5' }))
    await user.type(dialog.getByLabelText('Motivo de la anulación'), 'Error')
    await user.click(dialog.getByRole('button', { name: 'Anular venta' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Cancelar' }))

    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument())
    expect(api.requests.some((request) => request.method === 'POST')).toBe(false)
  })

  it('reports why a sale could not be voided and keeps it completed', async () => {
    const api = fakeSalesApi([sale(5)])
    const user = userEvent.setup()
    renderWithProviders(<SalesPage />)

    await openSale(user, 5)
    const dialog = within(await screen.findByRole('dialog', { name: 'Venta N.º 5' }))
    await user.type(dialog.getByLabelText('Motivo de la anulación'), 'Error')
    await user.click(dialog.getByRole('button', { name: 'Anular venta' }))
    api.failures.push({ status: 409, code: 'sale_already_voided', message: 'La venta ya está anulada' })
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Confirmar anulación' }))

    expect((await screen.findAllByText('No se pudo anular la venta'))[0]).toBeInTheDocument()
    expect(screen.getAllByText('La venta ya está anulada')[0]).toBeInTheDocument()
    expect(within(screen.getByRole('dialog', { name: 'Venta N.º 5' })).getByRole('button', { name: 'Anular venta' })).toBeInTheDocument()
  })
})

// The list renders a table and a card stack (one is hidden by CSS), so the
// buttons are looked up in the table.
async function openSale(user: ReturnType<typeof userEvent.setup>, number: number) {
  const table = await screen.findByRole('table', { name: 'Ventas' })
  await user.click(within(table).getByRole('button', { name: `Ver la venta N.º ${number}` }))
}
