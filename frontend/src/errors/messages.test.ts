import { describe, expect, it } from 'vitest'
import en from '../i18n/en.json'
import ru from '../i18n/ru.json'
import { errorCodes } from './codes'

describe.each([
  ['en', en.errors],
  ['ru', ru.errors],
])('%s error messages', (_, messages: Record<string, string>) => {
  it.each(errorCodes)('has a message for %s', (code) => {
    expect(messages[code]).toBeTruthy()
  })

  it('has no message for an unknown code', () => {
    const known: readonly string[] = errorCodes
    expect(Object.keys(messages).filter((k) => !known.includes(k))).toEqual([])
  })
})
