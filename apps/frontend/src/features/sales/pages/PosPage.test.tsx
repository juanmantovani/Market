import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { formatMoney } from '../../../shared/lib/money'
import { renderWithProviders } from '../../../test/render'
import { CART_STORAGE_KEY } from '../hooks/use-cart'
import { SELLER_STORAGE_KEY } from '../hooks/use-seller'
import type { CatalogItem, CatalogParams, CategoryOption, CustomerOption, PartyParams, SellerOption } from '../types'
import PosPage from './PosPage'

const CATEGORIES: CategoryOption[] = [
  { id: 'c1', name: 'Bebidas' },
  { id: 'c2', name: 'Panadería' },
]

const ITEMS: CatalogItem[] = [
  { id: 'agua', code: 'BEB-002', barcode: '7790000000066', name: 'Agua mineral', unit: 'unit', salePrice: 110_000, categoryId: 'c1' },
  { id: 'facturas', code: 'PAN-002', barcode: null, name: 'Facturas', unit: 'unit', salePrice: 60_000, categoryId: 'c2' },
  { id: 'coca', code: 'BEB-001', barcode: '7790000000059', name: 'Gaseosa cola', unit: 'unit', salePrice: 320_000, categoryId: 'c1' },
  { id: 'pan', code: 'PAN-001', barcode: null, name: 'Pan francés', unit: 'kg', salePrice: 280_000, categoryId: 'c2' },
]

// Answers like GET /api/v1/items: search by name, code or barcode, filter by
// category, sorted by name.
function fakeCatalog() {
  return vi.fn(async ({ search, categoryId, limit }: CatalogParams) => {
    const wanted = search.trim().toLowerCase()
    const matches = ITEMS.filter(
      (item) =>
        (!categoryId || item.categoryId === categoryId) &&
        (!wanted || [item.name, item.code, item.barcode ?? ''].some((text) => text.toLowerCase().includes(wanted))),
    )
    return { items: matches.slice(0, limit), total: matches.length, page: 1, limit }
  })
}

const CUSTOMERS: CustomerOption[] = [
  { id: 'ana', name: 'Ana Pérez', documentType: 'dni', documentNumber: '28456789', system: false },
  { id: 'final', name: 'Consumidor Final', documentType: 'none', documentNumber: '', system: true },
  { id: 'esquina', name: 'Kiosco La Esquina SRL', documentType: 'cuit', documentNumber: '30712345678', system: false },
]

const SELLERS: SellerOption[] = [
  { id: 'jp', name: 'Juan Pérez', code: 'JP' },
  { id: 'lf', name: 'Lucía Fernández', code: 'LF' },
  { id: 'mg', name: 'María Gómez', code: 'MG' },
]

// Answers like the customers and sellers APIs: search by name or code,
// sorted by name.
function fakeParties<T extends { name: string }>(records: T[], fields: (record: T) => string[]) {
  return vi.fn(async ({ search, limit }: PartyParams) => {
    const wanted = search.trim().toLowerCase()
    const matches = records.filter((record) => !wanted || fields(record).some((text) => text.toLowerCase().includes(wanted)))
    return { items: matches.slice(0, limit), total: matches.length, page: 1, limit }
  })
}

function renderPos(searchCatalog = fakeCatalog(), searchCustomers = fakeParties(CUSTOMERS, (customer) => [customer.name, customer.documentNumber])) {
  const user = userEvent.setup()
  const searchSellers = fakeParties(SELLERS, (seller) => [seller.name, seller.code])
  const view = renderWithProviders(
    <PosPage categories={CATEGORIES} searchCatalog={searchCatalog} searchCustomers={searchCustomers} searchSellers={searchSellers} />,
  )
  return { user, searchCatalog, searchCustomers, ...view }
}

// Testing Library compares text with its spaces collapsed, and formatMoney
// puts a non-breaking space after the sign.
const money = (cents: number) => formatMoney(cents).replace(/\s/g, ' ')
const moneyPattern = (cents: number) => money(cents).replace(/[$.]/g, '\\$&')

const catalog = () => within(screen.getByRole('list', { name: 'Artículos' }))
const cart = () => within(screen.getByRole('region', { name: 'Venta actual' }))
const search = () => screen.getByRole('searchbox', { name: 'Buscar o escanear' })

beforeEach(() => {
  window.localStorage.clear()
})

describe('PosPage', () => {
  it('shows the catalog and narrows it by category and by search', async () => {
    const { user } = renderPos()

    expect(screen.getByRole('heading', { name: 'Punto de venta', level: 1 })).toBeInTheDocument()
    expect(await catalog().findByRole('button', { name: /Gaseosa cola/ })).toBeInTheDocument()
    expect(catalog().getByRole('button', { name: /Pan francés/ })).toHaveTextContent(`${money(280_000)} / kg`)
    expect(search()).toHaveAccessibleDescription(/Por nombre, código o código de barras/)

    await user.click(screen.getByRole('button', { name: 'Panadería' }))
    expect(screen.getByRole('button', { name: 'Panadería' })).toHaveAttribute('aria-pressed', 'true')
    await waitFor(() => expect(catalog().queryByRole('button', { name: /Gaseosa cola/ })).not.toBeInTheDocument())
    expect(catalog().getAllByRole('button').map((button) => button.textContent)).toEqual([
      expect.stringContaining('Facturas'),
      expect.stringContaining('Pan francés'),
    ])

    await user.type(search(), 'fact')
    await waitFor(() => expect(catalog().getAllByRole('button')).toHaveLength(1))

    await user.type(search(), 'xyz')
    expect(await screen.findByRole('heading', { name: 'No hay artículos que coincidan' })).toBeInTheDocument()
  })

  it('adds items, changes quantities and takes lines back with undo', async () => {
    const { user } = renderPos()
    expect(cart().getByText('Todavía no hay artículos en la venta')).toBeInTheDocument()

    await user.click(await catalog().findByRole('button', { name: /Gaseosa cola/ }))
    await user.click(catalog().getByRole('button', { name: /Gaseosa cola/ }))
    expect(cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' })).toHaveValue('2')
    expect(cart().getByText(money(640_000), { selector: 'dd' })).toBeInTheDocument()
    expect(screen.getByText(/Se agregó Gaseosa cola\. Total:/)).toBeInTheDocument()

    await user.click(cart().getByRole('button', { name: 'Sumar 1 u. a Gaseosa cola' }))
    expect(cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' })).toHaveValue('3')
    await user.click(cart().getByRole('button', { name: 'Restar 1 u. de Gaseosa cola' }))
    expect(cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' })).toHaveValue('2')

    const quantity = cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' })
    await user.clear(quantity)
    await user.type(quantity, '5{Enter}')
    expect(cart().getByText(money(1_600_000), { selector: 'dd' })).toBeInTheDocument()

    await user.clear(cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' }))
    await user.type(cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' }), '1,5{Enter}')
    expect(cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' })).toHaveAccessibleDescription('Ingresar una cantidad entera mayor a cero')
    await user.keyboard('{Escape}')
    expect(cart().getByRole('textbox', { name: 'Cantidad de Gaseosa cola' })).toHaveValue('5')
    expect(cart().queryByText('Ingresar una cantidad entera mayor a cero')).not.toBeInTheDocument()

    await user.click(catalog().getByRole('button', { name: /Agua mineral/ }))
    await user.click(cart().getByRole('button', { name: 'Quitar Gaseosa cola' }))
    expect(cart().queryByText('Gaseosa cola')).not.toBeInTheDocument()
    expect(cart().getByText(money(110_000), { selector: 'dd' })).toBeInTheDocument()

    await user.click(await screen.findByRole('button', { name: 'Deshacer' }))
    expect(cart().getAllByRole('listitem').map((item) => item.textContent)).toEqual([
      expect.stringContaining('Gaseosa cola'),
      expect.stringContaining('Agua mineral'),
    ])

    await user.click(cart().getByRole('button', { name: 'Vaciar' }))
    expect(cart().getByText('Todavía no hay artículos en la venta')).toBeInTheDocument()
    expect(await screen.findByText('Se vació la venta')).toBeInTheDocument()
    const undoButtons = screen.getAllByRole('button', { name: 'Deshacer' })
    await user.click(undoButtons[undoButtons.length - 1])
    expect(cart().getByText(money(1_710_000), { selector: 'dd' })).toBeInTheDocument()
  })

  it('adds the item a barcode scanner reads and explains a code it cannot find', async () => {
    const { user } = renderPos()
    await catalog().findByRole('button', { name: /Gaseosa cola/ })

    // Two scans in a row: the second starts before the first lookup answers.
    await user.type(search(), '7790000000059{Enter}beb-002{Enter}')
    expect(await cart().findByText('Gaseosa cola')).toBeInTheDocument()
    expect(await cart().findByText('Agua mineral')).toBeInTheDocument()
    expect(search()).toHaveValue('')

    await user.type(search(), 'nada{Enter}')
    expect(await screen.findByRole('status')).toHaveTextContent('No hay artículos con “nada”. Probar con otra búsqueda.')

    await user.clear(search())
    await user.type(search(), 'BEB{Enter}')
    expect(await screen.findByRole('status')).toHaveTextContent('Hay varios artículos que coinciden: elegir uno de la lista.')
    expect(cart().getAllByRole('listitem')).toHaveLength(2)

    await user.keyboard('{Escape}')
    expect(search()).toHaveValue('')
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('asks for the quantity of an item sold by weight', async () => {
    const { user } = renderPos()

    await user.click(await catalog().findByRole('button', { name: /Pan francés/ }))
    const dialog = within(await screen.findByRole('dialog', { name: 'Cantidad de Pan francés' }))
    await user.click(dialog.getByRole('button', { name: 'Agregar' }))
    expect(dialog.getByRole('textbox', { name: 'Cantidad (kg)' })).toHaveAccessibleDescription(
      'Ingresar una cantidad mayor a cero, por ejemplo 0,750 kg',
    )

    await user.type(dialog.getByRole('textbox', { name: 'Cantidad (kg)' }), '0,750')
    expect(dialog.getByText(new RegExp(`Importe: ${moneyPattern(210_000)}`))).toBeInTheDocument()
    await user.keyboard('{Enter}')

    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(cart().getByRole('textbox', { name: 'Cantidad de Pan francés' })).toHaveValue('0,75')
    expect(cart().getByText(money(210_000), { selector: 'dd' })).toBeInTheDocument()
    await user.click(cart().getByRole('button', { name: 'Sumar 0,1 kg a Pan francés' }))
    expect(cart().getByRole('textbox', { name: 'Cantidad de Pan francés' })).toHaveValue('0,85')

    await user.click(catalog().getByRole('button', { name: /Pan francés/ }))
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(cart().getByRole('textbox', { name: 'Cantidad de Pan francés' })).toHaveValue('0,85')
  })

  it('works from the keyboard: F2 searches and the arrows walk the catalog', async () => {
    const { user } = renderPos()
    await catalog().findByRole('button', { name: /Gaseosa cola/ })

    await user.keyboard('{F2}')
    expect(search()).toHaveFocus()

    await user.keyboard('{ArrowDown}')
    expect(catalog().getByRole('button', { name: /Agua mineral/ })).toHaveFocus()
    await user.keyboard('{ArrowRight}')
    expect(catalog().getByRole('button', { name: /Facturas/ })).toHaveFocus()
    await user.keyboard('{Enter}')
    expect(cart().getByText('Facturas')).toBeInTheDocument()
    await user.keyboard('{ArrowLeft}{ArrowUp}')
    expect(search()).toHaveFocus()

    await user.keyboard('{ArrowDown}{Escape}')
    expect(search()).toHaveFocus()
  })

  it('keeps the cart after a reload and ignores a damaged one', async () => {
    const { user, unmount } = renderPos()
    await user.click(await catalog().findByRole('button', { name: /Gaseosa cola/ }))
    expect(JSON.parse(window.localStorage.getItem(CART_STORAGE_KEY)!)).toMatchObject({ version: 2, lines: [{ itemId: 'coca', quantity: 1_000 }] })
    unmount()

    renderPos()
    expect(cart().getByText('Gaseosa cola')).toBeInTheDocument()
    await user.click(cart().getByRole('button', { name: 'Quitar Gaseosa cola' }))
    expect(window.localStorage.getItem(CART_STORAGE_KEY)).toBeNull()
  })

  it('reads a cart saved before discounts existed and one with discounts', () => {
    const stored = { code: 'BEB-001', name: 'Gaseosa cola', unit: 'unit', unitPrice: 320_000, quantity: 2_000 }
    window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify({ version: 1, lines: [{ itemId: 'coca', ...stored }] }))
    const { unmount } = renderPos()
    expect(cart().getByText('Gaseosa cola')).toBeInTheDocument()
    expect(cart().getByText(money(640_000), { selector: 'dd' })).toBeInTheDocument()
    unmount()

    window.localStorage.setItem(
      CART_STORAGE_KEY,
      JSON.stringify({
        version: 2,
        lines: [{ itemId: 'coca', ...stored, discount: { type: 'percent', value: 1_000 } }],
        cartDiscount: { type: 'amount', value: 10_000 },
      }),
    )
    renderPos()
    expect(cart().getByText('Descuento 10 %')).toBeInTheDocument()
    expect(cart().getByText(money(566_000), { selector: 'dd' })).toBeInTheDocument()
  })

  it('starts empty when the stored cart is damaged', () => {
    window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify({ version: 1, lines: [{ itemId: 'coca', quantity: -1 }] }))
    renderPos()
    expect(cart().getByText('Todavía no hay artículos en la venta')).toBeInTheDocument()
  })

  it('offers to retry when the catalog fails, and says so when a code cannot be looked up', async () => {
    const working = fakeCatalog()
    const searchCatalog = vi.fn(working).mockRejectedValueOnce(new Error('down'))
    const { user } = renderPos(searchCatalog)

    expect(await screen.findByRole('heading', { name: 'No se pudieron cargar los artículos' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))
    expect(await catalog().findByRole('button', { name: /Gaseosa cola/ })).toBeInTheDocument()

    searchCatalog.mockRejectedValueOnce(new Error('down'))
    await user.type(search(), 'BEB-001{Enter}')
    expect(await screen.findByRole('status')).toHaveTextContent('Ocurrió un error inesperado')
  })

  it('opens the cart in a sheet that always shows the total', async () => {
    const { user } = renderPos()
    await user.click(await catalog().findByRole('button', { name: /Gaseosa cola/ }))

    await user.click(screen.getByRole('button', { name: `Ver venta: 1 artículo, total ${formatMoney(320_000)}` }))

    const sheet = within(await screen.findByRole('dialog', { name: 'Venta actual' }))
    expect(sheet.getByRole('textbox', { name: 'Cantidad de Gaseosa cola' })).toHaveValue('1')
    await user.click(sheet.getByRole('button', { name: 'Cerrar' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())

    await user.click(screen.getByRole('button', { name: /Ver venta/ }))
    await screen.findByRole('dialog', { name: 'Venta actual' })
    await user.keyboard('{F2}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(search()).toHaveFocus()
  })
})

describe('PosPage customer and seller', () => {
  const parties = () => within(screen.getByRole('region', { name: 'Venta actual' }))

  it('sells to Consumidor Final until another customer is chosen with F6', async () => {
    const { user } = renderPos()
    expect(parties().getByRole('button', { name: 'Cliente: Consumidor Final. Cambiar (F6)' })).toBeInTheDocument()

    await user.keyboard('{F6}')
    const picker = within(await screen.findByRole('dialog', { name: 'Elegir cliente' }))
    const results = within(picker.getByRole('list', { name: 'Resultados' }))
    expect(await results.findByRole('button', { name: /Consumidor Final/ })).toHaveAttribute('aria-current', 'true')
    expect(results.getByRole('button', { name: /Ana Pérez/ })).toHaveTextContent('DNI 28.456.789')
    expect(results.getByRole('button', { name: /Kiosco La Esquina SRL/ })).toHaveTextContent('CUIT 30-71234567-8')

    await user.keyboard('esquina{Enter}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(parties().getByRole('button', { name: 'Cliente: Kiosco La Esquina SRL. Cambiar (F6)' })).toHaveTextContent('CUIT 30-71234567-8')
    expect(JSON.parse(window.localStorage.getItem(CART_STORAGE_KEY)!)).toMatchObject({ customer: { id: 'esquina' } })

    await user.click(parties().getByRole('button', { name: /^Cliente:/ }))
    const again = within(await screen.findByRole('dialog', { name: 'Elegir cliente' }))
    await user.click(await again.findByRole('button', { name: /Consumidor Final/ }))
    expect(parties().getByRole('button', { name: 'Cliente: Consumidor Final. Cambiar (F6)' })).toBeInTheDocument()
  })

  it('walks the customers with the arrows and says when nobody matches', async () => {
    const { user } = renderPos()

    await user.keyboard('{F6}')
    const picker = within(await screen.findByRole('dialog', { name: 'Elegir cliente' }))
    await picker.findByRole('button', { name: /Ana Pérez/ })
    await user.keyboard('{ArrowDown}')
    expect(picker.getByRole('button', { name: /Ana Pérez/ })).toHaveFocus()
    await user.keyboard('{ArrowDown}')
    expect(picker.getByRole('button', { name: /Consumidor Final/ })).toHaveFocus()
    await user.keyboard('{ArrowUp}{Enter}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(parties().getByRole('button', { name: /^Cliente: Ana Pérez/ })).toBeInTheDocument()

    await user.keyboard('{F6}')
    const second = within(await screen.findByRole('dialog', { name: 'Elegir cliente' }))
    await user.type(second.getByRole('searchbox', { name: 'Buscar cliente' }), 'nadie')
    expect(await second.findByText('No hay clientes que coincidan. Probar con otra búsqueda.')).toBeInTheDocument()
    await user.keyboard('{Enter}')
    expect(await second.findByRole('status')).toHaveTextContent(/No hay resultados para .nadie./)
    await user.keyboard('{Escape}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(parties().getByRole('button', { name: /^Cliente: Ana Pérez/ })).toBeInTheDocument()
  })

  it('picks the seller by code with F4 and keeps it for the next sales', async () => {
    const { user, unmount } = renderPos()
    expect(parties().getByRole('button', { name: 'Vendedor: Sin vendedor. Cambiar (F4)' })).toBeInTheDocument()

    await user.keyboard('{F4}')
    const picker = within(await screen.findByRole('dialog', { name: 'Elegir vendedor' }))
    expect(picker.getByRole('button', { name: /Sin vendedor/ })).toHaveAttribute('aria-current', 'true')
    await user.keyboard('lf{Enter}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(parties().getByRole('button', { name: 'Vendedor: Lucía Fernández. Cambiar (F4)' })).toHaveTextContent('LF')

    await user.click(await catalog().findByRole('button', { name: /Gaseosa cola/ }))
    await user.click(parties().getByRole('button', { name: /^Cliente:/ }))
    await user.click(await within(await screen.findByRole('dialog')).findByRole('button', { name: /Ana Pérez/ }))
    await user.click(cart().getByRole('button', { name: 'Vaciar' }))
    expect(parties().getByRole('button', { name: /^Cliente: Consumidor Final/ })).toBeInTheDocument()
    expect(parties().getByRole('button', { name: /^Vendedor: Lucía Fernández/ })).toBeInTheDocument()
    await user.click(await screen.findByRole('button', { name: 'Deshacer' }))
    expect(parties().getByRole('button', { name: /^Cliente: Ana Pérez/ })).toBeInTheDocument()
    expect(cart().getByText('Gaseosa cola')).toBeInTheDocument()

    unmount()
    renderPos()
    expect(parties().getByRole('button', { name: /^Vendedor: Lucía Fernández/ })).toBeInTheDocument()
    expect(JSON.parse(window.localStorage.getItem(SELLER_STORAGE_KEY)!)).toEqual(SELLERS[1])

    await user.keyboard('{F4}')
    await user.click(await within(await screen.findByRole('dialog', { name: 'Elegir vendedor' })).findByRole('button', { name: /Sin vendedor/ }))
    expect(parties().getByRole('button', { name: 'Vendedor: Sin vendedor. Cambiar (F4)' })).toBeInTheDocument()
    expect(window.localStorage.getItem(SELLER_STORAGE_KEY)).toBeNull()
  })

  it('picks the first seller that matches a name and offers to retry a failed search', async () => {
    const failing = vi.fn(async () => {
      throw new Error('down')
    })
    const user = userEvent.setup()
    renderWithProviders(
      <PosPage
        categories={CATEGORIES}
        searchCatalog={fakeCatalog()}
        searchCustomers={failing}
        searchSellers={fakeParties(SELLERS, (seller) => [seller.name, seller.code])}
      />,
    )

    await user.keyboard('{F4}')
    await user.type(await screen.findByRole('searchbox', { name: 'Buscar vendedor' }), 'maría{Enter}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(parties().getByRole('button', { name: /^Vendedor: María Gómez/ })).toBeInTheDocument()

    await user.keyboard('{F6}')
    const picker = within(await screen.findByRole('dialog', { name: 'Elegir cliente' }))
    expect(await picker.findByText('No se pudo buscar. Revisar la conexión.')).toBeInTheDocument()
    expect(picker.getByRole('button', { name: 'Reintentar' })).toBeInTheDocument()
    await user.keyboard('{Enter}')
    expect(await picker.findByRole('status')).toHaveTextContent('Ocurrió un error inesperado')
  })

  it('ignores a stored seller or customer that is damaged', () => {
    window.localStorage.setItem(SELLER_STORAGE_KEY, '{"id":1}')
    window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify({ version: 1, lines: [], customer: { id: 'x' } }))
    renderPos()
    expect(parties().getByRole('button', { name: /^Vendedor: Sin vendedor/ })).toBeInTheDocument()
    expect(parties().getByRole('button', { name: /^Cliente: Consumidor Final/ })).toBeInTheDocument()
  })
})
