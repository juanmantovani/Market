import { createBrowserRouter, type RouteObject } from 'react-router'
import AppLayout from './AppLayout'
import {
  CategoriesPage,
  CustomersPage,
  DashboardPage,
  ItemsRoute,
  NotFoundPage,
  SellersPage,
  SuppliersPage,
} from './lazy-pages'
import type { RouteHandle } from './use-page-title'

const title = (text: string): RouteHandle => ({ title: text })

export const routes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { index: true, element: <DashboardPage />, handle: title('Inicio') },
      { path: 'articulos', element: <ItemsRoute />, handle: title('Artículos') },
      { path: 'categorias', element: <CategoriesPage />, handle: title('Categorías') },
      { path: 'clientes', element: <CustomersPage />, handle: title('Clientes') },
      { path: 'proveedores', element: <SuppliersPage />, handle: title('Proveedores') },
      { path: 'vendedores', element: <SellersPage />, handle: title('Vendedores') },
      { path: '*', element: <NotFoundPage />, handle: title('Página no encontrada') },
    ],
  },
]

export const router = createBrowserRouter(routes)
