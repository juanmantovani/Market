import { Pencil, Trash2 } from 'lucide-react'
import { Button } from '../../../shared/ui/button'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import type { Category } from '../types'

type CategoryActionsProps = {
  category: Category
  onEdit: (category: Category) => void
  onDelete: (category: Category) => void
}

type CategoryListProps = Omit<CategoryActionsProps, 'category'> & {
  categories: readonly Category[]
  isLoading: boolean
}

function IconActions({ category, onEdit, onDelete }: CategoryActionsProps) {
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
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        aria-label={`Eliminar ${category.name}`}
        title="Eliminar"
        onClick={() => onDelete(category)}
      >
        <Trash2 aria-hidden />
      </Button>
    </div>
  )
}

function LabeledActions({ category, onEdit, onDelete }: CategoryActionsProps) {
  return (
    <div className="flex gap-2">
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
      <Button
        type="button"
        variant="outline"
        className="h-11 flex-1 text-destructive hover:bg-destructive/10 hover:text-destructive"
        aria-label={`Eliminar ${category.name}`}
        onClick={() => onDelete(category)}
      >
        <Trash2 aria-hidden />
        Eliminar
      </Button>
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
          <div className="flex min-w-0 flex-col gap-1">
            <span className="font-medium">{category.name}</span>
            {category.description && (
              <span className="text-sm text-muted-foreground">{category.description}</span>
            )}
          </div>
          <LabeledActions category={category} {...actions} />
        </div>
      )}
    />
  )
}
