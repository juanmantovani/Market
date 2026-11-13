import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatMoney } from '../../../shared/lib/money'
import { renderWithProviders } from '../../../test/render'
import type { CatalogItem, CreateSaleRequest, CustomerOption, SellerOption } from '../types'
import PosPage from './PosPage'

const ITEMS: CatalogItem[] = [
  { id: 'coca', code: 'BEB-001', barcode: null, name: 'Gaseosa cola', unit: 'unit', salePrice: 320_000, categoryId: 'c1' },
]

const money = (cents: number) => formatMoney(cents).replace(/\s/g, ' ')

type Failure = { status: number; code: string; message: string }

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

// Stands in for POST /api/v1/sales: records what it receives and answers
// like the backend does for a valid sale.
function fakeSalesApi() {
  const requests: Array<{ body: CreateSaleRequest; key: string | null }> = []
  const failures: Failure[] = []

  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: string, init: RequestInit = {}) => {
      if (new URL(input, 'http://localhost').pathname !== '/api/v1/sales' || init.method !== 'POST') {
        return json({ code: 'not_found', message: 'No existe' }, 404)
      }
      const body = JSON.parse(String(init.body)) as CreateSaleRequest
      requests.push({ body, key: new Headers(init.headers).get('Idempotency-Key') })

      const failure = failures.shift()
      if (failure) return json({ code: failure.code, message: failure.message }, failure.status)

      const paid = body.payments.reduce((sum, payment) => sum + payment.amount, 0)
      return json(
        {
          id: 's1',
          number: 12,
          total: body.expectedTotal,
          change: paid - body.expectedTotal,
          payments: body.payments,
          createdAt: '2026-11-13T15:00:00Z',
        },
        201,
      )
    }),
  )

  return { requests, failures }
}

type User = ReturnType<typeof userEvent.setup>

const CUSTOMERS: CustomerOption[] = [
  { id: 'ana', name: 'Ana Pérez', documentType: 'dni', documentNumber: '28456789', system: false },
  { id: 'final', name: 'Consumidor Final', documentType: 'none', documentNumber: '', system: true },
]

const SELLERS: SellerOption[] = [{ id: 'jp', name: 'Juan Pérez', code: 'JP' }]

function renderPos(): User {
  const user = userEvent.setup()
  const searchCatalog = vi.fn(async () => ({ items: ITEMS, total: ITEMS.length, page: 1, limit: 24 }))
  const searchCustomers = vi.fn(async () => ({ items: CUSTOMERS, total: CUSTOMERS.length, page: 1, limit: 20 }))
  const searchSellers = vi.fn(async () => ({ items: SELLERS, total: SELLERS.length, page: 1, limit: 20 }))
  renderWithProviders(
    <PosPage categories={[]} searchCatalog={searchCatalog} searchCustomers={searchCustomers} searchSellers={searchSellers} />,
  )
  return user
}

async function addCola(user: User) {
  const list = screen.getByRole('list', { name: 'Artículos' })
  await user.click(await within(list).findByRole('button', { name: /Gaseosa cola/ }))
}

const checkout = () => within(screen.getByRole('dialog', { name: 'Cobrar' }))

beforeEach(() => {
  window.localStorage.clear()
})

afterEach(() => vi.unstubAllGlobals())

describe('PosPage checkout', () => {
  it('collects a sale with the keyboard, shows the change and starts a new sale', async () => {
    const api = fakeSalesApi()
    const user = renderPos()
    await addCola(user)

    await user.keyboard('{F9}')
    const dialog = await screen.findByRole('dialog', { name: 'Cobrar' })
    expect(within(dialog).getByText(money(320_000))).toBeInTheDocument()
    const amount = within(dialog).getByRole('textbox', { name: 'Importe' })
    expect(amount).toHaveFocus()
    expect(amount).toHaveValue('3200')

    await user.clear(amount)
    await user.type(amount, '5000')
    expect(within(dialog).getByText('Vuelto')).toBeInTheDocument()
    expect(within(dialog).getByText(money(180_000))).toBeInTheDocument()
    await user.keyboard('{Enter}')

    expect(await screen.findByRole('dialog', { name: 'Venta registrada' })).toBeInTheDocument()
    expect(screen.getByText('Venta N.º 12')).toBeInTheDocument()
    expect(screen.getByText(money(180_000))).toBeInTheDocument()
    expect(api.requests).toHaveLength(1)
    expect(api.requests[0].body).toEqual({
      customerId: null,
      sellerId: null,
      lines: [{ itemId: 'coca', quantity: 1000 }],
      payments: [{ method: 'cash', amount: 500_000, reference: null }],
      expectedTotal: 320_000,
    })
    expect(api.requests[0].key).toBeTruthy()
    expect(screen.getByText('Todavía no hay artículos en la venta')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Nueva venta' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  })

  it('sells to the chosen customer and seller, then starts the next sale for Consumidor Final with the same seller', async () => {
    const api = fakeSalesApi()
    const user = renderPos()
    await addCola(user)
    const sale = () => within(screen.getByRole('region', { name: 'Venta actual' }))

    await user.keyboard('{F6}')
    await user.click(await within(await screen.findByRole('dialog', { name: 'Elegir cliente' })).findByRole('button', { name: /Ana Pérez/ }))
    await user.keyboard('{F4}')
    await user.type(await screen.findByRole('searchbox', { name: 'Buscar vendedor' }), 'jp{Enter}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())

    await user.keyboard('{F9}')
    await screen.findByRole('dialog', { name: 'Cobrar' })
    await user.keyboard('{F4}{F6}')
    expect(screen.queryByRole('dialog', { name: /Elegir/ })).not.toBeInTheDocument()
    await user.click(checkout().getByRole('button', { name: 'Confirmar venta' }))

    await screen.findByRole('dialog', { name: 'Venta registrada' })
    expect(api.requests[0].body).toMatchObject({ customerId: 'ana', sellerId: 'jp' })
    await user.click(screen.getByRole('button', { name: 'Nueva venta' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(sale().getByRole('button', { name: /^Cliente: Consumidor Final/ })).toBeInTheDocument()
    expect(sale().getByRole('button', { name: /^Vendedor: Juan Pérez/ })).toBeInTheDocument()
  })

  it('splits the sale between two payment methods', async () => {
    const api = fakeSalesApi()
    const user = renderPos()
    await addCola(user)
    await user.click(screen.getAllByRole('button', { name: /^Cobrar/ })[0])

    const dialog = await screen.findByRole('dialog', { name: 'Cobrar' })
    const first = within(dialog).getByRole('textbox', { name: 'Importe' })
    await user.clear(first)
    await user.type(first, '2000')
    expect(within(dialog).getByText('Falta cobrar')).toBeInTheDocument()
    expect(within(dialog).getByRole('button', { name: 'Confirmar venta' })).toBeDisabled()

    await user.click(within(dialog).getByRole('button', { name: 'Agregar otro medio de pago' }))
    expect(within(dialog).getByRole('textbox', { name: 'Importe 2' })).toHaveValue('1200')
    await user.type(within(dialog).getByRole('textbox', { name: 'Referencia 2 (opcional)' }), 'OP-9')
    await user.click(within(dialog).getByRole('button', { name: 'Confirmar venta' }))

    await screen.findByRole('dialog', { name: 'Venta registrada' })
    expect(api.requests[0].body.payments).toEqual([
      { method: 'cash', amount: 200_000, reference: null },
      { method: 'debit_card', amount: 120_000, reference: 'OP-9' },
    ])
  })

  it('fills the amount with the quick buttons', async () => {
    fakeSalesApi()
    const user = renderPos()
    await addCola(user)
    await user.keyboard('{F9}')
    await screen.findByRole('dialog', { name: 'Cobrar' })

    const amount = checkout().getByRole('textbox', { name: 'Importe' })
    await user.clear(amount)
    await user.click(checkout().getByRole('button', { name: /10\.000/ }))
    expect(amount).toHaveValue('10000')
    await user.click(checkout().getByRole('button', { name: 'Exacto' }))
    expect(amount).toHaveValue('3200')
  })

  it('refuses non-cash payments above the total and invalid amounts', async () => {
    fakeSalesApi()
    const user = renderPos()
    await addCola(user)
    await user.keyboard('{F9}')
    await screen.findByRole('dialog', { name: 'Cobrar' })

    await user.selectOptions(checkout().getByRole('combobox', { name: 'Medio de pago' }), 'qr')
    const amount = checkout().getByRole('textbox', { name: 'Importe' })
    await user.clear(amount)
    await user.type(amount, '4000')
    expect(checkout().getByText('Los pagos que no son en efectivo no pueden superar el total.')).toBeInTheDocument()
    expect(checkout().getByRole('button', { name: 'Confirmar venta' })).toBeDisabled()

    await user.clear(amount)
    await user.type(amount, 'abc')
    expect(checkout().getByText(/Ingresar un importe mayor a cero/)).toBeInTheDocument()
  })

  it('shows the backend error, keeps the key while the same sale is retried and changes it when the sale changes', async () => {
    const api = fakeSalesApi()
    const user = renderPos()
    await addCola(user)
    await user.keyboard('{F9}')
    await screen.findByRole('dialog', { name: 'Cobrar' })

    api.failures.push({ status: 409, code: 'insufficient_stock', message: 'No hay stock suficiente de Gaseosa cola' })
    await user.click(checkout().getByRole('button', { name: 'Confirmar venta' }))
    expect(await checkout().findByRole('alert')).toHaveTextContent('No hay stock suficiente de Gaseosa cola')

    api.failures.push({ status: 500, code: 'storage_unavailable', message: 'x' })
    await user.click(checkout().getByRole('button', { name: 'Confirmar venta' }))
    await waitFor(() => expect(api.requests).toHaveLength(2))
    expect(api.requests[1].key).toBe(api.requests[0].key)

    const amount = checkout().getByRole('textbox', { name: 'Importe' })
    await user.clear(amount)
    await user.type(amount, '4000')
    await user.click(checkout().getByRole('button', { name: 'Confirmar venta' }))
    await screen.findByRole('dialog', { name: 'Venta registrada' })
    expect(api.requests[2].key).not.toBe(api.requests[0].key)
  })

  it('does not open without items and cancels without touching the sale', async () => {
    fakeSalesApi()
    const user = renderPos()
    await within(screen.getByRole('list', { name: 'Artículos' })).findByRole('button', { name: /Gaseosa cola/ })

    await user.keyboard('{F9}')
    expect(screen.getByRole('status')).toHaveTextContent('Agregar al menos un artículo antes de cobrar.')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: /^Cobrar/ })[0]).toBeDisabled()

    await addCola(user)
    await user.keyboard('{F9}')
    await screen.findByRole('dialog', { name: 'Cobrar' })
    await user.click(checkout().getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(within(screen.getByRole('region', { name: 'Venta actual' })).getByText('Gaseosa cola')).toBeInTheDocument()
  })
})
