import { useMemo } from 'react'
import { useCategoryOptions } from '../features/categories/hooks/use-categories'
import ItemsPage from '../features/items/pages/ItemsPage'
import type { RenderStockActions, StockQuantities } from '../features/items/types'
import { ReceiveStock } from '../features/stock/components/ReceiveStock'
import { StockItemActions } from '../features/stock/components/StockItemActions'
import { useStockLevelsOf } from '../features/stock/hooks/use-stock'
import { useSupplierOptions } from '../features/suppliers/hooks/use-suppliers'

function useStockQuantities(itemIds: readonly string[]): StockQuantities {
  const { data, isLoading } = useStockLevelsOf(itemIds)

  return useMemo(() => {
    const levels = data?.items ?? []
    return {
      quantities: new Map(levels.map((level) => [level.itemId, level.quantity])),
      withoutMovements: new Set(levels.filter((level) => level.updatedAt === null).map((level) => level.itemId)),
      isLoading,
    }
  }, [data, isLoading])
}

// Stock is managed from Artículos: this route puts the stock feature's
// actions (adjust, first count, movements, receive goods) next to the items.
export default function ItemsRoute() {
  const { categories, isLoading } = useCategoryOptions()
  const { suppliers } = useSupplierOptions()
  const supplierNames = useMemo(() => new Map(suppliers.map((supplier) => [supplier.id, supplier.name])), [suppliers])

  const renderStockActions: RenderStockActions = (item, stock, layout) => (
    <StockItemActions
      item={item}
      quantity={stock.quantity}
      hasMovements={stock.hasMovements}
      supplierName={(id) => supplierNames.get(id) ?? 'Proveedor eliminado'}
      layout={layout}
    />
  )

  return (
    <ItemsPage
      categories={categories}
      categoriesLoading={isLoading}
      suppliers={suppliers}
      useStockQuantities={useStockQuantities}
      renderStockActions={renderStockActions}
      headerActions={<ReceiveStock suppliers={suppliers} />}
    />
  )
}
