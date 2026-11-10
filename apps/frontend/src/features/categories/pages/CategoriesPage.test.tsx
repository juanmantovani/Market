import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import type { Category } from '../types'
import CategoriesPage from './CategoriesPage'

const NOW = '2026-11-10T15:00:00Z'

type Failure = { status: number; code: string; message: string }

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

// A small in-memory stand-in for /api/v1/categories that follows the same
// rules as the backend: active names are unique, deletes deactivate, lists
// are sorted by name and paginated.
function fakeCategoryApi(initial: Array<Partial<Category> & { name: string }> = []) {
  let sequence = 0
  const categories: Category[] = initial.map((category) => ({
    id: `c${++sequence}`,
    description: null,
    active: true,
    createdAt: NOW,
    updatedAt: NOW,
    ...category,
  }))
  const requests: Array<{ method: string; path: string; body: unknown }> = []
  const failures: Failure[] = []

  const nameTaken = (name: string, exceptId?: string) =>
    categories.some(
      (category) =>
        category.active && category.id !== exceptId && category.name.toLowerCase() === name.toLowerCase(),
    )

  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const url = new URL(input, 'http://localhost')
      const method = init.method ?? 'GET'
      const body = init.body ? JSON.parse(String(init.body)) : undefined
      requests.push({ method, path: url.pathname, body })

      const failure = failures.shift()
      if (failure) return json({ code: failure.code, message: failure.message }, failure.status)

      if (url.pathname === '/api/v1/categories' && method === 'GET') {
        const search = (url.searchParams.get('q') ?? '').toLowerCase()
        const status = url.searchParams.get('status') ?? 'active'
        const page = Number(url.searchParams.get('page') ?? 1)
        const limit = Number(url.searchParams.get('limit') ?? 25)
        const matches = categories
          .filter((category) => status === 'all' || category.active === (status === 'active'))
          .filter((category) => category.name.toLowerCase().includes(search))
          .sort((a, b) => a.name.localeCompare(b.name))
        return json({
          items: matches.slice((page - 1) * limit, page * limit),
          total: matches.length,
          page,
          limit,
        })
      }

      if (url.pathname === '/api/v1/categories' && method === 'POST') {
        if (nameTaken(body.name)) {
          return json({ code: 'category_name_in_use', message: 'Ya hay una categoría activa con ese nombre' }, 409)
        }
        const created: Category = {
          id: `c${++sequence}`,
          name: body.name,
          description: body.description,
          active: true,
          createdAt: NOW,
          updatedAt: NOW,
        }
        categories.push(created)
        return json(created, 201)
      }

      const match = url.pathname.match(/^\/api\/v1\/categories\/([^/]+)(\/reactivate)?$/)
      const category = categories.find((candidate) => candidate.id === match?.[1])
      if (!category) return json({ code: 'category_not_found', message: 'La categoría no existe' }, 404)

      if (method === 'PATCH') {
        category.name = body.name
        category.description = body.description || null
      } else if (method === 'DELETE') {
        category.active = false
      } else if (method === 'POST' && match?.[2]) {
        category.active = true
      }
      return json(category)
    }),
  )

  return {
    categories,
    requests,
    failNext: (failure: Failure) => failures.push(failure),
    writes: () => requests.filter((request) => request.method !== 'GET'),
  }
}

async function table() {
  return within(await screen.findByRole('table', { name: 'Categorías' }))
}

function rowNames(scope: Awaited<ReturnType<typeof table>>) {
  return scope
    .getAllByRole('row')
    .slice(1)
    .map((row: HTMLElement) => within(row).getAllByRole('cell')[0].textContent)
}

function cards() {
  return within(screen.getByRole('list', { name: 'Categorías' }))
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('CategoriesPage', () => {
  it('lists the active categories sorted by name, as a table and as cards', async () => {
    fakeCategoryApi([
      { name: 'Limpieza', description: 'Artículos de limpieza' },
      { name: 'Almacén' },
      { name: 'Bebidas', active: false },
    ])
    renderWithProviders(<CategoriesPage />)

    expect(screen.getByRole('status', { name: 'Cargando categorías' })).toBeInTheDocument()
    const rows = await table()
    expect(rowNames(rows)).toEqual(['Almacén', 'Limpieza'])
    expect(rows.getByText('Artículos de limpieza')).toBeInTheDocument()
    expect(rows.getAllByText('Activa')).toHaveLength(2)

    const cards = within(screen.getByRole('list', { name: 'Categorías' }))
    expect(cards.getAllByRole('listitem')).toHaveLength(2)
    expect(cards.getByRole('button', { name: 'Editar Almacén' })).toBeInTheDocument()

    expect(screen.getByText('1–2 de 2')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Activas' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Todas' })).toHaveAttribute('aria-pressed', 'false')
  })

  it('invites to create the first category when there are none', async () => {
    const api = fakeCategoryApi()
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    expect(await screen.findByRole('heading', { name: 'Todavía no hay categorías' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Crear la primera' }))

    const sheet = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(sheet.getByLabelText('Nombre'), '  Panadería ')
    await user.type(sheet.getByLabelText(/Descripción/), 'Pan y facturas')
    await user.click(sheet.getByRole('button', { name: 'Crear categoría' }))

    expect(await screen.findByText('Categoría creada')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      { method: 'POST', path: '/api/v1/categories', body: { name: 'Panadería', description: 'Pan y facturas' } },
    ])
    await waitFor(() => expect(screen.queryByRole('dialog', { name: /categoría/ })).not.toBeInTheDocument())
    expect(rowNames(await table())).toEqual(['Panadería'])
  })

  it('validates the form before sending it', async () => {
    const api = fakeCategoryApi([{ name: 'Almacén' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Nueva categoría' }))
    const sheet = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(sheet.getByLabelText('Nombre'), '   ')
    await user.type(sheet.getByLabelText(/Descripción/), 'x'.repeat(201))
    await user.click(sheet.getByRole('button', { name: 'Crear categoría' }))

    expect(sheet.getByText('El nombre es obligatorio')).toBeInTheDocument()
    expect(sheet.getByText('La descripción admite hasta 200 caracteres')).toBeInTheDocument()
    expect(sheet.getByLabelText('Nombre')).toHaveAttribute('aria-invalid', 'true')
    expect(api.writes()).toEqual([])

    await user.type(sheet.getByLabelText('Nombre'), 'Kiosco')
    expect(sheet.queryByText('El nombre es obligatorio')).not.toBeInTheDocument()
  })

  it('shows a duplicated name next to the name field', async () => {
    fakeCategoryApi([{ name: 'Bebidas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Nueva categoría' }))
    const sheet = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(sheet.getByLabelText('Nombre'), 'bebidas')
    await user.click(sheet.getByRole('button', { name: 'Crear categoría' }))

    expect(await sheet.findByText('Ya hay una categoría activa con ese nombre')).toBeInTheDocument()
    expect(sheet.getByLabelText('Nombre')).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByRole('dialog', { name: 'Nueva categoría' })).toBeInTheDocument()
  })

  it('shows unexpected save errors at the bottom of the form', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Nueva categoría' }))
    const sheet = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(sheet.getByLabelText('Nombre'), 'Kiosco')
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(sheet.getByRole('button', { name: 'Crear categoría' }))

    expect(await sheet.findByRole('alert')).toHaveTextContent('El almacenamiento no está disponible')
  })

  it('edits a category with its current values prefilled', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas', description: 'Gaseosas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Bebidas' }))
    const sheet = within(await screen.findByRole('dialog', { name: 'Editar categoría' }))
    expect(sheet.getByLabelText('Nombre')).toHaveValue('Bebidas')
    expect(sheet.getByLabelText(/Descripción/)).toHaveValue('Gaseosas')
    expect(sheet.getByText('8/200 caracteres')).toBeInTheDocument()

    await user.clear(sheet.getByLabelText('Nombre'))
    await user.type(sheet.getByLabelText('Nombre'), 'Bebidas frías')
    await user.clear(sheet.getByLabelText(/Descripción/))
    await user.click(sheet.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('Categoría actualizada')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      { method: 'PATCH', path: '/api/v1/categories/c1', body: { name: 'Bebidas frías', description: '' } },
    ])
    expect(rowNames(await table())).toEqual(['Bebidas frías'])
  })

  it('closes the form without saving when cancelled', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Bebidas' }))
    const sheet = within(await screen.findByRole('dialog', { name: 'Editar categoría' }))
    await user.click(sheet.getByRole('button', { name: 'Cancelar' }))

    await waitFor(() => expect(screen.queryByRole('dialog', { name: /categoría/ })).not.toBeInTheDocument())
    expect(api.writes()).toEqual([])
  })

  it('deactivates a category only after confirming', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }, { name: 'Kiosco' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    await user.click((await table()).getByRole('button', { name: 'Desactivar Kiosco' }))
    let dialog = within(await screen.findByRole('alertdialog'))
    expect(dialog.getByText('¿Desactivar «Kiosco»?')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument())
    expect(api.writes()).toEqual([])

    await user.click((await table()).getByRole('button', { name: 'Desactivar Kiosco' }))
    dialog = within(await screen.findByRole('alertdialog'))
    await user.click(dialog.getByRole('button', { name: 'Desactivar' }))

    expect(await screen.findByText('Categoría desactivada')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/categories/c2', body: undefined }])
    await waitFor(async () => expect(rowNames(await table())).toEqual(['Bebidas']))
  })

  it('reactivates a category from the inactive filter', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas', active: false }, { name: 'Almacén' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Inactivas' }))
    const rows = await table()
    await waitFor(() => expect(rowNames(rows)).toEqual(['Bebidas']))
    expect(rows.getByText('Inactiva')).toBeInTheDocument()

    await user.click(rows.getByRole('button', { name: 'Reactivar Bebidas' }))

    expect(await screen.findByText('Categoría reactivada')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'POST', path: '/api/v1/categories/c1/reactivate', body: undefined }])
  })

  it('offers the same actions on the phone cards', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }, { name: 'Kiosco', active: false }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(cards().getByRole('button', { name: 'Editar Bebidas' }))
    expect(await screen.findByRole('dialog', { name: 'Editar categoría' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cerrar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: /categoría/ })).not.toBeInTheDocument())

    await user.click(cards().getByRole('button', { name: 'Desactivar Bebidas' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Desactivar' }))
    await waitFor(() => expect(api.categories[0].active).toBe(false))

    await user.click(screen.getByRole('button', { name: 'Inactivas' }))
    await waitFor(() => expect(cards().getAllByRole('listitem')).toHaveLength(2))
    await user.click(cards().getByRole('button', { name: 'Reactivar Kiosco' }))

    await waitFor(() => expect(api.categories[1].active).toBe(true))
    expect(api.writes().map((request) => `${request.method} ${request.path}`)).toEqual([
      'DELETE /api/v1/categories/c1',
      'POST /api/v1/categories/c2/reactivate',
    ])
  })

  it('reports a failed status change as a toast', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas', active: false }, { name: 'bebidas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Todas' }))
    const rows = await table()
    await waitFor(() => expect(rows.getAllByRole('row')).toHaveLength(3))
    api.failNext({ status: 409, code: 'category_name_in_use', message: 'Ya hay una categoría activa con ese nombre' })
    await user.click(rows.getByRole('button', { name: 'Reactivar Bebidas' }))

    expect((await screen.findAllByText('No se pudo completar la acción'))[0]).toBeInTheDocument()
    expect(screen.getAllByText('Ya hay una categoría activa con ese nombre')[0]).toBeInTheDocument()
  })

  it('searches by name and explains when nothing matches', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }, { name: 'Almacén' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.type(screen.getByRole('searchbox', { name: 'Buscar categorías' }), 'beb')
    await waitFor(async () => expect(rowNames(await table())).toEqual(['Bebidas']))
    expect(api.requests.at(-1)?.path).toBe('/api/v1/categories')

    await user.type(screen.getByRole('searchbox', { name: 'Buscar categorías' }), 'xyz')
    expect(await screen.findByRole('heading', { name: 'No encontramos categorías' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Limpiar búsqueda' }))
    await waitFor(async () => expect(rowNames(await table())).toEqual(['Almacén', 'Bebidas']))
  })

  it('pages through long listings', async () => {
    fakeCategoryApi(Array.from({ length: 30 }, (_, index) => ({ name: `Categoría ${String(index + 1).padStart(2, '0')}` })))
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    expect(screen.getByText('1–25 de 30')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Anterior' })).toBeDisabled()

    await user.click(screen.getByRole('button', { name: 'Siguiente' }))

    expect(await screen.findByText('26–30 de 30')).toBeInTheDocument()
    expect(rowNames(await table())[0]).toBe('Categoría 26')
    expect(screen.getByRole('button', { name: 'Siguiente' })).toBeDisabled()
  })

  it('offers to retry when the list cannot be loaded', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }])
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    expect(await screen.findByRole('heading', { name: 'No pudimos cargar las categorías' })).toBeInTheDocument()
    expect(screen.getByText('El almacenamiento no está disponible')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(rowNames(await table())).toEqual(['Bebidas'])
  })
})
