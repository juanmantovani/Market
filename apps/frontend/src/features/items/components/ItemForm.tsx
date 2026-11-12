import { useState, type FormEvent, type ReactNode } from 'react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { Button } from '../../../shared/ui/button'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import { NativeSelect } from '../../../shared/ui/native-select'
import { Textarea } from '../../../shared/ui/textarea'
import {
  ITEM_FIELD_ERRORS,
  itemFormSchema,
  TAX_RATES,
  UNITS,
  unitShort,
  type CategoryOption,
  type SupplierOption,
  type ItemField,
  type ItemFormValues,
  type ItemRequest,
} from '../types'

type ItemFormProps = {
  initialValues: ItemFormValues
  categories: readonly CategoryOption[]
  suppliers: readonly SupplierOption[]
  submitLabel: string
  onSubmit: (request: ItemRequest) => Promise<unknown>
  onCancel: () => void
}

type FieldErrors = Partial<Record<ItemField, string>>

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <fieldset className="flex flex-col gap-4">
      <legend className="mb-1 text-sm font-semibold text-muted-foreground">{title}</legend>
      {children}
    </fieldset>
  )
}

function errorProps(field: ItemField, errors: FieldErrors) {
  return {
    id: `item-${field}`,
    'aria-invalid': Boolean(errors[field]),
    'aria-describedby': errors[field] ? `item-${field}-error` : undefined,
  }
}

export function ItemForm({ initialValues, categories, suppliers, submitLabel, onSubmit, onCancel }: ItemFormProps) {
  const [values, setValues] = useState(initialValues)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function change<F extends ItemField>(field: F, value: ItemFormValues[F]) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
  }

  function toggleSupplier(supplierId: string, selected: boolean) {
    setValues((current) => ({
      ...current,
      supplierIds: selected
        ? [...current.supplierIds, supplierId]
        : current.supplierIds.filter((id) => id !== supplierId),
    }))
    setFieldErrors((current) => ({ ...current, supplierIds: undefined }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = itemFormSchema.safeParse(values)
    if (!parsed.success) {
      const errors: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        const field = issue.path[0] as ItemField
        errors[field] ??= issue.message
      }
      setFieldErrors(errors)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      const field = error instanceof ApiError ? ITEM_FIELD_ERRORS[error.code] : undefined
      if (field) {
        setFieldErrors({ [field]: messageFromError(error) })
      } else {
        setFormError(messageFromError(error))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const error = (field: ItemField) => <FieldError id={`item-${field}-error`}>{fieldErrors[field]}</FieldError>

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-6">
      <Section title="Datos del artículo">
        <Field>
          <FieldLabel htmlFor="item-name">Nombre</FieldLabel>
          <Input
            {...errorProps('name', fieldErrors)}
            autoFocus
            autoComplete="off"
            value={values.name}
            onChange={(event) => change('name', event.target.value)}
          />
          {error('name')}
        </Field>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field>
            <FieldLabel htmlFor="item-code">Código</FieldLabel>
            <Input
              {...errorProps('code', fieldErrors)}
              autoComplete="off"
              value={values.code}
              onChange={(event) => change('code', event.target.value)}
            />
            {error('code')}
          </Field>
          <Field>
            <FieldLabel htmlFor="item-barcode">
              Código de barras <span className="font-normal text-muted-foreground">(opcional)</span>
            </FieldLabel>
            <Input
              {...errorProps('barcode', fieldErrors)}
              autoComplete="off"
              inputMode="numeric"
              value={values.barcode}
              onChange={(event) => change('barcode', event.target.value)}
            />
            {error('barcode')}
          </Field>
          <Field>
            <FieldLabel htmlFor="item-categoryId">Categoría</FieldLabel>
            <NativeSelect
              {...errorProps('categoryId', fieldErrors)}
              value={values.categoryId}
              onChange={(event) => change('categoryId', event.target.value)}
            >
              <option value="">Elegir una categoría</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </NativeSelect>
            {error('categoryId')}
          </Field>
          <Field>
            <FieldLabel htmlFor="item-unit">Se vende por</FieldLabel>
            <NativeSelect
              {...errorProps('unit', fieldErrors)}
              value={values.unit}
              onChange={(event) => change('unit', event.target.value as ItemFormValues['unit'])}
            >
              {UNITS.map((unit) => (
                <option key={unit.value} value={unit.value}>
                  {unit.label}
                </option>
              ))}
            </NativeSelect>
            {error('unit')}
          </Field>
        </div>
        <Field>
          <FieldLabel htmlFor="item-description">
            Descripción <span className="font-normal text-muted-foreground">(opcional)</span>
          </FieldLabel>
          <Textarea
            {...errorProps('description', fieldErrors)}
            rows={2}
            value={values.description}
            onChange={(event) => change('description', event.target.value)}
          />
          {error('description')}
        </Field>
      </Section>

      <Section title="Precio">
        <div className="grid gap-4 sm:grid-cols-3">
          <Field>
            <FieldLabel htmlFor="item-salePrice">Precio de venta</FieldLabel>
            <div className="relative">
              <span className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted-foreground">$</span>
              <Input
                {...errorProps('salePrice', fieldErrors)}
                inputMode="decimal"
                autoComplete="off"
                placeholder="0,00"
                className="pl-7"
                value={values.salePrice}
                onChange={(event) => change('salePrice', event.target.value)}
              />
            </div>
            {error('salePrice')}
          </Field>
          <Field>
            <FieldLabel htmlFor="item-cost">
              Costo <span className="font-normal text-muted-foreground">(opcional)</span>
            </FieldLabel>
            <div className="relative">
              <span className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted-foreground">$</span>
              <Input
                {...errorProps('cost', fieldErrors)}
                inputMode="decimal"
                autoComplete="off"
                placeholder="0,00"
                className="pl-7"
                value={values.cost}
                onChange={(event) => change('cost', event.target.value)}
              />
            </div>
            {error('cost')}
          </Field>
          <Field>
            <FieldLabel htmlFor="item-taxRate">IVA</FieldLabel>
            <NativeSelect
              {...errorProps('taxRate', fieldErrors)}
              value={values.taxRate}
              onChange={(event) => change('taxRate', event.target.value)}
            >
              {TAX_RATES.map((rate) => (
                <option key={rate.value} value={String(rate.value)}>
                  {rate.label}
                </option>
              ))}
            </NativeSelect>
            {error('taxRate')}
          </Field>
        </div>
        <p className="text-sm text-muted-foreground">El precio de venta incluye IVA.</p>
      </Section>

      <Section title="Stock">
        <label className="flex min-h-11 items-center gap-3 text-sm font-medium">
          <input
            type="checkbox"
            className="size-5 accent-primary"
            checked={values.tracksStock}
            onChange={(event) => change('tracksStock', event.target.checked)}
          />
          Controla stock
          <span className="font-normal text-muted-foreground">(desmarcar para servicios)</span>
        </label>
        <Field className="sm:max-w-xs">
          <FieldLabel htmlFor="item-minStock">
            Stock mínimo <span className="font-normal text-muted-foreground">({unitShort(values.unit)})</span>
          </FieldLabel>
          <Input
            {...errorProps('minStock', fieldErrors)}
            inputMode="decimal"
            autoComplete="off"
            disabled={!values.tracksStock}
            value={values.minStock}
            onChange={(event) => change('minStock', event.target.value)}
          />
          {error('minStock')}
        </Field>
      </Section>

      <fieldset
        className="flex flex-col gap-3"
        aria-describedby={fieldErrors.supplierIds ? 'item-supplierIds-error' : undefined}
      >
        <legend className="mb-1 text-sm font-semibold text-muted-foreground">
          Proveedores <span className="font-normal">(opcional)</span>
        </legend>
        {suppliers.length === 0 ? (
          <p className="text-sm text-muted-foreground">Todavía no hay proveedores cargados.</p>
        ) : (
          <div className="grid max-h-48 gap-1 overflow-y-auto rounded-lg border p-2 sm:grid-cols-2">
            {suppliers.map((supplier) => (
              <label key={supplier.id} className="flex min-h-11 items-center gap-3 rounded-md px-2 text-sm hover:bg-muted">
                <input
                  type="checkbox"
                  className="size-5 shrink-0 accent-primary"
                  checked={values.supplierIds.includes(supplier.id)}
                  onChange={(event) => toggleSupplier(supplier.id, event.target.checked)}
                />
                {supplier.name}
              </label>
            ))}
          </div>
        )}
        {error('supplierIds')}
      </fieldset>

      {formError && (
        <p role="alert" className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {formError}
        </p>
      )}

      <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={onCancel}>
          Cancelar
        </Button>
        <Button type="submit" size="lg" className="h-11 md:h-9" disabled={isSubmitting}>
          {isSubmitting ? 'Guardando…' : submitLabel}
        </Button>
      </div>
    </form>
  )
}
