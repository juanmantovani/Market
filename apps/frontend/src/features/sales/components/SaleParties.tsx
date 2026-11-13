import { ChevronRight, UserRound, UserRoundCheck, type LucideIcon } from 'lucide-react'
import { documentText } from '../lib/parties'
import type { CustomerOption, SellerOption } from '../types'

type SalePartiesProps = {
  customer: CustomerOption | null
  seller: SellerOption | null
  onChooseCustomer: () => void
  onChooseSeller: () => void
}

// Who the sale is for and who attends it, each a button that opens its
// picker (F6 and F4 from anywhere on the screen).
export function SaleParties({ customer, seller, onChooseCustomer, onChooseSeller }: SalePartiesProps) {
  return (
    <div className="flex flex-col border-b px-2 py-1">
      <PartyButton
        icon={UserRound}
        role="Cliente"
        name={customer?.name ?? 'Consumidor Final'}
        detail={customer ? documentText(customer) : null}
        shortcut="F6"
        onClick={onChooseCustomer}
      />
      <PartyButton
        icon={UserRoundCheck}
        role="Vendedor"
        name={seller ? seller.name : 'Sin vendedor'}
        detail={seller?.code ?? null}
        shortcut="F4"
        onClick={onChooseSeller}
      />
    </div>
  )
}

type PartyButtonProps = {
  icon: LucideIcon
  role: string
  name: string
  detail: string | null
  shortcut: string
  onClick: () => void
}

function PartyButton({ icon: Icon, role, name, detail, shortcut, onClick }: PartyButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={`${role}: ${name}. Cambiar (${shortcut})`}
      className="flex min-h-11 w-full items-center gap-3 rounded-lg px-2 py-1.5 text-left transition-colors hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none"
    >
      <Icon className="size-4 shrink-0 text-muted-foreground" aria-hidden />
      <span className="flex min-w-0 flex-1 flex-col">
        <span className="text-xs text-muted-foreground">{role}</span>
        <span className="truncate text-sm font-medium">
          {name}
          {detail && <span className="font-normal text-muted-foreground"> · {detail}</span>}
        </span>
      </span>
      <kbd className="hidden rounded border bg-muted px-1.5 py-0.5 font-sans text-xs text-muted-foreground lg:inline">{shortcut}</kbd>
      <ChevronRight className="size-4 shrink-0 text-muted-foreground" aria-hidden />
    </button>
  )
}
