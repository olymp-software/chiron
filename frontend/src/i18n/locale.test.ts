import { describe, expect, it } from 'vitest'
import { pickLocale } from './locale'

describe('pickLocale', () => {
  it.each([
    ['ru', 'ru'],
    ['ru-RU', 'ru'],
    ['RU-ru', 'ru'],
    ['en-US', 'en'],
    ['de-DE', 'en'],
    ['', 'en'],
    [undefined, 'en'],
  ])('%s → %s', (language, expected) => {
    expect(pickLocale(language)).toBe(expected)
  })
})
