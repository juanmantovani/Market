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
}
