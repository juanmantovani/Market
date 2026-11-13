import { useEffect, useRef, useState, type KeyboardEvent, type ReactNode } from 'react'
import { messageFromError } from '../../../shared/api/client'
import { useDebouncedValue } from '../../../shared/hooks/use-debounced-value'
import { formatMoney } from '../../../shared/lib/money'
import { Button } from '../../../shared/ui/button'
import { PageHeader } from '../../../shared/ui/page-header'
import { SearchField } from '../../../shared/ui/search-field'
import { useToast } from '../../../shared/ui/use-toast'
import { CartPanel } from '../components/CartPanel'
import { CartSheet } from '../components/CartSheet'
import { CatalogGrid } from '../components/CatalogGrid'
import { CategoryFilter } from '../components/CategoryFilter'
import { CheckoutDialog } from '../components/CheckoutDialog'
import { DiscountDialog } from '../components/DiscountDialog'
import { CustomerPicker, SellerPicker } from '../components/PartyPickers'
import { QuantityDialog } from '../components/QuantityDialog'
import { SaleParties } from '../components/SaleParties'
import { useCart } from '../hooks/use-cart'
import { useSeller } from '../hooks/use-seller'
import { CATALOG_LIMIT, useCatalog, useCodeLookup } from '../hooks/use-catalog'
import { useShortcuts } from '../hooks/use-shortcuts'
import { addItem } from '../lib/cart'
import { priceCart } from '../lib/discounts'
import type { CatalogItem, CategoryOption, Discount, SearchCatalog, SearchCustomers, SearchSellers } from '../types'

type PosPageProps = {
  categories: CategoryOption[]
  searchCatalog: SearchCatalog
  searchCustomers: SearchCustomers
  searchSellers: SearchSellers
}

// Phones and tablets type with an on-screen keyboard: focusing the search
// on arrival would cover half the catalog.
function hasFinePointer() {
  return typeof window.matchMedia === 'function' && window.matchMedia('(pointer: fine)').matches
}

export default function PosPage({ categories, searchCatalog, searchCustomers, searchSellers }: PosPageProps) {
  const cart = useCart()
  const [seller, setSeller] = useSeller()
  const [picker, setPicker] = useState<'customer' | 'seller' | null>(null)
  const toast = useToast()
  const [search, setSearch] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [notice, setNotice] = useState<string | null>(null)
  const [announcement, setAnnouncement] = useState('')
  const [weighing, setWeighing] = useState<CatalogItem | null>(null)
  const [weighOpen, setWeighOpen] = useState(false)
  const [sheetOpen, setSheetOpen] = useState(false)
  const [checkoutOpen, setCheckoutOpen] = useState(false)
  // Which discount is being edited: a line's (by item) or the general one
  // (null). It stays set while the dialog closes, so the closing animation
  // keeps its content.
  const [discountTarget, setDiscountTarget] = useState<{ itemId: string | null } | null>(null)
  const [discountOpen, setDiscountOpen] = useState(false)
  const searchRef = useRef<HTMLInputElement>(null)
  const debouncedSearch = useDebouncedValue(search.trim(), 250)
  const catalog = useCatalog(searchCatalog, { search: debouncedSearch, categoryId, page: 1, limit: CATALOG_LIMIT })
  const lookup = useCodeLookup(searchCatalog)

  useShortcuts({
    F2: () => {
      setSheetOpen(false)
      searchRef.current?.focus()
      searchRef.current?.select()
    },
    F4: () => {
      if (!checkoutOpen) setPicker('seller')
    },
    F6: () => {
      if (!checkoutOpen) setPicker('customer')
    },
    F9: () => {
      if (!checkoutOpen && picker === null) startCheckout()
    },
  })

  useEffect(() => {
    if (hasFinePointer()) searchRef.current?.focus()
  }, [])

  function add(item: CatalogItem, quantity: number) {
    const total = priceCart(addItem(cart.lines, item, quantity), cart.cartDiscount).total
    cart.add(item, quantity)
    setAnnouncement(`Se agregó ${item.name}. Total: ${formatMoney(total)}`)
  }

  function pick(item: CatalogItem) {
    setNotice(null)
    if (item.unit === 'unit') {
      add(item, 1000)
    } else {
      setWeighing(item)
      setWeighOpen(true)
    }
  }

  // Enter adds the item whose exact code or barcode was typed (a barcode
  // scanner types the code and Enter), or the only one that matches. The
  // field empties right away, so the next scan doesn't land on this one;
  // the text comes back when there is nothing to add and nobody typed since.
  async function submitSearch() {
    const code = search.trim()
    if (!code) return
    setSearch('')
    setNotice(null)
    const showAgain = () => setSearch((current) => current || code)
    try {
      const result = await lookup(code)
      if (result.item) {
        pick(result.item)
      } else if (result.results === 0) {
        showAgain()
        setNotice(`No hay artículos con “${code}”. Probar con otra búsqueda.`)
      } else {
        showAgain()
        setNotice('Hay varios artículos que coinciden: elegir uno de la lista.')
      }
    } catch (error) {
      showAgain()
      setNotice(messageFromError(error))
    }
  }

  function handleSearchKey(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter') {
      event.preventDefault()
      void submitSearch()
    } else if (event.key === 'ArrowDown') {
      const first = document.querySelector<HTMLButtonElement>('[data-pos-catalog] button')
      if (first) {
        event.preventDefault()
        first.focus()
      }
    } else if (event.key === 'Escape' && search) {
      event.preventDefault()
      setSearch('')
      setNotice(null)
    }
  }

  function remove(itemId: string) {
    const removed = cart.remove(itemId)
    if (removed) toast.undoable(`Se quitó ${removed.line.name}`, () => cart.restore(removed))
  }

  function clear() {
    const cleared = cart.clear()
    if (cleared.lines.length > 0 || cleared.customer || cleared.cartDiscount) toast.undoable('Se vació la venta', () => cart.restoreAll(cleared))
  }

  function startCheckout() {
    if (cart.lines.length === 0) {
      setNotice('Agregar al menos un artículo antes de cobrar.')
      return
    }
    setNotice(null)
    setSheetOpen(false)
    setCheckoutOpen(true)
  }

  function editDiscount(itemId: string | null) {
    setDiscountTarget({ itemId })
    setDiscountOpen(true)
  }

  const discountingLine = discountTarget?.itemId != null
  const discountLineIndex = discountingLine ? cart.lines.findIndex((line) => line.itemId === discountTarget.itemId) : -1
  const discountLine = discountLineIndex >= 0 ? cart.lines[discountLineIndex] : null
  // A line's discount is taken from its amount; the general one from what
  // the lines add up to after their own discounts.
  const discountBase = discountingLine
    ? (cart.pricing.lines[discountLineIndex]?.amount ?? 0)
    : cart.pricing.lines.reduce((sum, line) => sum + line.amount - line.discountAmount, 0)

  function applyDiscount(discount: Discount | null) {
    if (discountTarget?.itemId) cart.setLineDiscount(discountTarget.itemId, discount)
    else cart.setCartDiscount(discount)
    setDiscountOpen(false)
  }

  function focusSearch() {
    searchRef.current?.focus()
  }

  const panel = (headerAction?: ReactNode, className?: string) => (
    <CartPanel
      lines={cart.lines}
      pricing={cart.pricing}
      onQuantityChange={cart.setQuantity}
      onLineDiscount={editDiscount}
      onCartDiscount={() => editDiscount(null)}
      onRemove={remove}
      onClear={clear}
      parties={
        <SaleParties
          customer={cart.customer}
          seller={seller}
          onChooseCustomer={() => setPicker('customer')}
          onChooseSeller={() => setPicker('seller')}
        />
      }
      headerAction={headerAction}
      className={className}
    >
      <Button type="button" size="lg" className="h-12 w-full text-base" disabled={cart.lines.length === 0} onClick={startCheckout}>
        Cobrar
        <kbd className="ml-1 hidden rounded border border-primary-foreground/40 px-1.5 text-xs font-medium lg:inline">F9</kbd>
      </Button>
    </CartPanel>
  )

  return (
    <div className="flex flex-col gap-4 pb-20 lg:pb-0">
      <PageHeader title="Punto de venta" />
      <p className="sr-only" aria-live="polite">
        {announcement}
      </p>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem] lg:items-start">
        <section aria-label="Catálogo" className="flex min-w-0 flex-col gap-3">
          <SearchField
            label="Buscar o escanear"
            hint="Por nombre, código o código de barras. Enter agrega el artículo; F2 vuelve a la búsqueda."
            value={search}
            onChange={(value) => {
              setSearch(value)
              setNotice(null)
            }}
            inputRef={searchRef}
            onKeyDown={handleSearchKey}
          />
          {notice && (
            <p role="status" className="rounded-lg bg-muted px-3 py-2 text-sm">
              {notice}
            </p>
          )}
          <CategoryFilter categories={categories} value={categoryId} onChange={setCategoryId} />
          <div data-pos-catalog>
            <CatalogGrid
              items={catalog.data?.items}
              isPending={catalog.isPending}
              isError={catalog.isError}
              onRetry={() => void catalog.refetch()}
              onPick={pick}
              onEscape={focusSearch}
            />
          </div>
        </section>

        {panel(undefined, 'hidden lg:sticky lg:top-20 lg:flex lg:max-h-[calc(100dvh-6rem)]')}
      </div>

      <CartSheet open={sheetOpen} onOpenChange={setSheetOpen} lineCount={cart.lines.length} total={cart.total}>
        {(closeButton) => panel(closeButton, 'min-h-0 flex-1 rounded-none border-0')}
      </CartSheet>

      <CheckoutDialog
        open={checkoutOpen}
        pricing={cart.pricing}
        customerId={cart.customer?.id ?? null}
        sellerId={seller?.id ?? null}
        onSold={() => {
          // The next sale starts empty and for Consumidor Final; the seller
          // stays.
          cart.clear()
          setNotice(null)
        }}
        onClose={() => {
          setCheckoutOpen(false)
          if (hasFinePointer()) focusSearch()
        }}
      />

      <CustomerPicker
        open={picker === 'customer'}
        onClose={() => setPicker(null)}
        search={searchCustomers}
        current={cart.customer}
        onPick={cart.setCustomer}
      />
      <SellerPicker open={picker === 'seller'} onClose={() => setPicker(null)} search={searchSellers} current={seller} onPick={setSeller} />

      <DiscountDialog
        open={discountOpen && (!discountingLine || discountLine !== null)}
        title={discountingLine ? `Descuento: ${discountLine?.name ?? ''}` : 'Descuento general'}
        base={discountBase}
        current={discountingLine ? (discountLine?.discount ?? null) : cart.cartDiscount}
        onApply={applyDiscount}
        onClose={() => setDiscountOpen(false)}
      />

      <QuantityDialog
        open={weighOpen}
        item={weighing}
        onAdd={(item, quantity) => {
          add(item, quantity)
          setWeighOpen(false)
        }}
        onClose={() => setWeighOpen(false)}
      />
    </div>
  )
}
