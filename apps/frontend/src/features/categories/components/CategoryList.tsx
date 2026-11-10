import { Archive, ArchiveRestore, Pencil } from 'lucide-react'
import { Button } from '../../../shared/ui/button'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import type { Category } from '../types'
import { CategoryStatusBadge } from './CategoryStatusBadge'

type CategoryActionsProps = {
  category: Category
  onEdit: (category: Category) => void
  onDeactivate: (category: Category) => void
  onReactivate: (category: Category) => void
}

type CategoryListProps = Omit<CategoryActionsProps, 'category'> & {
  categories: readonly Category[]
  isLoading: boolean
}

function IconActions({ category, onEdit, onDeactivate, onReactivate }: CategoryActionsProps) {
  return (
    <div className="flex justify-end gap-1">
      <Button
        type="button"
        variant="ghost"
        size="icon"
        aria-label={`Editar ${category.name}`}
        title="Editar"
        onClick={() => onEdit(category)}
      >
        <Pencil aria-hidden />
      </Button>
      {category.active ? (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`Desactivar ${category.name}`}
          title="Desactivar"
          onClick={() => onDeactivate(category)}
        >
          <Archive aria-hidden />
        </Button>
      ) : (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`Reactivar ${category.name}`}
          title="Reactivar"
          onClick={() => onReactivate(category)}
        >
          <ArchiveRestore aria-hidden />
        </Button>
      )}
    </div>
  )
}

function LabeledActions({ category, onEdit, onDeactivate, onReactivate }: CategoryActionsProps) {
  return (
    <div className="flex flex-wrap gap-2">
      <Button
        type="button"
        variant="outline"
        className="h-11 flex-1"
        aria-label={`Editar ${category.name}`}
        onClick={() => onEdit(category)}
      >
        <Pencil aria-hidden />
        Editar
      </Button>
      {category.active ? (
        <Button
          type="button"
          variant="outline"
          className="h-11 flex-1"
          aria-label={`Desactivar ${category.name}`}
          onClick={() => onDeactivate(category)}
        >
          <Archive aria-hidden />
          Desactivar
        </Button>
      ) : (
        <Button
          type="button"
          variant="outline"
          className="h-11 flex-1"
          aria-label={`Reactivar ${category.name}`}
          onClick={() => onReactivate(category)}
        >
          <ArchiveRestore aria-hidden />
          Reactivar
        </Button>
      )}
    </div>
  )
}

export function CategoryList({ categories, isLoading, ...actions }: CategoryListProps) {
  const columns: Column<Category>[] = [
    { header: 'Nombre', cell: (category) => <span className="font-medium">{category.name}</span> },
    {
      header: 'Descripción',
      className: 'w-full max-w-0',
      cell: (category) => (
        <span className="block truncate text-muted-foreground" title={category.description ?? undefined}>
          {category.description ?? '—'}
        </span>
      ),
    },
    { header: 'Estado', cell: (category) => <CategoryStatusBadge active={category.active} /> },
    {
      header: 'Acciones',
      className: 'text-right',
      cell: (category) => <IconActions category={category} {...actions} />,
    },
  ]

  return (
    <ResponsiveList
      label="Categorías"
      items={categories}
      columns={columns}
      isLoading={isLoading}
      getKey={(category) => category.id}
      renderCard={(category) => (
        <div className="flex flex-col gap-3">
          <div className="flex items-start justify-between gap-3">
            <div className="flex min-w-0 flex-col gap-1">
              <span className="font-medium">{category.name}</span>
              {category.description && (
                <span className="text-sm text-muted-foreground">{category.description}</span>
              )}
            </div>
            <CategoryStatusBadge active={category.active} />
          </div>
          <LabeledActions category={category} {...actions} />
        </div>
      )}
    />
  )
}
