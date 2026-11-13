import { useCategoryOptions } from '../features/categories/hooks/use-categories'
import { listItems } from '../features/items/api/items'
import PosPage from '../features/sales/pages/PosPage'

export default function PosRoute() {
  const { categories } = useCategoryOptions()

  return <PosPage categories={categories} searchCatalog={listItems} />
}
