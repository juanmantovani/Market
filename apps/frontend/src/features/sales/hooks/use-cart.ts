import { useEffect, useMemo, useState } from 'react'
import { addItem, removeLine, restoreLine, setQuantity, type RemovedLine } from '../lib/cart'
import { priceCart } from '../lib/discounts'
import { storedCartSchema, type CartLine, type CatalogItem, type CustomerOption, type Discount } from '../types'

export const CART_STORAGE_KEY = 'market.pos.cart'
const STORAGE_VERSION = 2

// The sale being built: its lines, its customer (null is Consumidor Final,
// the backend's default) and its general discount.
type Cart = { lines: CartLine[]; customer: CustomerOption | null; cartDiscount: Discount | null }

const EMPTY: Cart = { lines: [], customer: null, cartDiscount: null }

// The cart lives only in the browser (the backend receives the finished
// sale): a reload or a closed tab must not lose it. Storage may be full or
// blocked, and an old or tampered value is ignored.
function readCart(): Cart {
  try {
    const raw = window.localStorage.getItem(CART_STORAGE_KEY)
    if (!raw) return EMPTY
    const stored = storedCartSchema.safeParse(JSON.parse(raw))
    return stored.success
      ? { lines: stored.data.lines, customer: stored.data.customer ?? null, cartDiscount: stored.data.cartDiscount ?? null }
      : EMPTY
  } catch {
    return EMPTY
  }
}

function writeCart(cart: Cart) {
  try {
    if (cart.lines.length === 0 && cart.customer === null && cart.cartDiscount === null) {
      window.localStorage.removeItem(CART_STORAGE_KEY)
    } else {
      window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify({ version: STORAGE_VERSION, ...cart }))
    }
  } catch {
    // Without storage the cart still works; it just won't survive a reload.
  }
}

export function useCart() {
  const [cart, setCart] = useState(readCart)
  const { lines, customer, cartDiscount } = cart
  const pricing = useMemo(() => priceCart(lines, cartDiscount), [lines, cartDiscount])

  useEffect(() => {
    writeCart(cart)
  }, [cart])

  function updateLines(change: (current: CartLine[]) => CartLine[]) {
    setCart((current) => ({ ...current, lines: change(current.lines) }))
  }

  return {
    lines,
    customer,
    cartDiscount,
    pricing,
    total: pricing.total,
    add: (item: CatalogItem, quantity: number) => updateLines((current) => addItem(current, item, quantity)),
    setQuantity: (itemId: string, quantity: number) => updateLines((current) => setQuantity(current, itemId, quantity)),
    // null takes the line's discount off.
    setLineDiscount: (itemId: string, discount: Discount | null) =>
      updateLines((current) => current.map((line) => (line.itemId === itemId ? { ...line, discount } : line))),
    setCartDiscount: (discount: Discount | null) => setCart((current) => ({ ...current, cartDiscount: discount })),
    // Consumidor Final is the default, so choosing it clears the choice.
    setCustomer: (next: CustomerOption | null) => setCart((current) => ({ ...current, customer: next?.system ? null : next })),
    // Returns what was removed, so it can be put back with restore.
    remove: (itemId: string): RemovedLine | null => {
      const { lines: remaining, removed } = removeLine(lines, itemId)
      updateLines(() => remaining)
      return removed
    },
    restore: (removed: RemovedLine) => updateLines((current) => restoreLine(current, removed)),
    // Starts a new sale: no lines, no discount and Consumidor Final. Returns
    // what it cleared, so it can be put back with restoreAll.
    clear: (): Cart => {
      const cleared = cart
      setCart(EMPTY)
      return cleared
    },
    restoreAll: (cleared: Cart) =>
      setCart((current) => ({
        lines: cleared.lines.reduce((restored, line, index) => restoreLine(restored, { line, index }), current.lines),
        customer: current.customer ?? cleared.customer,
        cartDiscount: current.cartDiscount ?? cleared.cartDiscount,
      })),
  }
}
