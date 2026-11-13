import { describe, expect, it } from 'vitest'
import type { CartLine, CatalogItem } from '../types'
import {
  addItem,
  isValidQuantity,
  lineAmount,
  MAX_LINE_QUANTITY,
  pickByCode,
  quantityStep,
  removeLine,
  restoreLine,
  setQuantity,
} from './cart'

const coca: CatalogItem = { id: 'coca', code: 'BEB-1', barcode: '7790000000059', name: 'Gaseosa cola', unit: 'unit', salePrice: 320_000, categoryId: 'c1' }
const pan: CatalogItem = { id: 'pan', code: 'PAN-1', barcode: null, name: 'Pan francés', unit: 'kg', salePrice: 280_000, categoryId: 'c2' }

function line(item: CatalogItem, quantity: number): CartLine {
  return { itemId: item.id, code: item.code, name: item.name, unit: item.unit, unitPrice: item.salePrice, quantity }
}

describe('lineAmount', () => {
  it.each([
    { name: 'whole units', unitPrice: 120_000, quantity: 6_000, want: 720_000 },
    { name: 'a fraction rounds half up', unitPrice: 99_999, quantity: 333, want: 33_300 },
    { name: 'a fraction rounds down', unitPrice: 1_001, quantity: 1, want: 1 },
    { name: 'half a cent rounds up', unitPrice: 1, quantity: 500, want: 1 },
    { name: 'beyond 2^53 before dividing', unitPrice: 100_000_000_000, quantity: 1_000_000_000, want: 100_000_000_000_000_000 },
  ])('$name', ({ unitPrice, quantity, want }) => {
    expect(lineAmount({ unitPrice, quantity })).toBe(want)
  })
})

describe('cart lines', () => {
  it('adds an item once and then adds up its quantity', () => {
    const once = addItem([], coca, 1_000)
    const twice = addItem(once, coca, 2_000)
    const withBread = addItem(twice, pan, 750)

    expect(once).toEqual([line(coca, 1_000)])
    expect(twice).toEqual([line(coca, 3_000)])
    expect(withBread.map((current) => current.itemId)).toEqual(['coca', 'pan'])
    expect(addItem(withBread, coca, MAX_LINE_QUANTITY)[0].quantity).toBe(MAX_LINE_QUANTITY)
    expect(addItem([], coca, MAX_LINE_QUANTITY + 1)[0].quantity).toBe(MAX_LINE_QUANTITY)
  })

  it('changes a quantity only to a valid one', () => {
    const lines = [line(coca, 1_000), line(pan, 750)]

    expect(setQuantity(lines, 'pan', 1_250)[1].quantity).toBe(1_250)
    expect(setQuantity(lines, 'pan', 0)).toEqual(lines)
    expect(setQuantity(lines, 'pan', MAX_LINE_QUANTITY + 1)).toEqual(lines)
    expect(isValidQuantity(1.5)).toBe(false)
    expect(isValidQuantity(-1)).toBe(false)
    expect(isValidQuantity(MAX_LINE_QUANTITY)).toBe(true)
  })

  it('steps by a unit, or by 100 g for items sold by weight', () => {
    expect(quantityStep('unit')).toBe(1_000)
    expect(quantityStep('kg')).toBe(100)
  })

  it('removes a line and puts it back where it was', () => {
    const agua: CatalogItem = { ...coca, id: 'agua', code: 'BEB-2', name: 'Agua' }
    const lines = [line(coca, 1_000), line(pan, 750), line(agua, 2_000)]

    const { lines: without, removed } = removeLine(lines, 'pan')

    expect(without.map((current) => current.itemId)).toEqual(['coca', 'agua'])
    expect(removed).toEqual({ line: line(pan, 750), index: 1 })
    expect(restoreLine(without, removed!)).toEqual(lines)
    expect(restoreLine([], removed!)).toEqual([line(pan, 750)])
    expect(restoreLine(addItem(without, pan, 250), removed!).find((current) => current.itemId === 'pan')?.quantity).toBe(1_000)
    expect(removeLine(lines, 'nada')).toEqual({ lines, removed: null })
  })
})

describe('pickByCode', () => {
  const results = [coca, { ...coca, id: 'coca-light', code: 'BEB-10', barcode: '7790000000066' }]

  it('picks the item with that exact code or barcode', () => {
    expect(pickByCode(' beb-1 ', results)).toBe(coca)
    expect(pickByCode('7790000000066', results)?.id).toBe('coca-light')
  })

  it('picks the only result, and nothing when the code is ambiguous or blank', () => {
    expect(pickByCode('pan', [pan])).toBe(pan)
    expect(pickByCode('BEB', results)).toBeNull()
    expect(pickByCode('  ', [pan])).toBeNull()
    expect(pickByCode('x', [])).toBeNull()
  })
})
