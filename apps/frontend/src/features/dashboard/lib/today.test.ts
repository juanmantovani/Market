import { describe, expect, it } from 'vitest'
import { todayRange } from './today'

describe('todayRange', () => {
  it('goes from local midnight to the next local midnight', () => {
    const { from, to } = todayRange(new Date(2026, 10, 12, 18, 30))

    expect(new Date(from).getTime()).toBe(new Date(2026, 10, 12).getTime())
    expect(new Date(to).getTime()).toBe(new Date(2026, 10, 13).getTime())
  })

  it('rolls over the end of the month', () => {
    const { to } = todayRange(new Date(2026, 10, 30, 23, 59))

    expect(new Date(to).getTime()).toBe(new Date(2026, 11, 1).getTime())
  })
})
