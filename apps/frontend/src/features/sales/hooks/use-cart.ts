import { useEffect, useState } from 'react'
import { addItem, cartTotal, removeLine, restoreLine, setQuantity, type RemovedLine } from '../lib/cart'
import { storedCartSchema, type CartLine, type CatalogItem, type CustomerOption } from '../types'

export const CART_STORAGE_KEY = 'market.pos.cart'

// The sale being built: its lines and its customer (null is Consumidor
// Final, the backend's default).
type Cart = { lines: CartLine[]; customer: CustomerOption | null }

const EMPTY: Cart = { lines: [], customer: null }

// The cart lives only in the browser (the backend receives the finished
// sale): a reload or a closed tab must not lose it. Storage may be full or
// blocked, and an old or tampered value is ignored.
function readCart(): Cart {
  try {
    const raw = window.localStorage.getItem(CART_STORAGE_KEY)
    if (!raw) return EMPTY
    const stored = storedCartSchema.safeParse(JSON.parse(raw))
    return stored.success ? { lines: stored.data.lines, customer: stored.data.customer ?? null } : EMPTY
  } catch {
    return EMPTY
  }
}

function writeCart(cart: Cart) {
  try {
    if (cart.lines.length === 0 && cart.customer === null) {
      window.localStorage.removeItem(CART_STORAGE_KEY)
    } else {
      window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify({ version: 1, ...cart }))
    }
  } catch {
    // Without storage the cart still works; it just won't survive a reload.
  }
}

export function useCart() {
  const [cart, setCart] = useState(readCart)
  const { lines, customer } = cart

  useEffect(() => {
    writeCart(cart)
  }, [cart])

  function updateLines(change: (current: CartLine[]) => CartLine[]) {
    setCart((current) => ({ ...current, lines: change(current.lines) }))
  }

  return {
    lines,
    customer,
    total: cartTotal(lines),
    add: (item: CatalogItem, quantity: number) => updateLines((current) => addItem(current, item, quantity)),
    setQuantity: (itemId: string, quantity: number) => updateLines((current) => setQuantity(current, itemId, quantity)),
    // Consumidor Final is the default, so choosing it clears the choice.
    setCustomer: (next: CustomerOption | null) => setCart((current) => ({ ...current, customer: next?.system ? null : next })),
    // Returns what was removed, so it can be put back with restore.
    remove: (itemId: string): RemovedLine | null => {
      const { lines: remaining, removed } = removeLine(lines, itemId)
      updateLines(() => remaining)
      return removed
    },
    restore: (removed: RemovedLine) => updateLines((current) => restoreLine(current, removed)),
    // Starts a new sale: no lines and Consumidor Final. Returns what it
    // cleared, so it can be put back with restoreAll.
    clear: (): Cart => {
      const cleared = cart
      setCart(EMPTY)
      return cleared
    },
    restoreAll: (cleared: Cart) =>
      setCart((current) => ({
        lines: cleared.lines.reduce((restored, line, index) => restoreLine(restored, { line, index }), current.lines),
        customer: current.customer ?? cleared.customer,
      })),
  }
}
