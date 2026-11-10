import { createBrowserRouter, type RouteObject } from 'react-router'
import CategoriesPage from '../features/categories/pages/CategoriesPage'
import DashboardPage from '../features/dashboard/pages/DashboardPage'
import AppLayout from './AppLayout'
import NotFoundPage from './NotFoundPage'

export const routes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'categorias', element: <CategoriesPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
]

export const router = createBrowserRouter(routes)
