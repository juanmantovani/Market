import { cn } from '@/shared/lib/utils'

type ContactLinesProps = {
  phone: string | null
  email: string | null
  className?: string
}

// Phone and email on their own lines, never cut: a long email wraps
// anywhere instead of hiding behind an ellipsis.
export function ContactLines({ phone, email, className }: ContactLinesProps) {
  return (
    <span className={cn('flex flex-col', className)}>
      {phone && <span>{phone}</span>}
      {phone && email && <span className="sr-only">, </span>}
      {email && <span className="[overflow-wrap:anywhere]">{email}</span>}
    </span>
  )
}
