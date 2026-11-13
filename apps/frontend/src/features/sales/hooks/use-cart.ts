import { useEffect, useState } from 'react'
import { addItem, cartTotal, removeLine, restoreLine, setQuantity, type RemovedLine } from '../lib/cart'
import { storedCartSchema, type CartLine, type CatalogItem } from '../types'

export const CART_STORAGE_KEY = 'market.pos.cart'

// The cart lives only in the browser (the backend receives the finished
// sale): a reload or a closed tab must not lose it. Storage may be full or
// blocked, and an old or tampered value is ignored.
function readCart(): CartLine[] {
  try {
    const raw = window.localStorage.getItem(CART_STORAGE_KEY)
    if (!raw) return []
    const stored = storedCartSchema.safeParse(JSON.parse(raw))
    return stored.success ? stored.data.lines : []
  } catch {
    return []
  }
}

function writeCart(lines: CartLine[]) {
  try {
    if (lines.length === 0) {
      window.localStorage.removeItem(CART_STORAGE_KEY)
    } else {
      window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify({ version: 1, lines }))
    }
  } catch {
    // Without storage the cart still works; it just won't survive a reload.
  }
}

export function useCart() {
  const [lines, setLines] = useState(readCart)

  useEffect(() => {
    writeCart(lines)
  }, [lines])

  return {
    lines,
    total: cartTotal(lines),
    add: (item: CatalogItem, quantity: number) => setLines((current) => addItem(current, item, quantity)),
    setQuantity: (itemId: string, quantity: number) => setLines((current) => setQuantity(current, itemId, quantity)),
    // Returns what was removed, so it can be put back with restore.
    remove: (itemId: string): RemovedLine | null => {
      const { lines: remaining, removed } = removeLine(lines, itemId)
      setLines(remaining)
      return removed
    },
    restore: (removed: RemovedLine) => setLines((current) => restoreLine(current, removed)),
    // Returns the lines it cleared, so they can be put back with restoreAll.
    clear: (): CartLine[] => {
      const cleared = lines
      setLines([])
      return cleared
    },
    restoreAll: (cleared: CartLine[]) =>
      setLines((current) => cleared.reduce((restored, line, index) => restoreLine(restored, { line, index }), current)),
  }
}
