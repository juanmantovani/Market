import { useCategoryOptions } from '../features/categories/hooks/use-categories'
import { listCustomers } from '../features/customers/api/customers'
import { listItems } from '../features/items/api/items'
import PosPage from '../features/sales/pages/PosPage'
import { listSellers } from '../features/sellers/api/sellers'

// The counter reads the catalog, the customers and the sellers through
// their own features' clients, so it shares their cache.
export default function PosRoute() {
  const { categories } = useCategoryOptions()

  return (
    <PosPage
      categories={categories}
      searchCatalog={listItems}
      searchCustomers={listCustomers}
      searchSellers={listSellers}
    />
  )
}
