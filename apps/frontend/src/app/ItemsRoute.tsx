import { useCategoryOptions } from '../features/categories/hooks/use-categories'
import ItemsPage from '../features/items/pages/ItemsPage'

export default function ItemsRoute() {
  const { categories, isLoading } = useCategoryOptions()

  return <ItemsPage categories={categories} categoriesLoading={isLoading} />
}
