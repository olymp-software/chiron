export type Locale = 'en' | 'ru'

// pickLocale maps a BCP 47 language tag to a UI locale: ru* → ru, anything else → en.
export function pickLocale(language: string | undefined): Locale {
  return language?.toLowerCase().startsWith('ru') ? 'ru' : 'en'
}
