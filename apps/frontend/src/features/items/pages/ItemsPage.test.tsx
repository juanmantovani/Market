import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import type { CategoryOption, Item, SupplierOption } from '../types'
import ItemsPage from './ItemsPage'

const NOW = '2026-11-10T15:00:00Z'
const CATEGORIES: CategoryOption[] = [
  { id: 'c1', name: 'Almacén' },
  { id: 'c2', name: 'Fiambrería' },
]
const SUPPLIERS: SupplierOption[] = [
  { id: 's1', name: 'Distribuidora del Centro' },
  { id: 's2', name: 'Mayorista Norte' },
]

type Failure = { status: number; code: string; message: string }

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function item(overrides: Partial<Item> & { name: string; code: string }): Omit<Item, 'id'> & { id?: string } {
  return {
    barcode: null,
    description: null,
    categoryId: 'c1',
    supplierIds: [],
    unit: 'unit',
    salePrice: 100_000,
    cost: null,
    taxRate: 2100,
    minStock: 0,
    createdAt: NOW,
    updatedAt: NOW,
    ...overrides,
  }
}

// An in-memory stand-in for /api/v1/items that follows the backend's rules:
// unique codes and barcodes, deletes remove the record, sorted by name.
function fakeItemApi(initial: Array<ReturnType<typeof item>> = []) {
  let sequence = 0
  const items: Item[] = initial.map((entry) => ({ ...entry, id: `i${++sequence}` }) as Item)
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

      if (url.pathname === '/api/v1/items' && method === 'GET') {
        const search = (url.searchParams.get('q') ?? '').toLowerCase()
        const categoryId = url.searchParams.get('categoryId') ?? ''
        const page = Number(url.searchParams.get('page') ?? 1)
        const limit = Number(url.searchParams.get('limit') ?? 25)
        const matches = items
          .filter((entry) => !categoryId || entry.categoryId === categoryId)
          .filter((entry) => `${entry.name} ${entry.code} ${entry.barcode ?? ''}`.toLowerCase().includes(search))
          .sort((a, b) => a.name.localeCompare(b.name))
        return json({ items: matches.slice((page - 1) * limit, page * limit), total: matches.length, page, limit })
      }

      if (url.pathname === '/api/v1/items' && method === 'POST') {
        if (items.some((entry) => entry.code.toLowerCase() === body.code.toLowerCase())) {
          return json({ code: 'item_code_in_use', message: 'Ya hay un artículo con ese código' }, 409)
        }
        const created = { ...body, id: `i${++sequence}`, createdAt: NOW, updatedAt: NOW } as Item
        items.push(created)
        return json(created, 201)
      }

      const id = url.pathname.match(/^\/api\/v1\/items\/([^/]+)$/)?.[1]
      const index = items.findIndex((entry) => entry.id === id)
      if (index === -1) return json({ code: 'item_not_found', message: 'El artículo no existe' }, 404)

      if (method === 'DELETE') {
        items.splice(index, 1)
        return new Response(null, { status: 204 })
      }
      if (method === 'PUT') {
        items[index] = { ...items[index], ...body }
      }
      return json(items[index])
    }),
  )

  return {
    items,
    requests,
    failNext: (failure: Failure) => failures.push(failure),
    writes: () => requests.filter((request) => request.method !== 'GET'),
  }
}

function renderPage(categories: CategoryOption[] = CATEGORIES, suppliers: SupplierOption[] = SUPPLIERS) {
  return renderWithProviders(
    <MemoryRouter>
      <ItemsPage categories={categories} categoriesLoading={false} suppliers={suppliers} />
    </MemoryRouter>,
  )
}

async function table() {
  return within(await screen.findByRole('table', { name: 'Artículos' }))
}

function rowTexts(scope: Awaited<ReturnType<typeof table>>) {
  return scope
    .getAllByRole('row')
    .slice(1)
    .map((row: HTMLElement) => within(row).getAllByRole('cell').slice(0, 4).map((cell) => cell.textContent))
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ItemsPage', () => {
  it('lists items with their category and price, as a table and as cards', async () => {
    fakeItemApi([
      item({ name: 'Yerba mate 1 kg', code: 'ALM-001', salePrice: 450_000 }),
      item({ name: 'Jamón cocido', code: 'FIA-002', categoryId: 'c2', unit: 'kg', salePrice: 1_250_000 }),
    ])
    renderPage()

    const rows = rowTexts(await table())
    expect(rows).toHaveLength(2)
    expect(rows[0].slice(0, 3)).toEqual(['FIA-002', 'Jamón cocido', 'Fiambrería'])
    expect(rows[0][3]).toMatch(/12\.500,00 \/ kg$/)
    expect(rows[1].slice(0, 3)).toEqual(['ALM-001', 'Yerba mate 1 kg', 'Almacén'])
    expect(rows[1][3]).toMatch(/4\.500,00$/)

    const cards = within(screen.getByRole('list', { name: 'Artículos' }))
    expect(cards.getAllByRole('listitem')).toHaveLength(2)
    expect(cards.getByRole('button', { name: 'Eliminar Yerba mate 1 kg' })).toBeInTheDocument()
  })

  it('asks to create a category first when there are none', async () => {
    fakeItemApi()
    renderPage([])

    expect(await screen.findByRole('heading', { name: 'Primero, crear una categoría' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ir a categorías' })).toHaveAttribute('href', '/categorias')
    expect(screen.getByRole('button', { name: 'Nuevo artículo' })).toBeDisabled()
  })

  it('creates an item, sending amounts as integers', async () => {
    const api = fakeItemApi()
    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByRole('heading', { name: 'Todavía no hay artículos' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Crear el primero' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nuevo artículo' }))

    await user.type(form.getByLabelText('Nombre'), 'Queso cremoso')
    await user.type(form.getByLabelText('Código'), 'FIA-001')
    await user.type(form.getByLabelText(/Código de barras/), '7790000000011')
    await user.selectOptions(form.getByLabelText('Categoría'), 'Fiambrería')
    await user.selectOptions(form.getByLabelText('Se vende por'), 'Kilogramo')
    await user.type(form.getByLabelText('Precio de venta'), '9.800,50')
    await user.type(form.getByLabelText(/Costo/), '6900')
    await user.selectOptions(form.getByLabelText('IVA'), '10,5 %')
    await user.clear(form.getByLabelText(/Stock mínimo/))
    await user.type(form.getByLabelText(/Stock mínimo/), '2,5')
    expect(form.getByText('(kg)')).toBeInTheDocument()
    await user.click(form.getByRole('button', { name: 'Proveedores (opcional)' }))
    await user.click(form.getByRole('checkbox', { name: 'Mayorista Norte' }))
    await user.click(form.getByRole('checkbox', { name: 'Distribuidora del Centro' }))
    expect(form.getByRole('button', { name: 'Proveedores · 2 elegidos' })).toBeInTheDocument()
    await user.click(form.getByRole('button', { name: 'Crear artículo' }))

    expect(await screen.findByText('Artículo creado')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      {
        method: 'POST',
        path: '/api/v1/items',
        body: {
          code: 'FIA-001',
          barcode: '7790000000011',
          name: 'Queso cremoso',
          description: null,
          categoryId: 'c2',
          supplierIds: ['s2', 's1'],
          unit: 'kg',
          salePrice: 980_050,
          cost: 690_000,
          taxRate: 1050,
          minStock: 2500,
        },
      },
    ])
    expect(rowTexts(await table())[0].slice(0, 3)).toEqual(['FIA-001', 'Queso cremoso', 'Fiambrería'])
  })

  it('validates the form before sending it', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage()
    await table()

    await user.click(screen.getByRole('button', { name: 'Nuevo artículo' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nuevo artículo' }))
    await user.type(form.getByLabelText(/Código de barras/), '779 000')
    await user.type(form.getByLabelText('Precio de venta'), 'gratis')
    await user.type(form.getByLabelText(/Costo/), 'mucho')
    await user.clear(form.getByLabelText(/Stock mínimo/))
    await user.type(form.getByLabelText(/Stock mínimo/), 'dos')
    await user.click(form.getByRole('button', { name: 'Crear artículo' }))

    expect(form.getByText('El nombre es obligatorio')).toBeInTheDocument()
    expect(form.getByText('El código es obligatorio')).toBeInTheDocument()
    expect(form.getByText('El código de barras no puede tener espacios')).toBeInTheDocument()
    expect(form.getByLabelText('Categoría')).toHaveAccessibleDescription('Elegir una categoría')
    expect(form.getByText('Ingresar un precio mayor a cero, por ejemplo 1.250,50')).toBeInTheDocument()
    expect(form.getByText('Ingresar un costo válido, por ejemplo 800,00')).toBeInTheDocument()
    expect(form.getByText('Ingresar una cantidad, por ejemplo 2,5')).toBeInTheDocument()
    expect(form.getByLabelText('Precio de venta')).toHaveAttribute('aria-invalid', 'true')
    expect(api.writes()).toEqual([])

    await user.type(form.getByLabelText('Nombre'), 'A')
    expect(form.queryByText('El nombre es obligatorio')).not.toBeInTheDocument()
  })

  it('asks for a whole minimum stock for items sold by the piece', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage()

    await user.click((await table()).getByRole('button', { name: 'Editar Yerba' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    await user.clear(form.getByLabelText(/Stock mínimo/))
    await user.type(form.getByLabelText(/Stock mínimo/), '1,5')
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(form.getByText('Ingresar una cantidad entera')).toBeInTheDocument()
    expect(api.writes()).toEqual([])

    await user.selectOptions(form.getByLabelText('Se vende por'), 'Kilogramo')
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))
    await waitFor(() => expect(api.writes()).toHaveLength(1))
    expect(api.writes()[0].body).toMatchObject({ unit: 'kg', minStock: 1500 })
  })

  it('shows a duplicated code next to the code field', async () => {
    fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage()
    await table()

    await user.click(screen.getByRole('button', { name: 'Nuevo artículo' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nuevo artículo' }))
    await user.type(form.getByLabelText('Nombre'), 'Otra yerba')
    await user.type(form.getByLabelText('Código'), 'alm-001')
    await user.selectOptions(form.getByLabelText('Categoría'), 'Almacén')
    await user.type(form.getByLabelText('Precio de venta'), '100')
    await user.click(form.getByRole('button', { name: 'Crear artículo' }))

    expect(await form.findByText('Ya hay un artículo con ese código')).toBeInTheDocument()
    expect(form.getByLabelText('Código')).toHaveAttribute('aria-invalid', 'true')
  })

  it('shows unexpected save errors at the bottom of the form', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage()

    await user.click((await table()).getByRole('button', { name: 'Editar Yerba' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await form.findByRole('alert')).toHaveTextContent('El almacenamiento no está disponible')
  })

  it('edits an item with its current values and can clear the cost', async () => {
    const api = fakeItemApi([
      item({ name: 'Yerba', code: 'ALM-001', barcode: '7790000000011', salePrice: 450_000, cost: 310_000, minStock: 6000 }),
    ])
    const user = userEvent.setup()
    renderPage()

    await user.click((await table()).getByRole('button', { name: 'Editar Yerba' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    expect(form.getByLabelText('Precio de venta')).toHaveValue('4500')
    expect(form.getByLabelText(/Costo/)).toHaveValue('3100')
    expect(form.getByLabelText(/Stock mínimo/)).toHaveValue('6')
    expect(form.getByLabelText('Categoría')).toHaveValue('c1')
    expect(form.getByLabelText('IVA')).toHaveValue('2100')

    await user.clear(form.getByLabelText('Precio de venta'))
    await user.type(form.getByLabelText('Precio de venta'), '4999,99')
    await user.clear(form.getByLabelText(/Costo/))
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('Artículo actualizado')).toBeInTheDocument()
    const [write] = api.writes()
    expect(write.method).toBe('PUT')
    expect(write.path).toBe('/api/v1/items/i1')
    expect(write.body).toMatchObject({ salePrice: 499_999, cost: null, barcode: '7790000000011', minStock: 6000 })
  })

  it('deletes an item only after confirming', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' }), item({ name: 'Arroz', code: 'ALM-002' })])
    const user = userEvent.setup()
    renderPage()

    await user.click((await table()).getByRole('button', { name: 'Eliminar Yerba' }))
    const dialog = within(await screen.findByRole('alertdialog'))
    expect(dialog.getByText('¿Eliminar «Yerba»?')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('Artículo eliminado')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/items/i1', body: undefined }])
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[1])).toEqual(['Arroz']))
  })

  it('offers the same actions on the phone cards', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage()
    await table()
    const cards = () => within(screen.getByRole('list', { name: 'Artículos' }))

    await user.click(cards().getByRole('button', { name: 'Editar Yerba' }))
    expect(await screen.findByRole('dialog', { name: 'Editar artículo' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: /artículo/ })).not.toBeInTheDocument())

    await user.click(cards().getByRole('button', { name: 'Eliminar Yerba' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    await waitFor(() => expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/items/i1', body: undefined }]))
  })

  it('reports a failed deletion as a toast', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage()

    await user.click((await table()).getByRole('button', { name: 'Eliminar Yerba' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect((await screen.findAllByText('No se pudo eliminar el artículo'))[0]).toBeInTheDocument()
    expect(api.items).toHaveLength(1)
  })

  it('goes back a page after deleting the last item of the page', async () => {
    fakeItemApi(Array.from({ length: 26 }, (_, index) => item({ name: `Artículo ${String(index + 1).padStart(2, '0')}`, code: `A-${index}` })))
    const user = userEvent.setup()
    renderPage()
    await table()

    await user.click(screen.getByRole('button', { name: 'Siguiente' }))
    expect(await screen.findByText('26–26 de 26')).toBeInTheDocument()
    await user.click((await table()).getByRole('button', { name: 'Eliminar Artículo 26' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('1–25 de 25')).toBeInTheDocument()
  })

  it('filters by category and searches by name or code', async () => {
    const api = fakeItemApi([
      item({ name: 'Yerba', code: 'ALM-001' }),
      item({ name: 'Jamón cocido', code: 'FIA-002', categoryId: 'c2' }),
    ])
    const user = userEvent.setup()
    renderPage()
    await table()

    await user.selectOptions(screen.getByRole('combobox', { name: 'Filtrar por categoría' }), 'Fiambrería')
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[1])).toEqual(['Jamón cocido']))
    expect(api.requests.at(-1)?.path).toContain('categoryId=c2')

    await user.selectOptions(screen.getByRole('combobox', { name: 'Filtrar por categoría' }), 'Todas las categorías')
    await user.type(screen.getByRole('searchbox', { name: 'Buscar artículos' }), 'alm-001')
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[1])).toEqual(['Yerba']))

    await user.type(screen.getByRole('searchbox', { name: 'Buscar artículos' }), 'zzz')
    expect(await screen.findByRole('heading', { name: 'No se encontraron artículos' })).toBeInTheDocument()
  })

  it('offers to retry when the list cannot be loaded', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByRole('heading', { name: 'No se pudieron cargar los artículos' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(rowTexts(await table()).map((row) => row[1])).toEqual(['Yerba'])
  })

  it('keeps the suppliers folded, showing how many are chosen, and opens them from the keyboard', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001', supplierIds: ['s1', 's2'] })])
    const user = userEvent.setup()
    renderPage()

    await user.click((await table()).getByRole('button', { name: 'Editar Yerba' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    const header = form.getByRole('button', { name: 'Proveedores · 2 elegidos' })
    expect(header).toHaveAttribute('aria-expanded', 'false')
    expect(form.queryByRole('checkbox', { name: 'Mayorista Norte' })).not.toBeInTheDocument()

    header.focus()
    await user.keyboard('{Enter}')
    expect(header).toHaveAttribute('aria-expanded', 'true')
    expect(form.getByRole('group', { name: 'Proveedores' })).toBeInTheDocument()
    expect(form.getByRole('checkbox', { name: 'Distribuidora del Centro' })).toBeChecked()
    expect(form.getByRole('checkbox', { name: 'Mayorista Norte' })).toBeChecked()

    await user.click(form.getByRole('checkbox', { name: 'Distribuidora del Centro' }))
    expect(header).toHaveAccessibleName('Proveedores · 1 elegido')
    await user.click(form.getByRole('checkbox', { name: 'Mayorista Norte' }))
    expect(header).toHaveAccessibleName('Proveedores (opcional)')

    header.focus()
    await user.keyboard(' ')
    expect(header).toHaveAttribute('aria-expanded', 'false')
    await waitFor(() => expect(form.queryByRole('group', { name: 'Proveedores' })).not.toBeInTheDocument())

    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))
    await waitFor(() => expect(api.writes()).toHaveLength(1))
    expect(api.writes()[0].body).toMatchObject({ supplierIds: [] })
  })

  it('opens the suppliers by itself to show a supplier that no longer exists', async () => {
    const api = fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage()

    await user.click((await table()).getByRole('button', { name: 'Editar Yerba' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    await user.click(form.getByRole('button', { name: 'Proveedores (opcional)' }))
    await user.click(form.getByRole('checkbox', { name: 'Mayorista Norte' }))
    const header = form.getByRole('button', { name: 'Proveedores · 1 elegido' })
    await user.click(header)
    expect(header).toHaveAttribute('aria-expanded', 'false')

    api.failNext({ status: 400, code: 'item_supplier_not_found', message: 'Uno de los proveedores elegidos no existe' })
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await form.findByText('Uno de los proveedores elegidos no existe')).toBeInTheDocument()
    expect(header).toHaveAttribute('aria-expanded', 'true')
    expect(form.getByRole('group', { name: 'Proveedores' })).toHaveAccessibleDescription('Uno de los proveedores elegidos no existe')

    await user.click(form.getByRole('checkbox', { name: 'Mayorista Norte' }))
    expect(form.queryByText('Uno de los proveedores elegidos no existe')).not.toBeInTheDocument()
  })

  it('says so when there are no suppliers to choose from', async () => {
    fakeItemApi([item({ name: 'Yerba', code: 'ALM-001' })])
    const user = userEvent.setup()
    renderPage(CATEGORIES, [])

    await user.click((await table()).getByRole('button', { name: 'Editar Yerba' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar artículo' }))
    await user.click(form.getByRole('button', { name: 'Proveedores (opcional)' }))
    expect(form.getByText('Todavía no hay proveedores cargados.')).toBeInTheDocument()
    expect(form.queryByRole('checkbox', { name: 'Mayorista Norte' })).not.toBeInTheDocument()
  })

  it('shows an unknown category as a dash', async () => {
    fakeItemApi([item({ name: 'Huérfano', code: 'X-1', categoryId: 'gone' })])
    renderPage()

    expect(rowTexts(await table())[0][2]).toBe('—')
  })
})
