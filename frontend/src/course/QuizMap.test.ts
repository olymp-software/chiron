// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import en from '../i18n/en.json'
import type { Result } from '../quiz/quiz'
import QuizMap from './QuizMap.vue'

function mountMap(results: Result[], open = 0) {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  return mount(QuizMap, { props: { results, open }, global: { plugins: [i18n] } })
}

describe('QuizMap', () => {
  it('shows one icon per question with its result', () => {
    const w = mountMap(['correct', 'incorrect', 'unchecked', 'unchecked'])

    expect(w.findAll('nav button').map((b) => b.attributes('aria-label'))).toEqual([
      'Question 1: answered correctly',
      'Question 2: answered incorrectly',
      'Question 3: not answered',
      'Question 4: not answered',
    ])
    expect(w.findAll('nav button').map((b) => b.classes())).toEqual([
      ['item', 'correct'],
      ['item', 'incorrect'],
      ['item', 'unchecked'],
      ['item', 'unchecked'],
    ])
  })

  it('marks the open question', () => {
    const w = mountMap(['unchecked', 'unchecked', 'unchecked'], 1)

    expect(w.findAll('nav button').map((b) => b.attributes('aria-current'))).toEqual([
      undefined,
      'step',
      undefined,
    ])
  })

  it('opens a question in any order', async () => {
    const w = mountMap(['unchecked', 'unchecked', 'unchecked', 'unchecked'])

    await w.findAll('nav button')[3].trigger('click')

    expect(w.emitted('open')).toEqual([[3]])
  })

  it('counts correct answers only', () => {
    expect(mountMap(['incorrect', 'unchecked']).get('.counter').text()).toBe('Correct: 0 of 2')
    expect(mountMap(['correct', 'incorrect', 'correct']).get('.counter').text()).toBe(
      'Correct: 2 of 3',
    )
  })
})
