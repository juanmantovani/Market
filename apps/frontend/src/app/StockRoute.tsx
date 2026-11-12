import { useCategoryOptions } from '../features/categories/hooks/use-categories'
import StockPage from '../features/stock/pages/StockPage'
import { useSupplierOptions } from '../features/suppliers/hooks/use-suppliers'

export default function StockRoute() {
  const { categories } = useCategoryOptions()
  const { suppliers } = useSupplierOptions()

  return <StockPage categories={categories} suppliers={suppliers} />
}
