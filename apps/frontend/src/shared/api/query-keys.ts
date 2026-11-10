import type { ListParams } from './pagination'

export const queryKeys = {
  health: ['health'] as const,
  categories: {
    all: ['categories'] as const,
    list: (params: ListParams) => ['categories', 'list', params] as const,
  },
}
