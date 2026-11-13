import { describe, expect, it, vi } from 'vitest'
import { listAllPages, type Page } from './pagination'

function pageOf(items: number[], page: number, total: number): Page<number> {
  return { items, total, page, limit: 100 }
}

describe('listAllPages', () => {
  it('asks for the next page until every item arrived', async () => {
    const fetchPage = vi
      .fn<(page: number, limit: number) => Promise<Page<number>>>()
      .mockResolvedValueOnce(pageOf(Array.from({ length: 100 }, (_, i) => i), 1, 230))
      .mockResolvedValueOnce(pageOf(Array.from({ length: 100 }, (_, i) => 100 + i), 2, 230))
      .mockResolvedValueOnce(pageOf(Array.from({ length: 30 }, (_, i) => 200 + i), 3, 230))

    const items = await listAllPages(fetchPage)

    expect(items).toHaveLength(230)
    expect(items[229]).toBe(229)
    expect(fetchPage.mock.calls).toEqual([[1, 100], [2, 100], [3, 100]])
  })

  it('stops after one page when everything fits, and when a page comes back empty', async () => {
    const single = vi.fn(async () => pageOf([1, 2], 1, 2))
    expect(await listAllPages(single)).toEqual([1, 2])
    expect(single).toHaveBeenCalledTimes(1)

    const shrunk = vi.fn(async (page: number) => (page === 1 ? pageOf([1], 1, 5) : pageOf([], page, 5)))
    expect(await listAllPages(shrunk)).toEqual([1])
    expect(shrunk).toHaveBeenCalledTimes(2)
  })
})
