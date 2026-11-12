import { lazy } from 'react'

export const CategoriesPage = lazy(() => import('../features/categories/pages/CategoriesPage'))
export const CustomersPage = lazy(() => import('../features/customers/pages/CustomersPage'))
export const DashboardPage = lazy(() => import('../features/dashboard/pages/DashboardPage'))
export const SellersPage = lazy(() => import('../features/sellers/pages/SellersPage'))
export const SuppliersPage = lazy(() => import('../features/suppliers/pages/SuppliersPage'))
export const ItemsRoute = lazy(() => import('./ItemsRoute'))
export const StockRoute = lazy(() => import('./StockRoute'))
export const NotFoundPage = lazy(() => import('./NotFoundPage'))
