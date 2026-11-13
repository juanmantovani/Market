import { createPortal } from 'react-dom'
import { formatDateTime } from '../../../shared/lib/date'
import { formatMoney } from '../../../shared/lib/money'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { formatDiscount } from '../lib/discounts'
import { PAYMENT_METHOD_LABELS } from '../lib/payments'
import type { Sale } from '../types'

// The sale as a receipt, for a thermal or office printer. It lives apart
// from the dialog (a direct child of the body, marked data-print-area), so
// printing hides the app and the dialog and leaves only the ticket.
export function SaleTicket({ sale }: { sale: Sale }) {
  return createPortal(
    <article
      data-print-area
      aria-label={`Ticket de la venta ${sale.number}`}
      className="mx-auto hidden w-full max-w-[80mm] bg-background font-mono text-xs leading-snug print:block"
    >
      <header className="flex flex-col items-center gap-0.5 border-b border-dashed pb-2 text-center">
        <p className="text-sm font-semibold">Ticket de venta</p>
        <p>Documento no válido como factura</p>
      </header>

      <dl className="flex flex-col gap-0.5 border-b border-dashed py-2">
        <TicketRow label="Venta N.º" value={String(sale.number)} />
        <TicketRow label="Fecha" value={formatDateTime(sale.createdAt)} />
        <TicketRow label="Cliente" value={sale.customer.name} />
        {sale.seller && <TicketRow label="Vendedor" value={sale.seller.name} />}
      </dl>

      <ul aria-label="Artículos" className="flex flex-col gap-1.5 border-b border-dashed py-2">
        {sale.lines.map((line) => (
          <li key={line.itemId}>
            <p>{line.name}</p>
            <p className="flex justify-between gap-2">
              <span>
                {formatUnitQuantity(line.quantity, line.unit)} × {formatMoney(line.unitPrice)}
              </span>
              <span>{formatMoney(line.lineTotal + line.discountAmount + line.cartDiscountAmount)}</span>
            </p>
            {line.discount && (
              <p className="flex justify-between gap-2">
                <span>Descuento {formatDiscount(line.discount)}</span>
                <span>−{formatMoney(line.discountAmount)}</span>
              </p>
            )}
          </li>
        ))}
      </ul>

      <dl className="flex flex-col gap-0.5 border-b border-dashed py-2">
        {sale.cartDiscount && (
          <>
            <TicketRow label="Subtotal" value={formatMoney(sale.total + sale.cartDiscountAmount)} />
            <TicketRow label={`Descuento general ${formatDiscount(sale.cartDiscount)}`} value={`−${formatMoney(sale.cartDiscountAmount)}`} />
          </>
        )}
        <div className="flex justify-between gap-2 text-sm font-semibold">
          <dt>Total</dt>
          <dd>{formatMoney(sale.total)}</dd>
        </div>
      </dl>

      <ul aria-label="Pagos" className="flex flex-col gap-0.5 py-2">
        {sale.payments.map((payment, index) => (
          <li key={index} className="flex justify-between gap-2">
            <span>{PAYMENT_METHOD_LABELS[payment.method]}</span>
            <span>{formatMoney(payment.amount)}</span>
          </li>
        ))}
        {sale.change > 0 && (
          <li className="flex justify-between gap-2 font-semibold">
            <span>Vuelto</span>
            <span>{formatMoney(sale.change)}</span>
          </li>
        )}
      </ul>

      <p className="border-t border-dashed pt-2 text-center">¡Gracias por su compra!</p>
    </article>,
    document.body,
  )
}

function TicketRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-2">
      <dt>{label}</dt>
      <dd className="text-right">{value}</dd>
    </div>
  )
}
