import type { ComponentProps } from 'react'

import { cn } from '@/shared/lib/utils'
import { Label } from '@/shared/ui/label'

function Field({ className, ...props }: ComponentProps<'div'>) {
  return (
    <div
      data-slot="field"
      className={cn('group/field flex w-full flex-col gap-1.5', className)}
      {...props}
    />
  )
}

function FieldLabel({ className, ...props }: ComponentProps<typeof Label>) {
  return <Label data-slot="field-label" className={cn('w-fit leading-snug', className)} {...props} />
}

function FieldDescription({ className, ...props }: ComponentProps<'p'>) {
  return (
    <p
      data-slot="field-description"
      className={cn('text-sm text-muted-foreground', className)}
      {...props}
    />
  )
}

function FieldError({ className, children, ...props }: ComponentProps<'p'>) {
  if (!children) {
    return null
  }

  return (
    <p
      data-slot="field-error"
      className={cn('text-sm text-destructive', className)}
      {...props}
    >
      {children}
    </p>
  )
}

export { Field, FieldLabel, FieldDescription, FieldError }
