// Range is the half-open range [start, end) of line indexes to render.
export interface Range {
  start: number
  end: number
}

// visibleRange returns the lines of a list of count lines, each lineHeight
// pixels high, that a viewport scrolled to scrollTop shows, plus overscan
// lines on each side so fast scrolling does not show blank space.
export function visibleRange(
  scrollTop: number,
  viewportHeight: number,
  lineHeight: number,
  count: number,
  overscan: number,
): Range {
  const first = Math.floor(scrollTop / lineHeight)
  const last = Math.ceil((scrollTop + viewportHeight) / lineHeight)
  const start = Math.min(Math.max(0, first - overscan), count)
  const end = Math.min(count, Math.max(start, last + overscan))
  return { start, end }
}
