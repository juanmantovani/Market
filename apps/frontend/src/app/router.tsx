import { createBrowserRouter, type RouteObject } from 'react-router'
import CategoriesPage from '../features/categories/pages/CategoriesPage'
import CustomersPage from '../features/customers/pages/CustomersPage'
import DashboardPage from '../features/dashboard/pages/DashboardPage'
import ItemsRoute from './ItemsRoute'
import AppLayout from './AppLayout'
import NotFoundPage from './NotFoundPage'

export const routes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'articulos', element: <ItemsRoute /> },
      { path: 'categorias', element: <CategoriesPage /> },
      { path: 'clientes', element: <CustomersPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
]

export const router = createBrowserRouter(routes)
