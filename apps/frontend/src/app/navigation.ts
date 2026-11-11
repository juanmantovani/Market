import { House, Package, Tags, Users, type LucideIcon } from 'lucide-react'

export type NavItem = {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
}

// Each phase adds its section here once its route exists.
export const NAVIGATION: readonly NavItem[] = [
  { to: '/', label: 'Inicio', icon: House, end: true },
  { to: '/articulos', label: 'Artículos', icon: Package },
  { to: '/categorias', label: 'Categorías', icon: Tags },
  { to: '/clientes', label: 'Clientes', icon: Users },
]
