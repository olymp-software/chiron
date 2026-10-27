import { createI18n } from 'vue-i18n'
import en from './en.json'
import { pickLocale } from './locale'
import ru from './ru.json'

export const i18n = createI18n({
  legacy: false,
  locale: pickLocale(navigator.language),
  fallbackLocale: 'en',
  messages: { en, ru },
})
