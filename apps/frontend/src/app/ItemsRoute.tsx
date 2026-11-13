import { useMemo } from 'react'
import { useCategoryOptions } from '../features/categories/hooks/use-categories'
import ItemsPage from '../features/items/pages/ItemsPage'
import type { StockQuantities } from '../features/items/types'
import { useStockLevelsOf } from '../features/stock/hooks/use-stock'
import { useSupplierOptions } from '../features/suppliers/hooks/use-suppliers'

function useStockQuantities(itemIds: readonly string[]): StockQuantities {
  const { data, isLoading } = useStockLevelsOf(itemIds)

  return useMemo(
    () => ({ quantities: new Map((data?.items ?? []).map((level) => [level.itemId, level.quantity])), isLoading }),
    [data, isLoading],
  )
}

export default function ItemsRoute() {
  const { categories, isLoading } = useCategoryOptions()
  const { suppliers } = useSupplierOptions()

  return (
    <ItemsPage
      categories={categories}
      categoriesLoading={isLoading}
      suppliers={suppliers}
      useStockQuantities={useStockQuantities}
    />
  )
}
