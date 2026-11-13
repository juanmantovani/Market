import { Package, Receipt, ShoppingCart, Tags, Truck, UserCheck, Users, type LucideIcon } from 'lucide-react'

export type NavItem = {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
  // The phone's bottom bar fits four buttons: these sections plus "Más",
  // which opens the rest.
  inMobileBar?: boolean
}

// Each phase adds its section here once its route exists.
export const NAVIGATION: readonly NavItem[] = [
  { to: '/', label: 'Vender', icon: ShoppingCart, end: true, inMobileBar: true },
  { to: '/ventas', label: 'Ventas', icon: Receipt, inMobileBar: true },
  { to: '/articulos', label: 'Artículos', icon: Package, inMobileBar: true },
  { to: '/categorias', label: 'Categorías', icon: Tags },
  { to: '/clientes', label: 'Clientes', icon: Users },
  { to: '/proveedores', label: 'Proveedores', icon: Truck },
  { to: '/vendedores', label: 'Vendedores', icon: UserCheck },
]
