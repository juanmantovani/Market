import type { ReactNode } from 'react'
import { ShoppingCart, X } from 'lucide-react'
import { formatMoney } from '../../../shared/lib/money'
import { Button } from '../../../shared/ui/button'
import { Dialog, DialogClose, DialogContent, DialogTitle } from '../../../shared/ui/dialog'

type CartSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  lineCount: number
  total: number
  // The cart itself; it gets the close button for its header.
  children: (closeButton: ReactNode) => ReactNode
}

// On a phone the catalog takes the whole screen: the cart is a sheet that
// slides up from a bar that always shows the total.
export function CartSheet({ open, onOpenChange, lineCount, total, children }: CartSheetProps) {
  const items = `${lineCount} ${lineCount === 1 ? 'artículo' : 'artículos'}`
  const closeButton = (
    <DialogClose
      render={<Button type="button" variant="ghost" size="icon" className="size-11 text-muted-foreground" aria-label="Cerrar" />}
    >
      <X aria-hidden />
    </DialogClose>
  )

  return (
    <div className="fixed inset-x-0 bottom-[calc(3.5rem+env(safe-area-inset-bottom))] z-10 border-t bg-background/95 px-4 py-2 backdrop-blur md:bottom-0 lg:hidden">
      <Button
        type="button"
        size="lg"
        className="h-12 w-full justify-between px-4 text-base"
        aria-label={`Ver venta: ${items}, total ${formatMoney(total)}`}
        onClick={() => onOpenChange(true)}
      >
        <span className="flex items-center gap-2">
          <ShoppingCart aria-hidden />
          Ver venta
          <span className="font-normal opacity-90">· {items}</span>
        </span>
        <span className="tabular-nums">{formatMoney(total)}</span>
      </Button>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="top-auto bottom-0 left-0 flex max-h-[90dvh] w-full max-w-none translate-x-0 translate-y-0 flex-col overflow-hidden rounded-b-none border-x-0 border-b-0 p-0 data-open:zoom-in-100 data-open:slide-in-from-bottom-8">
          <DialogTitle className="sr-only">Venta actual</DialogTitle>
          {children(closeButton)}
        </DialogContent>
      </Dialog>
    </div>
  )
}
