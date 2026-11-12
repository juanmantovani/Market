import { useCategoryOptions } from '../features/categories/hooks/use-categories'
import ItemsPage from '../features/items/pages/ItemsPage'
import { useSupplierOptions } from '../features/suppliers/hooks/use-suppliers'

export default function ItemsRoute() {
  const { categories, isLoading } = useCategoryOptions()
  const { suppliers } = useSupplierOptions()

  return <ItemsPage categories={categories} categoriesLoading={isLoading} suppliers={suppliers} />
}
