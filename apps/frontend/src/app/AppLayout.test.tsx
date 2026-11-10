import { render, screen, within } from '@testing-library/react'
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

function stubHealth(ok: boolean) {
  vi.stubGlobal(
    'fetch',
    vi.fn(() =>
      ok
        ? Promise.resolve(new Response(JSON.stringify({ status: 'ok' }), { status: 200 }))
        : Promise.reject(new TypeError('Failed to fetch')),
    ),
  )
}

beforeEach(() => {
  window.localStorage.clear()
  document.documentElement.classList.remove('dark')
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('AppLayout', () => {
  it('shows the home page with the main navigation', async () => {
    stubHealth(true)
    renderApp()

    expect(screen.getByRole('heading', { name: 'Inicio' })).toBeInTheDocument()
    const nav = screen.getByRole('navigation', { name: 'Principal' })
    expect(within(nav).getByRole('link', { name: 'Inicio' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('navigation', { name: 'Navegación móvil' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Saltar al contenido' })).toHaveAttribute('href', '#main')
  })

  it('reports when the server is online', async () => {
    stubHealth(true)
    renderApp()

    expect(await screen.findByText('En línea')).toBeInTheDocument()
  })

  it('reports when the server cannot be reached', async () => {
    stubHealth(false)
    renderApp()

    expect(screen.getByText('Conectando…')).toBeInTheDocument()
    expect(await screen.findByText('Sin conexión')).toBeInTheDocument()
  })

  it('collapses the sidebar and remembers the choice', async () => {
    stubHealth(true)
    const user = userEvent.setup()
    renderApp()

    await user.click(screen.getByRole('button', { name: 'Contraer menú' }))

    const expand = screen.getByRole('button', { name: 'Expandir menú' })
    expect(expand).toHaveAttribute('aria-expanded', 'false')
    expect(window.localStorage.getItem('market.sidebar-collapsed')).toBe('true')
    const nav = screen.getByRole('navigation', { name: 'Principal' })
    expect(within(nav).getByRole('link', { name: 'Inicio' })).toBeInTheDocument()

    await user.click(expand)
    expect(screen.getByRole('button', { name: 'Contraer menú' })).toHaveAttribute('aria-expanded', 'true')
  })

  it('cycles the theme between light, dark and automatic', async () => {
    stubHealth(true)
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

  it('shows a friendly page for unknown addresses', async () => {
    stubHealth(true)
    const user = userEvent.setup()
    const router = renderApp('/no-existe')

    expect(screen.getByRole('heading', { name: 'No encontramos esta página' })).toBeInTheDocument()

    await user.click(screen.getByRole('link', { name: 'Volver al inicio' }))
    expect(router.state.location.pathname).toBe('/')
    expect(screen.getByRole('heading', { name: 'Inicio' })).toBeInTheDocument()
  })
})
