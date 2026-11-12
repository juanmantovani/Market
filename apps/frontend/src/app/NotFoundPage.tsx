import { SearchX } from 'lucide-react'
import { Link } from 'react-router'
import { buttonVariants } from '../shared/ui/button-variants'
import { EmptyState } from '../shared/ui/empty-state'

export default function NotFoundPage() {
  return (
    <EmptyState
      icon={SearchX}
      title="No se encontró esta página"
      description="Puede que la dirección esté mal escrita o que la sección todavía no exista."
      action={
        <Link to="/" className={buttonVariants({ size: 'lg' })}>
          Volver al inicio
        </Link>
      }
    />
  )
}
