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
// rules as the backend: names are unique, deletes remove the record, lists
// are sorted by name and paginated.
function fakeCategoryApi(initial: Array<Partial<Category> & { name: string }> = []) {
  let sequence = 0
  const categories: Category[] = initial.map((category) => ({
    id: `c${++sequence}`,
    description: null,
    createdAt: NOW,
    updatedAt: NOW,
    ...category,
  }))
  const requests: Array<{ method: string; path: string; body: unknown }> = []
  const failures: Failure[] = []

  const nameTaken = (name: string) =>
    categories.some((category) => category.name.toLowerCase() === name.toLowerCase())

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
        const page = Number(url.searchParams.get('page') ?? 1)
        const limit = Number(url.searchParams.get('limit') ?? 25)
        const matches = categories
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
          return json({ code: 'category_name_in_use', message: 'Ya hay una categoría con ese nombre' }, 409)
        }
        const created: Category = {
          id: `c${++sequence}`,
          name: body.name,
          description: body.description,
          createdAt: NOW,
          updatedAt: NOW,
        }
        categories.push(created)
        return json(created, 201)
      }

      const id = url.pathname.match(/^\/api\/v1\/categories\/([^/]+)$/)?.[1]
      const index = categories.findIndex((candidate) => candidate.id === id)
      if (index === -1) return json({ code: 'category_not_found', message: 'La categoría no existe' }, 404)

      if (method === 'DELETE') {
        categories.splice(index, 1)
        return new Response(null, { status: 204 })
      }
      if (method === 'PATCH') {
        categories[index].name = body.name
        categories[index].description = body.description || null
      }
      return json(categories[index])
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
  it('lists the categories sorted by name, as a table and as cards', async () => {
    fakeCategoryApi([{ name: 'Limpieza', description: 'Artículos de limpieza' }, { name: 'Almacén' }])
    renderWithProviders(<CategoriesPage />)

    expect(screen.getByRole('status', { name: 'Cargando categorías' })).toBeInTheDocument()
    const rows = await table()
    expect(rowNames(rows)).toEqual(['Almacén', 'Limpieza'])
    expect(rows.getByText('Artículos de limpieza')).toBeInTheDocument()

    expect(cards().getAllByRole('listitem')).toHaveLength(2)
    expect(cards().getByRole('button', { name: 'Editar Almacén' })).toBeInTheDocument()
    expect(cards().getByRole('button', { name: 'Eliminar Almacén' })).toBeInTheDocument()

    expect(screen.getByText('1–2 de 2')).toBeInTheDocument()
  })

  it('invites to create the first category when there are none', async () => {
    const api = fakeCategoryApi()
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    expect(await screen.findByRole('heading', { name: 'Todavía no hay categorías' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Crear la primera' }))

    const form = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(form.getByLabelText('Nombre'), '  Panadería ')
    await user.type(form.getByLabelText(/Descripción/), 'Pan y facturas')
    await user.click(form.getByRole('button', { name: 'Crear categoría' }))

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
    const form = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(form.getByLabelText('Nombre'), '   ')
    await user.type(form.getByLabelText(/Descripción/), 'x'.repeat(201))
    await user.click(form.getByRole('button', { name: 'Crear categoría' }))

    expect(form.getByText('El nombre es obligatorio')).toBeInTheDocument()
    expect(form.getByText('La descripción admite hasta 200 caracteres')).toBeInTheDocument()
    expect(form.getByLabelText('Nombre')).toHaveAttribute('aria-invalid', 'true')
    expect(api.writes()).toEqual([])

    await user.type(form.getByLabelText('Nombre'), 'Kiosco')
    expect(form.queryByText('El nombre es obligatorio')).not.toBeInTheDocument()
  })

  it('shows a duplicated name next to the name field', async () => {
    fakeCategoryApi([{ name: 'Bebidas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Nueva categoría' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(form.getByLabelText('Nombre'), 'bebidas')
    await user.click(form.getByRole('button', { name: 'Crear categoría' }))

    expect(await form.findByText('Ya hay una categoría con ese nombre')).toBeInTheDocument()
    expect(form.getByLabelText('Nombre')).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByRole('dialog', { name: 'Nueva categoría' })).toBeInTheDocument()
  })

  it('shows unexpected save errors at the bottom of the form', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Nueva categoría' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nueva categoría' }))
    await user.type(form.getByLabelText('Nombre'), 'Kiosco')
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(form.getByRole('button', { name: 'Crear categoría' }))

    expect(await form.findByRole('alert')).toHaveTextContent('El almacenamiento no está disponible')
  })

  it('edits a category with its current values prefilled', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas', description: 'Gaseosas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Bebidas' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar categoría' }))
    expect(form.getByLabelText('Nombre')).toHaveValue('Bebidas')
    expect(form.getByLabelText(/Descripción/)).toHaveValue('Gaseosas')
    expect(form.getByText('8/200 caracteres')).toBeInTheDocument()

    await user.clear(form.getByLabelText('Nombre'))
    await user.type(form.getByLabelText('Nombre'), 'Bebidas frías')
    await user.clear(form.getByLabelText(/Descripción/))
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

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
    const form = within(await screen.findByRole('dialog', { name: 'Editar categoría' }))
    await user.click(form.getByRole('button', { name: 'Cancelar' }))

    await waitFor(() => expect(screen.queryByRole('dialog', { name: /categoría/ })).not.toBeInTheDocument())
    expect(api.writes()).toEqual([])
  })

  it('deletes a category only after confirming', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }, { name: 'Kiosco' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Kiosco' }))
    let dialog = within(await screen.findByRole('alertdialog'))
    expect(dialog.getByText('¿Eliminar «Kiosco»?')).toBeInTheDocument()
    expect(dialog.getByText('Se borra definitivamente y no se puede deshacer.')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument())
    expect(api.writes()).toEqual([])

    await user.click((await table()).getByRole('button', { name: 'Eliminar Kiosco' }))
    dialog = within(await screen.findByRole('alertdialog'))
    await user.click(dialog.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('Categoría eliminada')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/categories/c2', body: undefined }])
    await waitFor(async () => expect(rowNames(await table())).toEqual(['Bebidas']))
  })

  it('offers the same actions on the phone cards', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }, { name: 'Kiosco' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(cards().getByRole('button', { name: 'Editar Bebidas' }))
    expect(await screen.findByRole('dialog', { name: 'Editar categoría' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cerrar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: /categoría/ })).not.toBeInTheDocument())

    await user.click(cards().getByRole('button', { name: 'Eliminar Bebidas' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    await waitFor(() => expect(cards().getAllByRole('listitem')).toHaveLength(1))
    expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/categories/c1', body: undefined }])
  })

  it('reports a failed deletion as a toast', async () => {
    const api = fakeCategoryApi([{ name: 'Bebidas' }])
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Bebidas' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect((await screen.findAllByText('No se pudo eliminar la categoría'))[0]).toBeInTheDocument()
    expect(screen.getAllByText('El almacenamiento no está disponible')[0]).toBeInTheDocument()
    expect(api.categories).toHaveLength(1)
  })

  it('goes back a page after deleting the last category of the page', async () => {
    fakeCategoryApi(Array.from({ length: 26 }, (_, index) => ({ name: `Categoría ${String(index + 1).padStart(2, '0')}` })))
    const user = userEvent.setup()
    renderWithProviders(<CategoriesPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Siguiente' }))
    expect(await screen.findByText('26–26 de 26')).toBeInTheDocument()

    await user.click((await table()).getByRole('button', { name: 'Eliminar Categoría 26' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('1–25 de 25')).toBeInTheDocument()
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
