import { render, screen } from '@testing-library/react'
import { QueryClient } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import Providers from './Providers'
import { routes } from './router'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('routes', () => {
  it('shows a page skeleton inside the layout until the page code arrives', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(JSON.stringify({ items: [], total: 0, page: 1, limit: 25 }), {
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <Providers queryClient={queryClient}>
        <RouterProvider router={createMemoryRouter(routes, { initialEntries: ['/categorias'] })} />
      </Providers>,
    )

    expect(screen.getByRole('status', { name: 'Cargando la página' })).toBeInTheDocument()
    expect(screen.getByRole('navigation', { name: 'Principal' })).toBeInTheDocument()

    expect(await screen.findByRole('heading', { name: 'Categorías', level: 1 })).toBeInTheDocument()
    expect(screen.queryByRole('status', { name: 'Cargando la página' })).not.toBeInTheDocument()
  })
})
