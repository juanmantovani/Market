import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { useQuery } from '@tanstack/react-query'
import { listStockLevels } from '../api/stock'
import type { StockLevel, StockMovement, SupplierOption } from '../types'
import { ReceiveStock } from './ReceiveStock'
import { StockItemActions } from './StockItemActions'

const NOW = '2026-11-11T15:30:00Z'
const SUPPLIERS: SupplierOption[] = [
  { id: 's1', name: 'Distribuidora del Centro' },
  { id: 's2', name: 'Mayorista Norte' },
]

type Failure = { status: number; code: string; message: string }
type FakeItem = Omit<StockLevel, 'lowStock' | 'updatedAt'>

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function stockItem(overrides: Partial<FakeItem> & { itemId: string; name: string }): FakeItem {
  return { code: overrides.itemId.toUpperCase(), categoryId: 'c1', unit: 'unit', minStock: 0, quantity: 0, ...overrides }
}

const DEMO = [
  stockItem({ itemId: 'yer', code: 'ALM-001', name: 'Yerba mate', minStock: 6000, quantity: 4000 }),
  stockItem({ itemId: 'arr', code: 'ALM-002', name: 'Arroz', minStock: 10_000, quantity: 25_000 }),
  stockItem({ itemId: 'har', code: 'ALM-003', name: 'Harina' }),
  stockItem({ itemId: 'que', code: 'FIA-001', name: 'Queso cremoso', categoryId: 'c2', unit: 'kg', minStock: 2500, quantity: 1800 }),
]

// An in-memory stand-in for /api/v1/stock that follows the backend's rules:
// stock only changes through movements, adjustments need a reason and can't
// leave it below zero, an entry needs known suppliers, and low stock is
// minStock > 0 && quantity <= minStock.
function fakeStockApi(initial: FakeItem[] = DEMO, initialMovements?: StockMovement[]) {
  const items = initial.map((entry) => ({ ...entry }))
  // By default every item with stock already has its opening movement.
  const movements: StockMovement[] = initialMovements
    ? [...initialMovements]
    : items
        .filter((entry) => entry.quantity > 0)
        .map((entry) => ({
          id: `open-${entry.itemId}`, itemId: entry.itemId, type: 'initial' as const, quantity: entry.quantity, balanceAfter: entry.quantity,
          reason: null, reference: null, supplierId: null, unitCost: null, actorId: 'dev', createdAt: NOW,
        }))
  const requests: Array<{ method: string; path: string; body: unknown }> = []
  const failures: Failure[] = []
  let sequence = movements.length

  function record(itemId: string, type: StockMovement['type'], quantity: number, extra: Partial<StockMovement> = {}) {
    const item = items.find((entry) => entry.itemId === itemId) as FakeItem
    item.quantity += quantity
    const movement: StockMovement = {
      id: `m${++sequence}`, itemId, type, quantity, balanceAfter: item.quantity, reason: null, reference: null,
      supplierId: null, unitCost: null, actorId: 'dev', createdAt: NOW, ...extra,
    }
    movements.unshift(movement)
    return movement
  }

  const level = (entry: FakeItem): StockLevel => ({
    ...entry,
    lowStock: entry.minStock > 0 && entry.quantity <= entry.minStock,
    updatedAt: movements.some((movement) => movement.itemId === entry.itemId) ? NOW : null,
  })

  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const url = new URL(input, 'http://localhost')
      const method = init.method ?? 'GET'
      const body = init.body ? JSON.parse(String(init.body)) : undefined
      requests.push({ method, path: `${url.pathname}${url.search}`, body })

      const failure = failures.shift()
      if (failure) return json({ code: failure.code, message: failure.message }, failure.status)

      const page = Number(url.searchParams.get('page') ?? 1)
      const limit = Number(url.searchParams.get('limit') ?? 25)

      if (url.pathname === '/api/v1/stock' && method === 'GET') {
        const search = (url.searchParams.get('q') ?? '').toLowerCase()
        const categoryId = url.searchParams.get('categoryId') ?? ''
        const lowStock = url.searchParams.get('lowStock') === 'true'
        const matches = items
          .map(level)
          .filter((entry) => !categoryId || entry.categoryId === categoryId)
          .filter((entry) => !lowStock || entry.lowStock)
          .filter((entry) => `${entry.name} ${entry.code}`.toLowerCase().includes(search))
          .sort((a, b) => a.name.localeCompare(b.name))
        return json({ items: matches.slice((page - 1) * limit, page * limit), total: matches.length, page, limit })
      }

      if (url.pathname === '/api/v1/stock/movements' && method === 'GET') {
        const matches = movements.filter((movement) => movement.itemId === url.searchParams.get('itemId'))
        return json({ items: matches.slice((page - 1) * limit, page * limit), total: matches.length, page, limit })
      }

      if (url.pathname === '/api/v1/stock/adjustments' && method === 'POST') {
        const item = items.find((entry) => entry.itemId === body.itemId) as FakeItem
        if (!body.reason?.trim()) return json({ code: 'stock_reason_required', message: 'Ingresar el motivo del ajuste' }, 400)
        if (body.quantity < 0 && item.quantity + body.quantity < 0) {
          return json({ code: 'insufficient_stock', message: `No hay stock suficiente de ${item.name}` }, 409)
        }
        return json(record(item.itemId, 'adjustment', body.quantity, { reason: body.reason }), 201)
      }

      if (url.pathname === '/api/v1/stock/initial' && method === 'POST') {
        if (movements.some((movement) => movement.itemId === body.itemId)) {
          return json({ code: 'stock_already_initialized', message: 'El artículo ya tiene movimientos de stock' }, 409)
        }
        return json(record(body.itemId, 'initial', body.quantity), 201)
      }

      if (url.pathname === '/api/v1/stock/entries' && method === 'POST') {
        if (body.supplierId && !SUPPLIERS.some((supplier) => supplier.id === body.supplierId)) {
          return json({ code: 'stock_supplier_not_found', message: 'El proveedor elegido no existe' }, 400)
        }
        const recorded = body.lines.map((line: { itemId: string; quantity: number; unitCost: number | null }) =>
          record(line.itemId, 'entry', line.quantity, { supplierId: body.supplierId, reference: body.reference, unitCost: line.unitCost }),
        )
        return json({ movements: recorded }, 201)
      }

      return json({ code: 'route_not_found', message: 'No existe' }, 404)
    }),
  )

  return {
    items,
    movements,
    requests,
    failNext: (failure: Failure) => failures.push(failure),
    writes: () => requests.filter((request) => request.method !== 'GET'),
  }
}

// Artículos draws these actions next to each item; this stand-in lists the
// items of the fake API with them.
function Harness() {
  const levels = useQuery({
    queryKey: ['stock', 'levels', 'harness'],
    queryFn: ({ signal }) => listStockLevels({ search: '', page: 1, limit: 25 }, signal),
  })
  const names = new Map(SUPPLIERS.map((supplier) => [supplier.id, supplier.name]))

  return (
    <>
      <ReceiveStock suppliers={SUPPLIERS} />
      <ul aria-label="Artículos">
        {levels.data?.items.map((level) => (
          <li key={level.itemId} aria-label={level.name}>
            <span>{level.name}</span>
            <output>{formatUnitQuantity(level.quantity, level.unit)}</output>
            <StockItemActions
              item={{ id: level.itemId, code: level.code, name: level.name, categoryId: level.categoryId, unit: level.unit, minStock: level.minStock }}
              quantity={level.quantity}
              hasMovements={level.updatedAt !== null}
              supplierName={(id) => names.get(id) ?? 'Proveedor eliminado'}
              layout="icons"
            />
          </li>
        ))}
      </ul>
    </>
  )
}

function renderPage() {
  return renderWithProviders(<Harness />)
}

async function items() {
  await screen.findByText('Yerba mate')
  return within(screen.getByRole('list', { name: 'Artículos' }))
}

async function loaded() {
  await screen.findByText('Yerba mate')
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('Stock actions', () => {
  it('adjusts the stock of an item with a reason and shows what is left', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()

    await user.click((await items()).getByRole('button', { name: 'Ajustar stock de Yerba mate' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Ajustar stock' }))
    expect(dialog.getByText('4 u.')).toBeInTheDocument()

    await user.click(dialog.getByRole('button', { name: 'Guardar ajuste' }))
    expect(dialog.getByText('Ingresar una cantidad entera mayor a cero')).toBeInTheDocument()
    expect(dialog.getByText('Ingresar el motivo del ajuste')).toBeInTheDocument()

    await user.type(dialog.getByRole('textbox', { name: /Cantidad/ }), '5')
    expect(dialog.getByText('Queda en')).toHaveTextContent('Queda en -1 u.')
    await user.type(dialog.getByRole('combobox', { name: 'Motivo' }), 'Rotura')
    await user.click(dialog.getByRole('button', { name: 'Guardar ajuste' }))
    expect(dialog.getByText('No hay stock suficiente: hay 4 u.')).toBeInTheDocument()
    expect(api.writes()).toHaveLength(0)

    await user.clear(dialog.getByRole('textbox', { name: /Cantidad/ }))
    await user.type(dialog.getByRole('textbox', { name: /Cantidad/ }), '1,5')
    await user.click(dialog.getByRole('button', { name: 'Guardar ajuste' }))
    expect(dialog.getByText('Ingresar una cantidad entera mayor a cero')).toBeInTheDocument()

    await user.clear(dialog.getByRole('textbox', { name: /Cantidad/ }))
    await user.type(dialog.getByRole('textbox', { name: /Cantidad/ }), '2')
    expect(dialog.getByText('Queda en')).toHaveTextContent('Queda en 2 u.')
    await user.click(dialog.getByRole('button', { name: 'Guardar ajuste' }))

    expect(await screen.findByText('Stock ajustado')).toBeInTheDocument()
    expect(screen.getByText('Ahora hay 2 u. de «Yerba mate».')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      { method: 'POST', path: '/api/v1/stock/adjustments', body: { itemId: 'yer', quantity: -2000, reason: 'Rotura' } },
    ])
    await waitFor(() => expect(within(screen.getByRole('listitem', { name: 'Yerba mate' })).getByRole('status')).toHaveTextContent('2 u.'))
  })

  it('opens the stock of an item without movements with its first count', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByRole('button', { name: 'Ajustar stock de Arroz' })).toBeInTheDocument()
    await user.click((await items()).getByRole('button', { name: 'Cargar stock inicial de Harina' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Cargar stock inicial' }))

    await user.click(dialog.getByRole('button', { name: 'Cargar stock inicial' }))
    expect(dialog.getByText('Ingresar una cantidad entera mayor a cero')).toBeInTheDocument()
    expect(api.writes()).toHaveLength(0)

    await user.type(dialog.getByRole('textbox', { name: /Cantidad/ }), '12')
    await user.click(dialog.getByRole('button', { name: 'Cargar stock inicial' }))

    expect(await screen.findByText('Stock inicial cargado')).toBeInTheDocument()
    expect(screen.getByText('Ahora hay 12 u. de «Harina».')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'POST', path: '/api/v1/stock/initial', body: { itemId: 'har', quantity: 12_000 } }])
    expect(await screen.findByRole('button', { name: 'Ajustar stock de Harina' })).toBeInTheDocument()
  })

  it('shows the backend error when the stock was already opened elsewhere', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()

    await user.click((await items()).getByRole('button', { name: 'Cargar stock inicial de Harina' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Cargar stock inicial' }))
    await user.type(dialog.getByRole('textbox', { name: /Cantidad/ }), '3')
    api.failNext({ status: 409, code: 'stock_already_initialized', message: 'El artículo ya tiene movimientos de stock' })
    await user.click(dialog.getByRole('button', { name: 'Cargar stock inicial' }))
    expect(await dialog.findByRole('alert')).toHaveTextContent('El artículo ya tiene movimientos de stock')
  })

  it('adds stock with a positive adjustment and accepts fractions for items sold by weight', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()

    await user.click((await items()).getByRole('button', { name: 'Ajustar stock de Queso cremoso' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Ajustar stock' }))
    await user.click(dialog.getByRole('radio', { name: 'Sumar' }))
    await user.type(dialog.getByRole('textbox', { name: /Cantidad/ }), '0,25')
    await user.type(dialog.getByRole('combobox', { name: 'Motivo' }), 'Conteo')
    await user.click(dialog.getByRole('button', { name: 'Guardar ajuste' }))

    expect(await screen.findByText('Ahora hay 2,05 kg de «Queso cremoso».')).toBeInTheDocument()
    expect(api.writes()[0].body).toEqual({ itemId: 'que', quantity: 250, reason: 'Conteo' })
  })

  it('shows backend errors of an adjustment on its field or for the whole form', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()

    await user.click((await items()).getByRole('button', { name: 'Ajustar stock de Arroz' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Ajustar stock' }))
    await user.type(dialog.getByRole('textbox', { name: /Cantidad/ }), '3')
    await user.type(dialog.getByRole('combobox', { name: 'Motivo' }), 'Vencimiento')

    api.failNext({ status: 409, code: 'insufficient_stock', message: 'No hay stock suficiente de Arroz' })
    await user.click(dialog.getByRole('button', { name: 'Guardar ajuste' }))
    expect(await dialog.findByText('No hay stock suficiente de Arroz')).toBeInTheDocument()
    expect(dialog.getByRole('textbox', { name: /Cantidad/ })).toHaveAttribute('aria-invalid', 'true')

    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(dialog.getByRole('button', { name: 'Guardar ajuste' }))
    expect(await dialog.findByRole('alert')).toHaveTextContent('El almacenamiento no está disponible')

    await user.click(dialog.getByRole('button', { name: 'Cancelar' }))
    await user.click(within(await screen.findByRole('alertdialog', { name: '¿Descartar los cambios?' })).getByRole('button', { name: 'Descartar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Ajustar stock' })).not.toBeInTheDocument())
  })

  it('asks before discarding a stock entry in progress', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()
    await loaded()

    await user.click(screen.getByRole('button', { name: 'Ingresar mercadería' }))
    await screen.findByRole('dialog', { name: 'Ingresar mercadería' })
    await user.keyboard('{Escape}')
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Ingresar mercadería' })).not.toBeInTheDocument())
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Ingresar mercadería' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Ingresar mercadería' }))
    await user.type(dialog.getByLabelText(/Comprobante/), 'R-0001')
    await user.keyboard('{Escape}')
    await user.click(within(await screen.findByRole('alertdialog', { name: '¿Descartar los cambios?' })).getByRole('button', { name: 'Seguir editando' }))
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument())
    expect(dialog.getByLabelText(/Comprobante/)).toHaveValue('R-0001')

    await user.click(dialog.getByRole('button', { name: 'Cancelar' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Descartar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Ingresar mercadería' })).not.toBeInTheDocument())
    expect(api.writes()).toEqual([])
  })

  it('registers a stock entry with several items, found by search or with a scanner', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()
    await loaded()

    await user.click(screen.getByRole('button', { name: 'Ingresar mercadería' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Ingresar mercadería' }))
    expect(dialog.getByRole('searchbox', { name: 'Artículo 1' })).toHaveFocus()
    await user.click(dialog.getByRole('button', { name: 'Registrar ingreso' }))
    expect(dialog.getByText('Elegir el artículo')).toBeInTheDocument()

    await user.type(dialog.getByRole('searchbox', { name: 'Artículo 1' }), 'yer')
    await user.click(await dialog.findByRole('button', { name: /Yerba mate/ }))
    expect(dialog.getByRole('button', { name: 'Cambiar Yerba mate' })).toBeInTheDocument()
    expect(dialog.queryByText('Elegir el artículo')).not.toBeInTheDocument()
    await waitFor(() => expect(dialog.getByRole('textbox', { name: 'Cantidad (u.)' })).toHaveFocus())
    await user.keyboard('12')
    await user.type(dialog.getByRole('textbox', { name: /Costo unitario/ }), '3.200,50')

    await user.click(dialog.getByRole('button', { name: 'Agregar artículo' }))
    expect(dialog.getByRole('searchbox', { name: 'Artículo 2' })).toHaveFocus()
    await user.keyboard('FIA-001{Enter}')
    expect(await dialog.findByRole('button', { name: 'Cambiar Queso cremoso' })).toBeInTheDocument()
    const second = within(dialog.getByRole('listitem', { name: 'Línea 2' }))
    await waitFor(() => expect(second.getByRole('textbox', { name: 'Cantidad (kg)' })).toHaveFocus())
    await user.keyboard('1,5')

    await user.selectOptions(dialog.getByRole('combobox', { name: /Proveedor/ }), 'Mayorista Norte')
    await user.type(dialog.getByRole('textbox', { name: /Comprobante/ }), 'Remito 0001-00004512')
    await user.click(dialog.getByRole('button', { name: 'Registrar ingreso' }))

    expect(await screen.findByText('Ingreso registrado')).toBeInTheDocument()
    expect(screen.getByText('Se actualizó el stock de 2 artículos.')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      {
        method: 'POST',
        path: '/api/v1/stock/entries',
        body: {
          supplierId: 's2',
          reference: 'Remito 0001-00004512',
          lines: [
            { itemId: 'yer', quantity: 12_000, unitCost: 320_050 },
            { itemId: 'que', quantity: 1500, unitCost: null },
          ],
        },
      },
    ])
    await waitFor(() => expect(within(screen.getByRole('listitem', { name: 'Yerba mate' })).getByRole('status')).toHaveTextContent('16 u.'))
  })

  it('checks every line of an entry before sending it', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()
    await loaded()

    await user.click(screen.getByRole('button', { name: 'Ingresar mercadería' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Ingresar mercadería' }))
    await user.type(dialog.getByRole('searchbox', { name: 'Artículo 1' }), 'arroz{Enter}')
    const first = within(dialog.getByRole('listitem', { name: 'Línea 1' }))
    await waitFor(() => expect(first.getByRole('textbox', { name: 'Cantidad (u.)' })).toHaveFocus())
    await user.type(first.getByRole('textbox', { name: /Costo unitario/ }), 'mucho')
    await user.click(dialog.getByRole('button', { name: 'Agregar artículo' }))
    await user.keyboard('arroz{Enter}')
    const second = within(dialog.getByRole('listitem', { name: 'Línea 2' }))
    await waitFor(() => expect(second.getByRole('textbox', { name: 'Cantidad (u.)' })).toHaveFocus())
    await user.click(dialog.getByRole('button', { name: 'Registrar ingreso' }))

    expect(dialog.getAllByText('Ingresar una cantidad entera mayor a cero')).toHaveLength(2)
    expect(first.getByText('Ingresar un costo válido, por ejemplo 800,00')).toBeInTheDocument()
    expect(dialog.getByText('Este artículo ya está en otra línea')).toBeInTheDocument()

    await user.click(dialog.getByRole('button', { name: 'Quitar la línea 2' }))
    expect(dialog.queryByText('Este artículo ya está en otra línea')).not.toBeInTheDocument()
    expect(dialog.queryByRole('button', { name: 'Quitar la línea 1' })).not.toBeInTheDocument()

    await user.click(within(dialog.getByRole('listitem', { name: 'Línea 1' })).getByRole('button', { name: 'Cambiar Arroz' }))
    await waitFor(() => expect(dialog.getByRole('searchbox', { name: 'Artículo 1' })).toHaveFocus())
    await user.type(dialog.getByRole('searchbox', { name: 'Artículo 1' }), 'azúcar{Enter}')
    expect(await dialog.findAllByText('No se encontraron artículos. Probar con otra búsqueda.')).not.toHaveLength(0)
    expect(api.writes()).toHaveLength(0)
  })

  it('shows backend errors of an entry on its field or for the whole form', async () => {
    const api = fakeStockApi()
    const user = userEvent.setup()
    renderPage()
    await loaded()

    await user.click(screen.getByRole('button', { name: 'Ingresar mercadería' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Ingresar mercadería' }))
    await user.type(dialog.getByRole('searchbox', { name: 'Artículo 1' }), 'arroz{Enter}')
    await waitFor(() => expect(dialog.getByRole('textbox', { name: 'Cantidad (u.)' })).toHaveFocus())
    await user.keyboard('10')

    api.failNext({ status: 400, code: 'stock_supplier_not_found', message: 'El proveedor elegido no existe' })
    await user.click(dialog.getByRole('button', { name: 'Registrar ingreso' }))
    expect(await dialog.findByText('El proveedor elegido no existe')).toBeInTheDocument()
    expect(dialog.getByRole('combobox', { name: /Proveedor/ })).toHaveAttribute('aria-invalid', 'true')

    api.failNext({ status: 400, code: 'stock_quantity_invalid', message: 'La cantidad tiene que ser mayor a cero' })
    await user.click(dialog.getByRole('button', { name: 'Registrar ingreso' }))
    expect(await dialog.findByRole('alert')).toHaveTextContent('La cantidad tiene que ser mayor a cero')

    await user.click(dialog.getByRole('button', { name: 'Registrar ingreso' }))
    expect(await screen.findByText('Se actualizó el stock de 1 artículo.')).toBeInTheDocument()
  })

  it('shows the movements of an item, newest first, with their details', async () => {
    const movement = (overrides: Partial<StockMovement> & { id: string; type: StockMovement['type'] }): StockMovement => ({
      itemId: 'que', quantity: 0, balanceAfter: 0, reason: null, reference: null, supplierId: null, unitCost: null,
      actorId: 'dev', createdAt: NOW, ...overrides,
    })
    fakeStockApi(DEMO, [
      movement({ id: 'm3', type: 'adjustment', quantity: -200, balanceAfter: 1800, reason: 'Rotura' }),
      movement({ id: 'm2', type: 'entry', quantity: 1000, balanceAfter: 2000, supplierId: 's1', reference: 'Remito 12', unitCost: 690_000 }),
      movement({ id: 'm1', type: 'initial', quantity: 1000, balanceAfter: 1000, supplierId: 'gone' }),
    ])
    const user = userEvent.setup()
    renderPage()

    await user.click((await items()).getByRole('button', { name: 'Movimientos de Queso cremoso' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Movimientos' }))
    const list = within(await dialog.findByRole('list', { name: 'Movimientos de Queso cremoso' }))
    const entries = list.getAllByRole('listitem').map((entry) => entry.textContent)

    expect(entries[0]).toMatch(/^Ajuste-0,2 kg.*Queda: 1,8 kgMotivo: Rotura$/)
    expect(entries[1]).toMatch(/^Ingreso\+1 kg.*Queda: 2 kgDistribuidora del Centro · Remito 12 · Costo \$\s?6\.900,00 \/ kg$/)
    expect(entries[2]).toMatch(/^Stock inicial\+1 kg.*Queda: 1 kgProveedor eliminado$/)

    await user.click(dialog.getByRole('button', { name: 'Cerrar' }))
    await user.click((await items()).getByRole('button', { name: 'Movimientos de Arroz' }))
    expect(await screen.findByText('Todavía no hay movimientos')).toBeInTheDocument()
  })

  it('pages through a long history and offers a retry when it cannot be loaded', async () => {
    const many = Array.from({ length: 25 }, (_, index): StockMovement => ({
      id: `m${index}`, itemId: 'arr', type: 'adjustment', quantity: 1000, balanceAfter: (25 - index) * 1000,
      reason: `Conteo ${index}`, reference: null, supplierId: null, unitCost: null, actorId: 'dev', createdAt: NOW,
    }))
    const api = fakeStockApi(DEMO, many)
    const user = userEvent.setup()
    renderPage()
    await loaded()

    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click((await items()).getByRole('button', { name: 'Movimientos de Arroz' }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Movimientos' }))
    expect(await dialog.findByText('No se pudieron cargar los movimientos')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: 'Reintentar' }))

    const list = await dialog.findByRole('list', { name: 'Movimientos de Arroz' })
    expect(within(list).getAllByRole('listitem')).toHaveLength(20)
    expect(dialog.getByText('1–20 de 25')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: /Siguiente/ }))
    await waitFor(() => expect(within(dialog.getByRole('list', { name: 'Movimientos de Arroz' })).getAllByRole('listitem')).toHaveLength(5))
  })
})
