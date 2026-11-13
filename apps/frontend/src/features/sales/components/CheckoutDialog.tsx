import { useRef, useState, type FormEvent } from 'react'
import { CircleCheck, Plus, Trash2 } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import { centsToInput, formatMoney } from '../../../shared/lib/money'
import { Button } from '../../../shared/ui/button'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { Input } from '../../../shared/ui/input'
import { NativeSelect } from '../../../shared/ui/native-select'
import { useCreateSale } from '../hooks/use-create-sale'
import { newIdempotencyKey } from '../lib/idempotency-key'
import {
  amountDueExcluding,
  MAX_PAYMENTS,
  MAX_REFERENCE_LENGTH,
  PAYMENT_METHOD_LABELS,
  PAYMENT_METHODS,
  summarizePayments,
  type PaymentDraft,
} from '../lib/payments'
import type { CartLine, PaymentMethod, PaymentRequest, Sale } from '../types'

// Bills the cashier is most likely to be handed, offered as quick amounts.
const QUICK_CASH_AMOUNTS = [1_000_000, 2_000_000]

type CheckoutDialogProps = {
  open: boolean
  lines: CartLine[]
  total: number
  // Called once the sale is recorded, so the cart can start empty.
  onSold: (sale: Sale) => void
  onClose: () => void
}

export function CheckoutDialog({ open, lines, total, onSold, onClose }: CheckoutDialogProps) {
  const create = useCreateSale()
  const [sale, setSale] = useState<Sale | null>(null)
  // The key is reused while the same sale is retried after a failure, and
  // replaced when the sale changes or is recorded.
  const attempt = useRef<{ key: string; signature: string } | null>(null)

  function close() {
    setSale(null)
    create.reset()
    onClose()
  }

  function submit(payments: PaymentRequest[]) {
    const request = {
      customerId: null,
      sellerId: null,
      lines: lines.map((line) => ({ itemId: line.itemId, quantity: line.quantity })),
      payments,
      expectedTotal: total,
    }
    const signature = JSON.stringify(request)
    if (attempt.current?.signature !== signature) attempt.current = { key: newIdempotencyKey(), signature }

    create.mutate(
      { request, idempotencyKey: attempt.current.key },
      {
        onSuccess: (recorded) => {
          attempt.current = null
          setSale(recorded)
          onSold(recorded)
        },
      },
    )
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => {
        // Closing while the request is in flight would hide its outcome.
        if (!next && !create.isPending) close()
      }}
      title={sale ? 'Venta registrada' : 'Cobrar'}
    >
      {open &&
        (sale ? (
          <SaleDone sale={sale} onNewSale={close} />
        ) : (
          <PaymentForm
            total={total}
            pending={create.isPending}
            error={create.isError ? messageFromError(create.error) : null}
            onSubmit={submit}
            onCancel={close}
          />
        ))}
    </FormDialog>
  )
}

type PaymentFormProps = {
  total: number
  pending: boolean
  error: string | null
  onSubmit: (payments: PaymentRequest[]) => void
  onCancel: () => void
}

function PaymentForm({ total, pending, error, onSubmit, onCancel }: PaymentFormProps) {
  const nextId = useRef(1)
  const [drafts, setDrafts] = useState<PaymentDraft[]>(() => [
    { id: 'p0', method: 'cash', amount: centsToInput(total), reference: '' },
  ])
  const summary = summarizePayments(total, drafts)

  function update(id: string, changes: Partial<PaymentDraft>) {
    setDrafts((current) => current.map((draft) => (draft.id === id ? { ...draft, ...changes } : draft)))
  }

  function addPayment() {
    const due = summary.remaining
    const id = `p${nextId.current++}`
    setDrafts((current) => [
      ...current,
      { id, method: 'debit_card', amount: due > 0 ? centsToInput(due) : '', reference: '' },
    ])
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (summary.canConfirm && !pending) onSubmit(summary.payments)
  }

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5">
      <dl className="flex items-baseline justify-between gap-4">
        <dt className="text-base font-medium">Total a cobrar</dt>
        <dd className="text-3xl font-semibold tabular-nums">{formatMoney(total)}</dd>
      </dl>

      <ul aria-label="Medios de pago" className="flex flex-col gap-4">
        {drafts.map((draft, index) => {
          const position = drafts.length > 1 ? ` ${index + 1}` : ''
          const invalid = summary.invalidIds.includes(draft.id)
          return (
            <li key={draft.id} className="flex flex-col gap-2 rounded-xl border p-3">
              <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] gap-3">
                <Field>
                  <FieldLabel htmlFor={`${draft.id}-method`}>Medio de pago{position}</FieldLabel>
                  <NativeSelect
                    id={`${draft.id}-method`}
                    value={draft.method}
                    onChange={(event) => update(draft.id, { method: event.target.value as PaymentMethod })}
                  >
                    {PAYMENT_METHODS.map((method) => (
                      <option key={method} value={method}>
                        {PAYMENT_METHOD_LABELS[method]}
                      </option>
                    ))}
                  </NativeSelect>
                </Field>
                <Field>
                  <FieldLabel htmlFor={`${draft.id}-amount`}>Importe{position}</FieldLabel>
                  <Input
                    id={`${draft.id}-amount`}
                    autoFocus={index === 0}
                    inputMode="decimal"
                    autoComplete="off"
                    value={draft.amount}
                    aria-invalid={invalid}
                    aria-describedby={invalid ? `${draft.id}-amount-error` : undefined}
                    onFocus={(event) => event.currentTarget.select()}
                    onChange={(event) => update(draft.id, { amount: event.target.value })}
                  />
                </Field>
              </div>
              <FieldError id={`${draft.id}-amount-error`}>
                {invalid ? 'Ingresar un importe mayor a cero, por ejemplo 1.500,50' : null}
              </FieldError>

              {draft.method === 'cash' ? (
                <div className="flex flex-wrap gap-2">
                  <QuickAmount label="Exacto" onClick={() => update(draft.id, { amount: centsToInput(amountDueExcluding(total, drafts, draft.id)) })} />
                  {QUICK_CASH_AMOUNTS.filter((amount) => amount >= amountDueExcluding(total, drafts, draft.id)).map((amount) => (
                    <QuickAmount key={amount} label={formatMoney(amount)} onClick={() => update(draft.id, { amount: centsToInput(amount) })} />
                  ))}
                </div>
              ) : (
                <Field>
                  <FieldLabel htmlFor={`${draft.id}-reference`}>Referencia{position} (opcional)</FieldLabel>
                  <Input
                    id={`${draft.id}-reference`}
                    autoComplete="off"
                    maxLength={MAX_REFERENCE_LENGTH}
                    value={draft.reference}
                    onChange={(event) => update(draft.id, { reference: event.target.value })}
                  />
                </Field>
              )}

              {drafts.length > 1 && (
                <Button
                  type="button"
                  variant="ghost"
                  size="lg"
                  className="h-11 self-end text-muted-foreground md:h-9"
                  onClick={() => setDrafts((current) => current.filter((other) => other.id !== draft.id))}
                >
                  <Trash2 aria-hidden />
                  Quitar medio de pago{position}
                </Button>
              )}
            </li>
          )
        })}
      </ul>

      {drafts.length < MAX_PAYMENTS && (
        <Button type="button" variant="outline" size="lg" className="h-11 self-start md:h-9" onClick={addPayment}>
          <Plus aria-hidden />
          Agregar otro medio de pago
        </Button>
      )}

      <div aria-live="polite" className="rounded-xl bg-muted px-4 py-3">
        {summary.nonCashExceedsTotal ? (
          <p className="text-sm font-medium text-destructive">
            Los pagos que no son en efectivo no pueden superar el total.
          </p>
        ) : summary.remaining > 0 ? (
          <p className="flex items-baseline justify-between gap-4">
            <span className="text-base font-medium">Falta cobrar</span>
            <span className="text-2xl font-semibold tabular-nums">{formatMoney(summary.remaining)}</span>
          </p>
        ) : (
          <p className="flex items-baseline justify-between gap-4">
            <span className="text-base font-medium">Vuelto</span>
            <span className="text-2xl font-semibold tabular-nums">{formatMoney(summary.change)}</span>
          </p>
        )}
      </div>

      {error && (
        <p role="alert" className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </p>
      )}

      <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={onCancel} disabled={pending}>
          Cancelar
        </Button>
        <Button type="submit" size="lg" className="h-11 md:h-9" disabled={!summary.canConfirm || pending}>
          {pending ? 'Registrando…' : 'Confirmar venta'}
        </Button>
      </div>
    </form>
  )
}

function QuickAmount({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <Button type="button" variant="secondary" size="lg" className="h-11 md:h-9" onClick={onClick}>
      {label}
    </Button>
  )
}

function SaleDone({ sale, onNewSale }: { sale: Sale; onNewSale: () => void }) {
  return (
    <div className="flex flex-col items-center gap-4 text-center">
      <CircleCheck className="size-10 text-primary" aria-hidden />
      <div role="status" className="flex flex-col gap-1">
        <p className="text-lg font-semibold">Venta N.º {sale.number}</p>
        <p className="text-sm text-muted-foreground">Total: {formatMoney(sale.total)}</p>
      </div>
      <ul aria-label="Pagos" className="w-full divide-y rounded-xl border text-sm">
        {sale.payments.map((payment, index) => (
          <li key={index} className="flex justify-between gap-4 px-3 py-2">
            <span>{PAYMENT_METHOD_LABELS[payment.method]}</span>
            <span className="tabular-nums">{formatMoney(payment.amount)}</span>
          </li>
        ))}
      </ul>
      {sale.change > 0 && (
        <p className="flex w-full items-baseline justify-between gap-4 rounded-xl bg-muted px-4 py-3">
          <span className="text-base font-medium">Vuelto</span>
          <span className="text-3xl font-semibold tabular-nums">{formatMoney(sale.change)}</span>
        </p>
      )}
      <Button type="button" size="lg" autoFocus className="h-12 w-full text-base" onClick={onNewSale}>
        Nueva venta
      </Button>
    </div>
  )
}
