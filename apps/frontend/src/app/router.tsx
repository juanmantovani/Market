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

export const routes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'articulos', element: <ItemsRoute /> },
      { path: 'categorias', element: <CategoriesPage /> },
      { path: 'clientes', element: <CustomersPage /> },
      { path: 'proveedores', element: <SuppliersPage /> },
      { path: 'vendedores', element: <SellersPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
]

export const router = createBrowserRouter(routes)
