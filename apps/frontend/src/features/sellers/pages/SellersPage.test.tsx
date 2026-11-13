import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import type { Seller } from '../types'
import SellersPage from './SellersPage'

const NOW = '2026-11-11T15:00:00Z'

type Failure = { status: number; code: string; message: string }
type SellerSeed = Omit<Seller, 'id'>

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function seller(overrides: Partial<Seller> & { name: string; code: string }): SellerSeed {
  return { commissionRate: null, createdAt: NOW, updatedAt: NOW, ...overrides }
}

// An in-memory stand-in for /api/v1/sellers that follows the backend's
// rules: unique codes, deletes remove the record, sorted by name,
// searchable by name and code.
function fakeSellerApi(initial: SellerSeed[] = []) {
  let sequence = 0
  const sellers: Seller[] = initial.map((entry) => ({ ...entry, id: `se${++sequence}` }))
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

      if (url.pathname === '/api/v1/sellers' && method === 'GET') {
        const search = (url.searchParams.get('q') ?? '').toLowerCase()
        const page = Number(url.searchParams.get('page') ?? 1)
        const limit = Number(url.searchParams.get('limit') ?? 25)
        const matches = sellers
          .filter((entry) => `${entry.name} ${entry.code}`.toLowerCase().includes(search))
          .sort((a, b) => a.name.localeCompare(b.name))
        return json({ items: matches.slice((page - 1) * limit, page * limit), total: matches.length, page, limit })
      }

      if (url.pathname === '/api/v1/sellers' && method === 'POST') {
        if (sellers.some((entry) => entry.code === body.code)) {
          return json({ code: 'seller_code_in_use', message: 'Ya hay un vendedor con ese código' }, 409)
        }
        const created: Seller = { ...body, id: `se${++sequence}`, createdAt: NOW, updatedAt: NOW }
        sellers.push(created)
        return json(created, 201)
      }

      const id = url.pathname.match(/^\/api\/v1\/sellers\/([^/]+)$/)?.[1]
      const index = sellers.findIndex((entry) => entry.id === id)
      if (index === -1) return json({ code: 'seller_not_found', message: 'El vendedor no existe' }, 404)

      if (method === 'DELETE') {
        sellers.splice(index, 1)
        return new Response(null, { status: 204 })
      }
      if (method === 'PUT') {
        sellers[index] = { ...sellers[index], ...body }
      }
      return json(sellers[index])
    }),
  )

  return {
    sellers,
    requests,
    failNext: (failure: Failure) => failures.push(failure),
    writes: () => requests.filter((request) => request.method !== 'GET'),
  }
}

async function table() {
  return within(await screen.findByRole('table', { name: 'Vendedores' }))
}

function rowTexts(scope: Awaited<ReturnType<typeof table>>) {
  return scope
    .getAllByRole('row')
    .slice(1)
    .map((row: HTMLElement) => within(row).getAllByRole('cell').slice(0, 3).map((cell) => cell.textContent))
}

async function openNewSeller(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'Nuevo vendedor' }))
  return within(await screen.findByRole('dialog', { name: 'Nuevo vendedor' }))
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('SellersPage', () => {
  it('lists sellers with their code and commission, as a table and as cards', async () => {
    fakeSellerApi([
      seller({ name: 'Juan Pérez', code: 'JP', commissionRate: 250 }),
      seller({ name: 'María Gómez', code: 'MG' }),
      seller({ name: 'Lucía Fernández', code: 'LF', commissionRate: 1225 }),
    ])
    renderWithProviders(<SellersPage />)

    expect(rowTexts(await table())).toEqual([
      ['JP', 'Juan Pérez', '2,5 %'],
      ['LF', 'Lucía Fernández', '12,25 %'],
      ['MG', 'María Gómez', '—'],
    ])

    const cards = within(screen.getByRole('list', { name: 'Vendedores' }))
    expect(cards.getAllByRole('listitem')).toHaveLength(3)
    expect(cards.getByText('Comisión 2,5 %')).toBeInTheDocument()
    expect(cards.getByText('Sin comisión')).toBeInTheDocument()
  })

  it('creates a seller, sending the code in upper case and the commission in hundredths', async () => {
    const api = fakeSellerApi()
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)

    expect(await screen.findByRole('heading', { name: 'Todavía no hay vendedores' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Crear el primero' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nuevo vendedor' }))

    await user.type(form.getByLabelText('Nombre'), 'Juan Pérez')
    await user.type(form.getByLabelText('Código'), 'jp')
    expect(form.getByLabelText('Código')).toHaveValue('JP')
    expect(form.getByLabelText('Código')).toHaveAccessibleDescription('Con este código se lo elige en la venta, por ejemplo JP.')
    await user.type(form.getByLabelText(/Comisión/), '2,5')
    await user.click(form.getByRole('button', { name: 'Crear vendedor' }))

    expect(await screen.findByText('Vendedor creado')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      { method: 'POST', path: '/api/v1/sellers', body: { name: 'Juan Pérez', code: 'JP', commissionRate: 250 } },
    ])
    expect(rowTexts(await table())[0]).toEqual(['JP', 'Juan Pérez', '2,5 %'])
  })

  it('saves a seller without commission', async () => {
    const api = fakeSellerApi([seller({ name: 'Ana', code: 'AN' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)
    await table()

    const form = await openNewSeller(user)
    await user.type(form.getByLabelText('Nombre'), 'María Gómez')
    await user.type(form.getByLabelText('Código'), 'mg')
    await user.click(form.getByRole('button', { name: 'Crear vendedor' }))

    await waitFor(() => expect(api.writes()).toHaveLength(1))
    expect(api.writes()[0].body).toEqual({ name: 'María Gómez', code: 'MG', commissionRate: null })
  })

  it('validates every field at once before sending', async () => {
    const api = fakeSellerApi([seller({ name: 'Ana', code: 'AN' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)
    await table()

    const form = await openNewSeller(user)
    await user.type(form.getByLabelText(/Comisión/), '120')
    await user.click(form.getByRole('button', { name: 'Crear vendedor' }))

    expect(form.getByText('El nombre es obligatorio')).toBeInTheDocument()
    expect(form.getByText('El código es obligatorio')).toBeInTheDocument()
    expect(form.getByText('Ingresar un porcentaje de 0 a 100, por ejemplo 2,5')).toBeInTheDocument()
    expect(form.getByLabelText(/Comisión/)).toHaveAttribute('aria-invalid', 'true')
    expect(api.writes()).toEqual([])

    await user.type(form.getByLabelText('Código'), 'J P')
    await user.clear(form.getByLabelText(/Comisión/))
    await user.type(form.getByLabelText(/Comisión/), 'mucho')
    await user.click(form.getByRole('button', { name: 'Crear vendedor' }))

    expect(form.getByText('El código lleva de 1 a 6 letras o números, sin espacios')).toBeInTheDocument()
    expect(form.getByText('Ingresar un porcentaje de 0 a 100, por ejemplo 2,5')).toBeInTheDocument()
    expect(api.writes()).toEqual([])
  })

  it('shows a duplicated code next to the code field', async () => {
    fakeSellerApi([seller({ name: 'Juan Pérez', code: 'JP' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)
    await table()

    const form = await openNewSeller(user)
    await user.type(form.getByLabelText('Nombre'), 'Juana Paz')
    await user.type(form.getByLabelText('Código'), 'jp')
    await user.click(form.getByRole('button', { name: 'Crear vendedor' }))

    expect(await form.findByText('Ya hay un vendedor con ese código')).toBeInTheDocument()
    expect(form.getByLabelText('Código')).toHaveAttribute('aria-invalid', 'true')
  })

  it('shows unexpected save errors at the bottom of the form', async () => {
    const api = fakeSellerApi([seller({ name: 'Ana', code: 'AN' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Ana' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar vendedor' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await form.findByRole('alert')).toHaveTextContent('El almacenamiento no está disponible')
  })

  it('edits a seller with its current values and can clear the commission', async () => {
    const api = fakeSellerApi([seller({ name: 'Juan Pérez', code: 'JP', commissionRate: 250 })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Juan Pérez' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar vendedor' }))
    expect(form.getByLabelText('Código')).toHaveValue('JP')
    expect(form.getByLabelText(/Comisión/)).toHaveValue('2,5')

    await user.clear(form.getByLabelText(/Comisión/))
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('Vendedor actualizado')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      { method: 'PUT', path: '/api/v1/sellers/se1', body: { name: 'Juan Pérez', code: 'JP', commissionRate: null } },
    ])
  })

  it('deletes a seller only after confirming', async () => {
    const api = fakeSellerApi([seller({ name: 'Ana', code: 'AN' }), seller({ name: 'Beto', code: 'BE' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Ana' }))
    const dialog = within(await screen.findByRole('alertdialog'))
    expect(dialog.getByText('¿Eliminar «Ana»?')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('Vendedor eliminado')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/sellers/se1', body: undefined }])
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[1])).toEqual(['Beto']))
  })

  it('offers the same actions on the phone cards', async () => {
    const api = fakeSellerApi([seller({ name: 'Ana', code: 'AN' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)
    await table()
    const cards = () => within(screen.getByRole('list', { name: 'Vendedores' }))

    await user.click(cards().getByRole('button', { name: 'Editar Ana' }))
    expect(await screen.findByRole('dialog', { name: 'Editar vendedor' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: /vendedor/ })).not.toBeInTheDocument())

    await user.click(cards().getByRole('button', { name: 'Eliminar Ana' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    await waitFor(() => expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/sellers/se1', body: undefined }]))
  })

  it('reports a failed deletion as a toast', async () => {
    const api = fakeSellerApi([seller({ name: 'Ana', code: 'AN' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Ana' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect((await screen.findAllByText('No se pudo eliminar el vendedor'))[0]).toBeInTheDocument()
    expect(api.sellers).toHaveLength(1)
  })

  it('goes back a page after deleting the last seller of the page', async () => {
    fakeSellerApi(Array.from({ length: 26 }, (_, index) => seller({ name: `Vendedor ${String(index + 1).padStart(2, '0')}`, code: `V${index}` })))
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Siguiente' }))
    expect(await screen.findByText('26–26 de 26')).toBeInTheDocument()
    await user.click((await table()).getByRole('button', { name: 'Eliminar Vendedor 26' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('1–25 de 25')).toBeInTheDocument()
  })

  it('searches by name or code', async () => {
    const api = fakeSellerApi([seller({ name: 'Juan Pérez', code: 'JP' }), seller({ name: 'María Gómez', code: 'MG' })])
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)
    await table()

    expect(screen.getByRole('searchbox', { name: 'Buscar vendedores' })).toHaveAccessibleDescription('Por nombre o código')
    await user.type(screen.getByRole('searchbox', { name: 'Buscar vendedores' }), 'mg')
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[1])).toEqual(['María Gómez']))
    expect(api.requests.at(-1)?.path).toContain('q=mg')

    await user.type(screen.getByRole('searchbox', { name: 'Buscar vendedores' }), 'zzz')
    expect(await screen.findByRole('heading', { name: 'No se encontraron vendedores' })).toBeInTheDocument()
  })

  it('offers to retry when the list cannot be loaded', async () => {
    const api = fakeSellerApi([seller({ name: 'Ana', code: 'AN' })])
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    const user = userEvent.setup()
    renderWithProviders(<SellersPage />)

    expect(await screen.findByRole('heading', { name: 'No se pudieron cargar los vendedores' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(rowTexts(await table()).map((row) => row[1])).toEqual(['Ana'])
  })
})
