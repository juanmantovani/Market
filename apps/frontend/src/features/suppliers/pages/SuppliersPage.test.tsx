import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import type { Supplier } from '../types'
import SuppliersPage from './SuppliersPage'

const NOW = '2026-11-11T15:00:00Z'

type Failure = { status: number; code: string; message: string }
type SupplierSeed = Omit<Supplier, 'id'>

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function supplier(overrides: Partial<Supplier> & { businessName: string }): SupplierSeed {
  return {
    cuit: null,
    contactName: null,
    phone: null,
    email: null,
    address: null,
    notes: null,
    createdAt: NOW,
    updatedAt: NOW,
    ...overrides,
  }
}

// An in-memory stand-in for /api/v1/suppliers that follows the backend's
// rules: one supplier per CUIT, deletes remove the record, sorted by
// business name, searchable by business name, contact, CUIT, email and phone.
function fakeSupplierApi(initial: SupplierSeed[] = []) {
  let sequence = 0
  const suppliers: Supplier[] = initial.map((entry) => ({ ...entry, id: `su${++sequence}` }))
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

      if (url.pathname === '/api/v1/suppliers' && method === 'GET') {
        const search = (url.searchParams.get('q') ?? '').toLowerCase()
        const digits = search.replace(/\D/g, '')
        const page = Number(url.searchParams.get('page') ?? 1)
        const limit = Number(url.searchParams.get('limit') ?? 25)
        const matches = suppliers
          .filter(
            (entry) =>
              `${entry.businessName} ${entry.contactName ?? ''} ${entry.email ?? ''} ${entry.phone ?? ''}`
                .toLowerCase()
                .includes(search) ||
              (digits !== '' && (entry.cuit ?? '').includes(digits)),
          )
          .sort((a, b) => a.businessName.localeCompare(b.businessName))
        return json({ items: matches.slice((page - 1) * limit, page * limit), total: matches.length, page, limit })
      }

      if (url.pathname === '/api/v1/suppliers' && method === 'POST') {
        if (body.cuit && suppliers.some((entry) => entry.cuit === body.cuit)) {
          return json({ code: 'supplier_cuit_in_use', message: 'Ya hay un proveedor con ese CUIT' }, 409)
        }
        const created: Supplier = { ...body, id: `su${++sequence}`, createdAt: NOW, updatedAt: NOW }
        suppliers.push(created)
        return json(created, 201)
      }

      const id = url.pathname.match(/^\/api\/v1\/suppliers\/([^/]+)$/)?.[1]
      const index = suppliers.findIndex((entry) => entry.id === id)
      if (index === -1) return json({ code: 'supplier_not_found', message: 'El proveedor no existe' }, 404)

      if (method === 'DELETE') {
        suppliers.splice(index, 1)
        return new Response(null, { status: 204 })
      }
      if (method === 'PUT') {
        suppliers[index] = { ...suppliers[index], ...body }
      }
      return json(suppliers[index])
    }),
  )

  return {
    suppliers,
    requests,
    failNext: (failure: Failure) => failures.push(failure),
    writes: () => requests.filter((request) => request.method !== 'GET'),
  }
}

async function table() {
  return within(await screen.findByRole('table', { name: 'Proveedores' }))
}

function rowTexts(scope: Awaited<ReturnType<typeof table>>) {
  return scope
    .getAllByRole('row')
    .slice(1)
    .map((row: HTMLElement) => within(row).getAllByRole('cell').slice(0, 4).map((cell) => cell.textContent))
}

async function openNewSupplier(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'Nuevo proveedor' }))
  return within(await screen.findByRole('dialog', { name: 'Nuevo proveedor' }))
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('SuppliersPage', () => {
  it('lists suppliers with their CUIT, contact, phone and email, as a table and as cards', async () => {
    fakeSupplierApi([
      supplier({
        businessName: 'Distribuidora del Centro SA',
        cuit: '30701234561',
        contactName: 'Laura Díaz',
        phone: '351 555-2000',
        email: 'pedidos@delcentro.com.ar',
      }),
      supplier({ businessName: 'Panificadora La Espiga', phone: '351 555-2200' }),
      supplier({ businessName: 'Mayorista Norte SA', cuit: '30715555559', email: 'ventas@norte.com.ar' }),
    ])
    renderWithProviders(<SuppliersPage />)

    expect(rowTexts(await table())).toEqual([
      ['Distribuidora del Centro SA', '30-70123456-1', 'Laura Díaz', '351 555-2000, pedidos@delcentro.com.ar'],
      ['Mayorista Norte SA', '30-71555555-9', '—', 'ventas@norte.com.ar'],
      ['Panificadora La Espiga', '—', '—', '351 555-2200'],
    ])

    const cards = within(screen.getByRole('list', { name: 'Proveedores' }))
    expect(cards.getAllByRole('listitem')).toHaveLength(3)
    expect(cards.getByText('CUIT 30-70123456-1 · Laura Díaz')).toBeInTheDocument()
    expect(cards.getByText('CUIT 30-71555555-9')).toBeInTheDocument()
    expect(cards.getByRole('button', { name: 'Eliminar Panificadora La Espiga' })).toBeInTheDocument()
  })

  it('creates a supplier, sending the CUIT as digits and empty fields as null', async () => {
    const api = fakeSupplierApi()
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)

    expect(await screen.findByRole('heading', { name: 'Todavía no hay proveedores' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Crear el primero' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nuevo proveedor' }))

    await user.type(form.getByLabelText('Razón social'), 'Distribuidora del Centro SA')
    await user.type(form.getByLabelText(/CUIT/), '30-70123456-1')
    await user.type(form.getByLabelText(/Persona de contacto/), 'Laura Díaz')
    await user.type(form.getByLabelText(/Email/), ' pedidos@delcentro.com.ar ')
    await user.click(form.getByRole('button', { name: 'Crear proveedor' }))

    expect(await screen.findByText('Proveedor creado')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      {
        method: 'POST',
        path: '/api/v1/suppliers',
        body: {
          businessName: 'Distribuidora del Centro SA',
          cuit: '30701234561',
          contactName: 'Laura Díaz',
          phone: null,
          email: 'pedidos@delcentro.com.ar',
          address: null,
          notes: null,
        },
      },
    ])
    expect(rowTexts(await table())[0].slice(0, 2)).toEqual(['Distribuidora del Centro SA', '30-70123456-1'])
  })

  it('saves a supplier with only its business name', async () => {
    const api = fakeSupplierApi([supplier({ businessName: 'Centro' })])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)
    await table()

    const form = await openNewSupplier(user)
    await user.type(form.getByLabelText('Razón social'), 'Verdulería Don José')
    await user.click(form.getByRole('button', { name: 'Crear proveedor' }))

    await waitFor(() => expect(api.writes()).toHaveLength(1))
    expect(api.writes()[0].body).toMatchObject({ businessName: 'Verdulería Don José', cuit: null, contactName: null })
  })

  it('validates every field at once before sending', async () => {
    const api = fakeSupplierApi([supplier({ businessName: 'Centro' })])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)
    await table()

    const form = await openNewSupplier(user)
    await user.type(form.getByLabelText(/CUIT/), '30-7012345-1')
    await user.type(form.getByLabelText(/Email/), 'pedidos@')
    await user.click(form.getByRole('button', { name: 'Crear proveedor' }))

    expect(form.getByText('La razón social es obligatoria')).toBeInTheDocument()
    expect(form.getByText('El CUIT lleva 11 números, por ejemplo 30-71234567-8')).toBeInTheDocument()
    expect(form.getByText('Ingresar un email válido, por ejemplo nombre@dominio.com')).toBeInTheDocument()
    expect(form.getByLabelText(/CUIT/)).toHaveAttribute('aria-invalid', 'true')
    expect(api.writes()).toEqual([])

    await user.type(form.getByLabelText('Razón social'), 'A')
    expect(form.queryByText('La razón social es obligatoria')).not.toBeInTheDocument()
  })

  it('shows a duplicated CUIT next to the CUIT field', async () => {
    fakeSupplierApi([supplier({ businessName: 'Centro', cuit: '30701234561' })])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)
    await table()

    const form = await openNewSupplier(user)
    await user.type(form.getByLabelText('Razón social'), 'Otro centro')
    await user.type(form.getByLabelText(/CUIT/), '30701234561')
    await user.click(form.getByRole('button', { name: 'Crear proveedor' }))

    expect(await form.findByText('Ya hay un proveedor con ese CUIT')).toBeInTheDocument()
    expect(form.getByLabelText(/CUIT/)).toHaveAttribute('aria-invalid', 'true')
  })

  it('shows unexpected save errors at the bottom of the form', async () => {
    const api = fakeSupplierApi([supplier({ businessName: 'Centro' })])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Centro' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar proveedor' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await form.findByRole('alert')).toHaveTextContent('El almacenamiento no está disponible')
  })

  it('edits a supplier with its current values and can clear optional fields', async () => {
    const api = fakeSupplierApi([
      supplier({
        businessName: 'Lácteos del Sur SRL',
        cuit: '30709876543',
        contactName: 'Martín Ruiz',
        phone: '351 555-2100',
        address: 'Ruta 9 km 3',
        notes: 'Pedidos los lunes',
      }),
    ])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Lácteos del Sur SRL' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar proveedor' }))
    expect(form.getByLabelText(/CUIT/)).toHaveValue('30-70987654-3')
    expect(form.getByLabelText(/Dirección/)).toHaveValue('Ruta 9 km 3')
    expect(form.getByLabelText(/Notas/)).toHaveValue('Pedidos los lunes')

    await user.clear(form.getByLabelText(/CUIT/))
    await user.clear(form.getByLabelText(/Notas/))
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('Proveedor actualizado')).toBeInTheDocument()
    const [write] = api.writes()
    expect(write.method).toBe('PUT')
    expect(write.path).toBe('/api/v1/suppliers/su1')
    expect(write.body).toEqual({
      businessName: 'Lácteos del Sur SRL',
      cuit: null,
      contactName: 'Martín Ruiz',
      phone: '351 555-2100',
      email: null,
      address: 'Ruta 9 km 3',
      notes: null,
    })
  })

  it('deletes a supplier only after confirming', async () => {
    const api = fakeSupplierApi([supplier({ businessName: 'Centro' }), supplier({ businessName: 'Norte' })])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Centro' }))
    const dialog = within(await screen.findByRole('alertdialog'))
    expect(dialog.getByText('¿Eliminar «Centro»?')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('Proveedor eliminado')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/suppliers/su1', body: undefined }])
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[0])).toEqual(['Norte']))
  })

  it('offers the same actions on the phone cards', async () => {
    const api = fakeSupplierApi([supplier({ businessName: 'Centro' })])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)
    await table()
    const cards = () => within(screen.getByRole('list', { name: 'Proveedores' }))

    await user.click(cards().getByRole('button', { name: 'Editar Centro' }))
    expect(await screen.findByRole('dialog', { name: 'Editar proveedor' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: /proveedor/ })).not.toBeInTheDocument())

    await user.click(cards().getByRole('button', { name: 'Eliminar Centro' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    await waitFor(() => expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/suppliers/su1', body: undefined }]))
  })

  it('reports a failed deletion as a toast', async () => {
    const api = fakeSupplierApi([supplier({ businessName: 'Centro' })])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Centro' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect((await screen.findAllByText('No se pudo eliminar el proveedor'))[0]).toBeInTheDocument()
    expect(api.suppliers).toHaveLength(1)
  })

  it('goes back a page after deleting the last supplier of the page', async () => {
    fakeSupplierApi(Array.from({ length: 26 }, (_, index) => supplier({ businessName: `Proveedor ${String(index + 1).padStart(2, '0')}` })))
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Siguiente' }))
    expect(await screen.findByText('26–26 de 26')).toBeInTheDocument()
    await user.click((await table()).getByRole('button', { name: 'Eliminar Proveedor 26' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('1–25 de 25')).toBeInTheDocument()
  })

  it('searches by contact or by CUIT written with dashes', async () => {
    const api = fakeSupplierApi([
      supplier({ businessName: 'Distribuidora del Centro', cuit: '30701234561', contactName: 'Laura Díaz' }),
      supplier({ businessName: 'Lácteos del Sur', cuit: '30709876543' }),
    ])
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)
    await table()

    expect(screen.getByRole('searchbox', { name: 'Buscar proveedores' })).toHaveAccessibleDescription('Por razón social, CUIT o contacto')
    await user.type(screen.getByRole('searchbox', { name: 'Buscar proveedores' }), '30-70987')
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[0])).toEqual(['Lácteos del Sur']))
    expect(api.requests.at(-1)?.path).toContain('q=30-70987')

    await user.clear(screen.getByRole('searchbox', { name: 'Buscar proveedores' }))
    await user.type(screen.getByRole('searchbox', { name: 'Buscar proveedores' }), 'laura')
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[0])).toEqual(['Distribuidora del Centro']))

    await user.type(screen.getByRole('searchbox', { name: 'Buscar proveedores' }), 'zzz')
    expect(await screen.findByRole('heading', { name: 'No se encontraron proveedores' })).toBeInTheDocument()
  })

  it('offers to retry when the list cannot be loaded', async () => {
    const api = fakeSupplierApi([supplier({ businessName: 'Centro' })])
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    const user = userEvent.setup()
    renderWithProviders(<SuppliersPage />)

    expect(await screen.findByRole('heading', { name: 'No se pudieron cargar los proveedores' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(rowTexts(await table()).map((row) => row[0])).toEqual(['Centro'])
  })
})
