import type { ListParams } from './pagination'

export const queryKeys = {
  categories: {
    all: ['categories'] as const,
    list: (params: ListParams) => ['categories', 'list', params] as const,
  },
  items: {
    all: ['items'] as const,
    list: (params: ListParams & { categoryId: string }) => ['items', 'list', params] as const,
  },
  customers: {
    all: ['customers'] as const,
    list: (params: ListParams) => ['customers', 'list', params] as const,
  },
  suppliers: {
    all: ['suppliers'] as const,
    list: (params: ListParams) => ['suppliers', 'list', params] as const,
  },
  sellers: {
    all: ['sellers'] as const,
    list: (params: ListParams) => ['sellers', 'list', params] as const,
  },
  stock: {
    all: ['stock'] as const,
    levels: (params: ListParams & { categoryId: string; lowStock: boolean }) => ['stock', 'levels', params] as const,
    levelsOf: (itemIds: readonly string[]) => ['stock', 'levels-of', itemIds] as const,
    movements: (params: { itemId: string; page: number; limit: number }) => ['stock', 'movements', params] as const,
  },
}
