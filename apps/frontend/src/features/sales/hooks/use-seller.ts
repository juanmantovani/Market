import { useEffect, useState } from 'react'
import { sellerOptionSchema, type SellerOption } from '../types'

export const SELLER_STORAGE_KEY = 'market.pos.seller'

// The seller usually attends sale after sale at the same counter, so the
// choice outlives each sale (and a reload). null: no seller.
function readSeller(): SellerOption | null {
  try {
    const raw = window.localStorage.getItem(SELLER_STORAGE_KEY)
    if (!raw) return null
    const stored = sellerOptionSchema.safeParse(JSON.parse(raw))
    return stored.success ? stored.data : null
  } catch {
    return null
  }
}

function writeSeller(seller: SellerOption | null) {
  try {
    if (seller) {
      window.localStorage.setItem(SELLER_STORAGE_KEY, JSON.stringify(seller))
    } else {
      window.localStorage.removeItem(SELLER_STORAGE_KEY)
    }
  } catch {
    // Without storage the choice lasts until the page reloads.
  }
}

export function useSeller() {
  const [seller, setSeller] = useState(readSeller)

  useEffect(() => {
    writeSeller(seller)
  }, [seller])

  return [seller, setSeller] as const
}
