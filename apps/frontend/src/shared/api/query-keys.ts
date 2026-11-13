import type { ListParams } from './pagination'

export const queryKeys = {
  categories: {
    all: ['categories'] as const,
    list: (params: ListParams) => ['categories', 'list', params] as const,
    options: ['categories', 'options'] as const,
  },
  items: {
    all: ['items'] as const,
    list: (params: ListParams & { categoryId: string; lowStock?: boolean; includeDiscontinued?: boolean }) => ['items', 'list', params] as const,
  },
  customers: {
    all: ['customers'] as const,
    list: (params: ListParams) => ['customers', 'list', params] as const,
  },
  suppliers: {
    all: ['suppliers'] as const,
    list: (params: ListParams) => ['suppliers', 'list', params] as const,
    options: ['suppliers', 'options'] as const,
  },
  sellers: {
    all: ['sellers'] as const,
    list: (params: ListParams) => ['sellers', 'list', params] as const,
  },
  sales: {
    all: ['sales'] as const,
    list: (params: { status: string; from: string; to: string; page: number; limit: number }) => ['sales', 'list', params] as const,
  },
  dashboard: {
    all: ['dashboard'] as const,
    summary: (from: string, to: string) => ['dashboard', 'summary', from, to] as const,
  },
  stock: {
    all: ['stock'] as const,
    levels: (params: ListParams) => ['stock', 'levels', params] as const,
    levelsOf: (itemIds: readonly string[]) => ['stock', 'levels-of', itemIds] as const,
    movements: (params: { itemId: string; page: number; limit: number }) => ['stock', 'movements', params] as const,
  },
}
