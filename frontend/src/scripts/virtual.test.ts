import { describe, expect, it } from 'vitest'
import { visibleRange } from './virtual'

describe('visibleRange', () => {
  it('renders the visible lines plus overscan', () => {
    expect(visibleRange(200, 100, 10, 1000, 5)).toEqual({ start: 15, end: 35 })
  })

  it('includes a partly visible line', () => {
    expect(visibleRange(205, 100, 10, 1000, 0)).toEqual({ start: 20, end: 31 })
  })

  it('clamps to the list', () => {
    expect(visibleRange(0, 100, 10, 1000, 5)).toEqual({ start: 0, end: 15 })
    expect(visibleRange(9900, 100, 10, 1000, 5)).toEqual({ start: 985, end: 1000 })
    expect(visibleRange(0, 100, 10, 3, 5)).toEqual({ start: 0, end: 3 })
    expect(visibleRange(0, 100, 10, 0, 5)).toEqual({ start: 0, end: 0 })
  })

  it('renders only overscan before the viewport is measured', () => {
    expect(visibleRange(0, 0, 10, 1000, 5)).toEqual({ start: 0, end: 5 })
  })

  it('stays empty when the list shrank below the scroll position', () => {
    expect(visibleRange(5000, 100, 10, 100, 5)).toEqual({ start: 100, end: 100 })
  })
})
