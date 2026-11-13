import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Providers from './Providers'
import { routes } from './router'

function renderApp(path = '/') {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createMemoryRouter(routes, { initialEntries: [path] })

  render(
    <Providers queryClient={queryClient}>
      <RouterProvider router={router} />
    </Providers>,
  )

  return router
}

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn())
  window.localStorage.clear()
  document.documentElement.classList.remove('dark')
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('AppLayout', () => {
  it('opens on the point of sale with the main navigation', async () => {
    renderApp()

    expect(await screen.findByRole('heading', { name: 'Punto de venta' })).toBeInTheDocument()
    const nav = screen.getByRole('navigation', { name: 'Principal' })
    expect(within(nav).getByRole('link', { name: 'Vender' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('navigation', { name: 'Navegación móvil' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Saltar al contenido' })).toHaveAttribute('href', '#main')
  })

  it('renders the layout without calling the server', async () => {
    renderApp('/no-existe')

    expect(await screen.findByRole('heading', { name: 'No se encontró esta página' })).toBeInTheDocument()
    expect(fetch).not.toHaveBeenCalled()
  })

  it('collapses the sidebar and remembers the choice', async () => {
    const user = userEvent.setup()
    renderApp()

    await user.click(screen.getByRole('button', { name: 'Contraer menú' }))

    const expand = screen.getByRole('button', { name: 'Expandir menú' })
    expect(expand).toHaveAttribute('aria-expanded', 'false')
    expect(window.localStorage.getItem('market.sidebar-collapsed')).toBe('true')
    const nav = screen.getByRole('navigation', { name: 'Principal' })
    expect(within(nav).getByRole('link', { name: 'Vender' })).toBeInTheDocument()

    await user.click(expand)
    expect(screen.getByRole('button', { name: 'Contraer menú' })).toHaveAttribute('aria-expanded', 'true')
  })

  it('cycles the theme between light, dark and automatic', async () => {
    const user = userEvent.setup()
    renderApp()

    await user.click(screen.getByRole('button', { name: /Tema automático\. Cambiar a claro/ }))
    expect(document.documentElement).not.toHaveClass('dark')

    await user.click(screen.getByRole('button', { name: /Tema claro\. Cambiar a oscuro/ }))
    expect(document.documentElement).toHaveClass('dark')
    expect(window.localStorage.getItem('market.theme')).toBe('dark')

    await user.click(screen.getByRole('button', { name: /Tema oscuro\. Cambiar a automático/ }))
    expect(document.documentElement).not.toHaveClass('dark')
  })

  it('opens the customers section from the navigation', async () => {
    // A new response per call: the point of sale reads the catalog first.
    vi.mocked(fetch).mockImplementation(async () =>
      new Response(JSON.stringify({ items: [], total: 0, page: 1, limit: 25 }), {
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    const user = userEvent.setup()
    const router = renderApp()

    await user.click(within(screen.getByRole('navigation', { name: 'Principal' })).getByRole('link', { name: 'Clientes' }))

    expect(router.state.location.pathname).toBe('/clientes')
    await waitFor(() => expect(document.title).toBe('Clientes · Market'))
    expect(await screen.findByRole('heading', { name: 'Clientes', level: 1 })).toBeInTheDocument()
    expect(await screen.findByRole('heading', { name: 'Todavía no hay clientes' })).toBeInTheDocument()
  })

  it('keeps the phone bar short and opens the other sections from Más', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ items: [], total: 0, page: 1, limit: 25 }), {
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    const user = userEvent.setup()
    const router = renderApp()
    const mobileNav = within(screen.getByRole('navigation', { name: 'Navegación móvil' }))

    expect(mobileNav.getAllByRole('link').map((link) => link.textContent)).toEqual(['Vender', 'Artículos', 'Clientes'])
    const more = mobileNav.getByRole('button', { name: 'Más' })
    await user.click(more)

    const sections = within(await screen.findByRole('dialog', { name: 'Más secciones' }))
    expect(sections.getAllByRole('link').map((link) => link.textContent)).toEqual(['Categorías', 'Stock', 'Proveedores', 'Vendedores'])
    await user.click(sections.getByRole('link', { name: 'Proveedores' }))

    expect(router.state.location.pathname).toBe('/proveedores')
    expect(await screen.findByRole('heading', { name: 'Proveedores', level: 1 })).toBeInTheDocument()
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Más secciones' })).not.toBeInTheDocument())
  })

  it('shows a friendly page for unknown addresses', async () => {
    const user = userEvent.setup()
    const router = renderApp('/no-existe')

    expect(await screen.findByRole('heading', { name: 'No se encontró esta página' })).toBeInTheDocument()
    expect(document.title).toBe('Página no encontrada · Market')

    await user.click(screen.getByRole('link', { name: 'Volver al inicio' }))
    expect(router.state.location.pathname).toBe('/')
    await waitFor(() => expect(document.title).toBe('Punto de venta · Market'))
    expect(await screen.findByRole('heading', { name: 'Punto de venta' })).toBeInTheDocument()
  })
})
