import { useId, useState, type ReactNode } from 'react'
import { formatDateTime } from '../../../shared/lib/date'
import { formatMoney } from '../../../shared/lib/money'
import { formatPercent } from '../../../shared/lib/percent'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { ConfirmDialog } from '../../../shared/ui/confirm-dialog'
import { Field, FieldDescription, FieldLabel } from '../../../shared/ui/field'
import { Textarea } from '../../../shared/ui/textarea'
import { PAYMENT_METHOD_LABELS } from '../lib/payments'
import { MAX_VOID_REASON_LENGTH, type SaleDiscount, type SaleRecord } from '../types'
import { SaleStatusBadge } from './SaleStatusBadge'

type SaleDetailProps = {
  sale: SaleRecord
  isVoiding: boolean
  onVoid: (reason: string) => Promise<void>
}

function discountLabel(discount: SaleDiscount) {
  return discount.type === 'percent' ? `${formatPercent(discount.value)} %` : formatMoney(discount.value)
}

export function SaleDetail({ sale, isVoiding, onVoid }: SaleDetailProps) {
  return (
    <div className="flex flex-col gap-5">
      <dl className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
        <Info label="Fecha" value={formatDateTime(sale.createdAt)} />
        <Info label="Estado" value={<SaleStatusBadge status={sale.status} />} />
        <Info label="Cliente" value={sale.customer.name} />
        <Info label="Vendedor" value={sale.seller?.name ?? 'Sin vendedor'} />
      </dl>

      {sale.status === 'voided' && (
        <section className="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm">
          <h3 className="font-medium">Venta anulada{sale.voidedAt ? ` el ${formatDateTime(sale.voidedAt)}` : ''}</h3>
          {sale.voidReason && <p className="mt-1 text-muted-foreground">Motivo: {sale.voidReason}</p>}
        </section>
      )}

      <section aria-labelledby="sale-lines" className="flex flex-col gap-2">
        <h3 id="sale-lines" className="text-sm font-semibold">
          Artículos
        </h3>
        <ul className="flex flex-col divide-y rounded-lg border">
          {sale.lines.map((line) => (
            <li key={line.itemId} className="flex items-start justify-between gap-3 px-3 py-2 text-sm">
              <div className="flex min-w-0 flex-col">
                <span className="font-medium">{line.name}</span>
                <span className="text-muted-foreground">
                  {formatUnitQuantity(line.quantity, line.unit)} × {formatMoney(line.unitPrice)}
                </span>
                {line.discount && (
                  <span className="text-muted-foreground">
                    Descuento {discountLabel(line.discount)}: −{formatMoney(line.discountAmount)}
                  </span>
                )}
              </div>
              <span className="shrink-0 font-medium tabular-nums">{formatMoney(line.lineTotal)}</span>
            </li>
          ))}
        </ul>
      </section>

      <section aria-label="Importes" className="flex flex-col gap-1 text-sm">
        <Amount label="Subtotal" cents={sale.subtotal} />
        {sale.cartDiscount && (
          <Amount label={`Descuento general (${discountLabel(sale.cartDiscount)})`} cents={-sale.cartDiscountAmount} />
        )}
        {sale.discountTotal > 0 && <Amount label="Descuentos en total" cents={-sale.discountTotal} muted />}
        <div className="flex justify-between border-t pt-2 text-base font-semibold">
          <span>Total</span>
          <span className="tabular-nums">{formatMoney(sale.total)}</span>
        </div>
      </section>

      <section aria-labelledby="sale-payments" className="flex flex-col gap-2">
        <h3 id="sale-payments" className="text-sm font-semibold">
          Pagos
        </h3>
        <ul className="flex flex-col gap-1 text-sm">
          {sale.payments.map((payment, index) => (
            <li key={index} className="flex justify-between gap-3">
              <span>
                {PAYMENT_METHOD_LABELS[payment.method]}
                {payment.reference && <span className="text-muted-foreground"> · {payment.reference}</span>}
              </span>
              <span className="tabular-nums">{formatMoney(payment.amount)}</span>
            </li>
          ))}
        </ul>
        {sale.change > 0 && <Amount label="Vuelto" cents={sale.change} />}
      </section>

      {sale.status === 'completed' && <VoidSection saleNumber={sale.number} isVoiding={isVoiding} onVoid={onVoid} />}
    </div>
  )
}

function Info({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  )
}

function Amount({ label, cents, muted = false }: { label: string; cents: number; muted?: boolean }) {
  return (
    <div className={muted ? 'flex justify-between text-muted-foreground' : 'flex justify-between'}>
      <span>{label}</span>
      <span className="tabular-nums">{formatMoney(cents)}</span>
    </div>
  )
}

type VoidSectionProps = { saleNumber: number; isVoiding: boolean; onVoid: (reason: string) => Promise<void> }

// The confirmation lives inside the detail so it opens on top of it.
function VoidSection({ saleNumber, isVoiding, onVoid }: VoidSectionProps) {
  const [reason, setReason] = useState('')
  const [asking, setAsking] = useState(false)
  const fieldId = useId()
  const trimmed = reason.trim()

  return (
    <section aria-labelledby={`${fieldId}-title`} className="flex flex-col gap-3 border-t pt-5">
      <h3 id={`${fieldId}-title`} className="text-sm font-semibold">
        Anular la venta
      </h3>
      <Field>
        <FieldLabel htmlFor={fieldId}>Motivo de la anulación</FieldLabel>
        <Textarea
          id={fieldId}
          value={reason}
          maxLength={MAX_VOID_REASON_LENGTH}
          onChange={(event) => setReason(event.target.value)}
          aria-describedby={`${fieldId}-hint`}
        />
        <FieldDescription id={`${fieldId}-hint`}>
          Es obligatorio. Al anular, el stock de los artículos vuelve a sumarse.
        </FieldDescription>
      </Field>
      <Button
        type="button"
        variant="destructive"
        size="lg"
        className="h-11 sm:self-end md:h-9"
        disabled={trimmed === ''}
        onClick={() => setAsking(true)}
      >
        Anular venta
      </Button>

      <ConfirmDialog
        open={asking}
        onOpenChange={setAsking}
        title={`¿Anular la venta N.º ${saleNumber}?`}
        description="La venta queda anulada, se repone el stock y no se puede deshacer."
        confirmLabel="Confirmar anulación"
        isPending={isVoiding}
        onConfirm={() => {
          void onVoid(trimmed).finally(() => setAsking(false))
        }}
      />
    </section>
  )
}
