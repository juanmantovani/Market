import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '../../../test/render'
import type { Customer } from '../types'
import CustomersPage from './CustomersPage'

const NOW = '2026-11-10T15:00:00Z'

type Failure = { status: number; code: string; message: string }
type CustomerSeed = Omit<Customer, 'id'>

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function customer(overrides: Partial<Customer> & { name: string }): CustomerSeed {
  return {
    documentType: 'none',
    documentNumber: '',
    taxCondition: 'final_consumer',
    email: null,
    phone: null,
    address: null,
    notes: null,
    createdAt: NOW,
    updatedAt: NOW,
    ...overrides,
  }
}

// An in-memory stand-in for /api/v1/customers that follows the backend's
// rules: one customer per document, deletes remove the record, sorted by
// name, searchable by name, document digits, email and phone.
function fakeCustomerApi(initial: CustomerSeed[] = []) {
  let sequence = 0
  const customers: Customer[] = initial.map((entry) => ({ ...entry, id: `cu${++sequence}` }))
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

      if (url.pathname === '/api/v1/customers' && method === 'GET') {
        const search = (url.searchParams.get('q') ?? '').toLowerCase()
        const digits = search.replace(/\D/g, '')
        const page = Number(url.searchParams.get('page') ?? 1)
        const limit = Number(url.searchParams.get('limit') ?? 25)
        const matches = customers
          .filter(
            (entry) =>
              `${entry.name} ${entry.email ?? ''} ${entry.phone ?? ''}`.toLowerCase().includes(search) ||
              (digits !== '' && entry.documentNumber.includes(digits)),
          )
          .sort((a, b) => a.name.localeCompare(b.name))
        return json({ items: matches.slice((page - 1) * limit, page * limit), total: matches.length, page, limit })
      }

      if (url.pathname === '/api/v1/customers' && method === 'POST') {
        const taken = customers.some(
          (entry) =>
            body.documentType !== 'none' &&
            entry.documentType === body.documentType &&
            entry.documentNumber === body.documentNumber,
        )
        if (taken) {
          return json({ code: 'customer_document_in_use', message: 'Ya hay un cliente con ese documento' }, 409)
        }
        const created: Customer = { ...body, id: `cu${++sequence}`, createdAt: NOW, updatedAt: NOW }
        customers.push(created)
        return json(created, 201)
      }

      const id = url.pathname.match(/^\/api\/v1\/customers\/([^/]+)$/)?.[1]
      const index = customers.findIndex((entry) => entry.id === id)
      if (index === -1) return json({ code: 'customer_not_found', message: 'El cliente no existe' }, 404)

      if (method === 'DELETE') {
        customers.splice(index, 1)
        return new Response(null, { status: 204 })
      }
      if (method === 'PUT') {
        customers[index] = { ...customers[index], ...body }
      }
      return json(customers[index])
    }),
  )

  return {
    customers,
    requests,
    failNext: (failure: Failure) => failures.push(failure),
    writes: () => requests.filter((request) => request.method !== 'GET'),
  }
}

async function table() {
  return within(await screen.findByRole('table', { name: 'Clientes' }))
}

function rowTexts(scope: Awaited<ReturnType<typeof table>>) {
  return scope
    .getAllByRole('row')
    .slice(1)
    .map((row: HTMLElement) => within(row).getAllByRole('cell').slice(0, 4).map((cell) => cell.textContent))
}

async function openNewCustomer(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'Nuevo cliente' }))
  return within(await screen.findByRole('dialog', { name: 'Nuevo cliente' }))
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('CustomersPage', () => {
  it('lists customers with their document, tax condition and contact, as a table and as cards', async () => {
    fakeCustomerApi([
      customer({ name: 'Ana Pérez', documentType: 'dni', documentNumber: '28456789', phone: '351 555-0101' }),
      customer({
        name: 'Kiosco La Esquina SRL',
        documentType: 'cuit',
        documentNumber: '30712345678',
        taxCondition: 'registered',
        email: 'compras@laesquina.com.ar',
        phone: '351 555-0303',
      }),
      customer({ name: 'Juan del barrio' }),
      customer({ name: 'Lucía Sosa', documentType: 'cuil', documentNumber: '27234567894', email: 'lucia@example.com' }),
    ])
    renderWithProviders(<CustomersPage />)

    expect(rowTexts(await table())).toEqual([
      ['Ana Pérez', 'DNI 28.456.789', 'Consumidor final', '351 555-0101'],
      ['Juan del barrio', '—', 'Consumidor final', '—'],
      ['Kiosco La Esquina SRL', 'CUIT 30-71234567-8', 'Responsable inscripto', '351 555-0303 · compras@laesquina.com.ar'],
      ['Lucía Sosa', 'CUIL 27-23456789-4', 'Consumidor final', 'lucia@example.com'],
    ])

    const cards = within(screen.getByRole('list', { name: 'Clientes' }))
    expect(cards.getAllByRole('listitem')).toHaveLength(4)
    expect(cards.getByText('CUIT 30-71234567-8 · Responsable inscripto')).toBeInTheDocument()
    expect(cards.getByText('Consumidor final')).toBeInTheDocument()
    expect(cards.getByRole('button', { name: 'Eliminar Ana Pérez' })).toBeInTheDocument()
  })

  it('creates a customer, sending the document as digits and empty fields as null', async () => {
    const api = fakeCustomerApi()
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)

    expect(await screen.findByRole('heading', { name: 'Todavía no hay clientes' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Crear el primero' }))
    const form = within(await screen.findByRole('dialog', { name: 'Nuevo cliente' }))

    expect(form.getByLabelText('Tipo de documento')).toHaveValue('dni')
    expect(form.getByLabelText('Condición frente al IVA')).toHaveValue('final_consumer')
    await user.type(form.getByLabelText('Nombre o razón social'), 'Kiosco La Esquina SRL')
    await user.selectOptions(form.getByLabelText('Tipo de documento'), 'CUIT')
    await user.type(form.getByLabelText('Número de documento'), '30-71234567-8')
    await user.selectOptions(form.getByLabelText('Condición frente al IVA'), 'Responsable inscripto')
    await user.type(form.getByLabelText(/Email/), '  compras@laesquina.com.ar ')
    await user.type(form.getByLabelText(/Dirección/), 'San Martín 123')
    await user.click(form.getByRole('button', { name: 'Crear cliente' }))

    expect(await screen.findByText('Cliente creado')).toBeInTheDocument()
    expect(api.writes()).toEqual([
      {
        method: 'POST',
        path: '/api/v1/customers',
        body: {
          name: 'Kiosco La Esquina SRL',
          documentType: 'cuit',
          documentNumber: '30712345678',
          taxCondition: 'registered',
          email: 'compras@laesquina.com.ar',
          phone: null,
          address: 'San Martín 123',
          notes: null,
        },
      },
    ])
    expect(rowTexts(await table())[0].slice(0, 2)).toEqual(['Kiosco La Esquina SRL', 'CUIT 30-71234567-8'])
  })

  it('validates every field at once before sending', async () => {
    const api = fakeCustomerApi([customer({ name: 'Ana' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)
    await table()

    const form = await openNewCustomer(user)
    await user.type(form.getByLabelText('Número de documento'), '12.345')
    await user.selectOptions(form.getByLabelText('Condición frente al IVA'), 'Monotributista')
    await user.type(form.getByLabelText(/Email/), 'ana@')
    await user.click(form.getByRole('button', { name: 'Crear cliente' }))

    expect(form.getByText('El nombre es obligatorio')).toBeInTheDocument()
    expect(form.getByText('El DNI lleva 7 u 8 números')).toBeInTheDocument()
    expect(form.getByText('Para monotributo, responsable inscripto o exento hace falta un CUIT')).toBeInTheDocument()
    expect(form.getByText('Ingresar un email válido, por ejemplo nombre@dominio.com')).toBeInTheDocument()
    expect(form.getByLabelText('Número de documento')).toHaveAttribute('aria-invalid', 'true')
    expect(api.writes()).toEqual([])

    await user.type(form.getByLabelText('Nombre o razón social'), 'A')
    expect(form.queryByText('El nombre es obligatorio')).not.toBeInTheDocument()
    await user.selectOptions(form.getByLabelText('Tipo de documento'), 'CUIT')
    expect(form.queryByText('El DNI lleva 7 u 8 números')).not.toBeInTheDocument()
    expect(form.queryByText(/hace falta un CUIT/)).not.toBeInTheDocument()
  })

  it.each([
    ['DNI', '', 'Ingresar el número de DNI'],
    ['DNI', '28456A89', 'El DNI lleva solo números'],
    ['CUIT', '20-1234567-8', 'El CUIT lleva 11 números'],
    ['CUIL', '27 2345 6789', 'El CUIL lleva 11 números'],
  ])('rejects a %s written as «%s»', async (type, number, message) => {
    const api = fakeCustomerApi([customer({ name: 'Ana' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)
    await table()

    const form = await openNewCustomer(user)
    await user.type(form.getByLabelText('Nombre o razón social'), 'Beto')
    await user.selectOptions(form.getByLabelText('Tipo de documento'), type)
    if (number) await user.type(form.getByLabelText('Número de documento'), number)
    await user.click(form.getByRole('button', { name: 'Crear cliente' }))

    expect(form.getByText(message)).toBeInTheDocument()
    expect(api.writes()).toEqual([])
  })

  it('saves a customer without document', async () => {
    const api = fakeCustomerApi([customer({ name: 'Ana' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)
    await table()

    const form = await openNewCustomer(user)
    await user.type(form.getByLabelText('Nombre o razón social'), 'Juan del barrio')
    await user.type(form.getByLabelText('Número de documento'), '123')
    await user.selectOptions(form.getByLabelText('Tipo de documento'), 'Sin documento')

    expect(form.getByLabelText('Número de documento')).toBeDisabled()
    expect(form.getByLabelText('Número de documento')).toHaveValue('')
    await user.click(form.getByRole('button', { name: 'Crear cliente' }))

    await waitFor(() => expect(api.writes()).toHaveLength(1))
    expect(api.writes()[0].body).toMatchObject({ documentType: 'none', documentNumber: '', taxCondition: 'final_consumer' })
  })

  it('shows a duplicated document next to the document field', async () => {
    fakeCustomerApi([customer({ name: 'Ana Pérez', documentType: 'dni', documentNumber: '28456789' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)
    await table()

    const form = await openNewCustomer(user)
    await user.type(form.getByLabelText('Nombre o razón social'), 'Otra Ana')
    await user.type(form.getByLabelText('Número de documento'), '28.456.789')
    await user.click(form.getByRole('button', { name: 'Crear cliente' }))

    expect(await form.findByText('Ya hay un cliente con ese documento')).toBeInTheDocument()
    expect(form.getByLabelText('Número de documento')).toHaveAttribute('aria-invalid', 'true')
  })

  it('shows unexpected save errors at the bottom of the form', async () => {
    const api = fakeCustomerApi([customer({ name: 'Ana' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar Ana' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar cliente' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await form.findByRole('alert')).toHaveTextContent('El almacenamiento no está disponible')
  })

  it('edits a customer with its current values and can clear optional fields', async () => {
    const api = fakeCustomerApi([
      customer({
        name: 'María López',
        documentType: 'cuit',
        documentNumber: '27234567894',
        taxCondition: 'monotax',
        email: 'maria@example.com',
        phone: '351 555-0404',
        notes: 'Paga a fin de mes',
      }),
    ])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)

    await user.click((await table()).getByRole('button', { name: 'Editar María López' }))
    const form = within(await screen.findByRole('dialog', { name: 'Editar cliente' }))
    expect(form.getByLabelText('Número de documento')).toHaveValue('27-23456789-4')
    expect(form.getByLabelText('Condición frente al IVA')).toHaveValue('monotax')
    expect(form.getByLabelText(/Notas/)).toHaveValue('Paga a fin de mes')

    await user.clear(form.getByLabelText(/Email/))
    await user.clear(form.getByLabelText(/Notas/))
    await user.click(form.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('Cliente actualizado')).toBeInTheDocument()
    const [write] = api.writes()
    expect(write.method).toBe('PUT')
    expect(write.path).toBe('/api/v1/customers/cu1')
    expect(write.body).toEqual({
      name: 'María López',
      documentType: 'cuit',
      documentNumber: '27234567894',
      taxCondition: 'monotax',
      email: null,
      phone: '351 555-0404',
      address: null,
      notes: null,
    })
  })

  it('deletes a customer only after confirming', async () => {
    const api = fakeCustomerApi([customer({ name: 'Ana' }), customer({ name: 'Beto' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Ana' }))
    const dialog = within(await screen.findByRole('alertdialog'))
    expect(dialog.getByText('¿Eliminar «Ana»?')).toBeInTheDocument()
    await user.click(dialog.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('Cliente eliminado')).toBeInTheDocument()
    expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/customers/cu1', body: undefined }])
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[0])).toEqual(['Beto']))
  })

  it('offers the same actions on the phone cards', async () => {
    const api = fakeCustomerApi([customer({ name: 'Ana' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)
    await table()
    const cards = () => within(screen.getByRole('list', { name: 'Clientes' }))

    await user.click(cards().getByRole('button', { name: 'Editar Ana' }))
    expect(await screen.findByRole('dialog', { name: 'Editar cliente' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: /cliente/ })).not.toBeInTheDocument())

    await user.click(cards().getByRole('button', { name: 'Eliminar Ana' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    await waitFor(() => expect(api.writes()).toEqual([{ method: 'DELETE', path: '/api/v1/customers/cu1', body: undefined }]))
  })

  it('reports a failed deletion as a toast', async () => {
    const api = fakeCustomerApi([customer({ name: 'Ana' })])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)

    await user.click((await table()).getByRole('button', { name: 'Eliminar Ana' }))
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect((await screen.findAllByText('No se pudo eliminar el cliente'))[0]).toBeInTheDocument()
    expect(api.customers).toHaveLength(1)
  })

  it('goes back a page after deleting the last customer of the page', async () => {
    fakeCustomerApi(Array.from({ length: 26 }, (_, index) => customer({ name: `Cliente ${String(index + 1).padStart(2, '0')}` })))
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)
    await table()

    await user.click(screen.getByRole('button', { name: 'Siguiente' }))
    expect(await screen.findByText('26–26 de 26')).toBeInTheDocument()
    await user.click((await table()).getByRole('button', { name: 'Eliminar Cliente 26' }))
    await user.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('1–25 de 25')).toBeInTheDocument()
  })

  it('searches by name or by document written with dots', async () => {
    const api = fakeCustomerApi([
      customer({ name: 'Ana Pérez', documentType: 'dni', documentNumber: '28456789' }),
      customer({ name: 'Beto Gómez', documentType: 'dni', documentNumber: '30111222' }),
    ])
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)
    await table()

    expect(screen.getByRole('searchbox', { name: 'Buscar clientes' })).toHaveAccessibleDescription('Por nombre, documento, email o teléfono')
    await user.type(screen.getByRole('searchbox', { name: 'Buscar clientes' }), '30.111')
    await waitFor(async () => expect(rowTexts(await table()).map((row) => row[0])).toEqual(['Beto Gómez']))
    expect(api.requests.at(-1)?.path).toContain('q=30.111')

    await user.clear(screen.getByRole('searchbox', { name: 'Buscar clientes' }))
    await user.type(screen.getByRole('searchbox', { name: 'Buscar clientes' }), 'zzz')
    expect(await screen.findByRole('heading', { name: 'No se encontraron clientes' })).toBeInTheDocument()
  })

  it('offers to retry when the list cannot be loaded', async () => {
    const api = fakeCustomerApi([customer({ name: 'Ana' })])
    api.failNext({ status: 503, code: 'storage_unavailable', message: 'El almacenamiento no está disponible' })
    const user = userEvent.setup()
    renderWithProviders(<CustomersPage />)

    expect(await screen.findByRole('heading', { name: 'No se pudieron cargar los clientes' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(rowTexts(await table()).map((row) => row[0])).toEqual(['Ana'])
  })
})
