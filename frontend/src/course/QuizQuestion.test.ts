// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { library } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import type { Answer } from '../quiz/quiz'
import QuizQuestion from './QuizQuestion.vue'

function question(fields: Partial<library.Question> = {}): library.Question {
  return library.Question.createFrom({
    kind: 'choice',
    prompt: 'Pick one',
    explanation: '',
    multiple: false,
    options: [
      { text: 'A', correct: true, feedback: '' },
      { text: 'B', correct: false, feedback: 'Not B' },
      { text: 'C', correct: false, feedback: 'Not C' },
    ],
    ...fields,
  })
}

function answer(fields: Partial<Answer> = {}): Answer {
  return { selected: [], checked: false, correct: false, ...fields }
}

async function mountQuestion(q = question(), a = answer()) {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(QuizQuestion, {
    props: { question: q, answer: a, n: 2, total: 5, resolveImage: () => null },
    global: { plugins: [i18n] },
  })
  await vi.waitFor(() => expect(w.findAll('.markdown').every((m) => m.text() !== '')).toBe(true))
  return w
}

describe('QuizQuestion', () => {
  it('shows the position, prompt and radio options', async () => {
    const w = await mountQuestion()

    expect(w.get('.position').text()).toBe('Question 2 of 5')
    expect(w.get('.markdown').text()).toBe('Pick one')
    expect(w.findAll('input[type=radio]')).toHaveLength(3)
    expect(w.findAll('.option .text').map((t) => t.text())).toEqual(['A', 'B', 'C'])
  })

  it('shows only the options without a prompt', async () => {
    const w = await mountQuestion(question({ prompt: '' }))

    expect(w.findAll('.markdown').map((m) => m.text())).toEqual(['A', 'B', 'C'])
  })

  it('shows checkboxes for multiple', async () => {
    const w = await mountQuestion(question({ multiple: true }))

    expect(w.findAll('input[type=checkbox]')).toHaveLength(3)
  })

  it('replaces the selection of a single choice', async () => {
    const w = await mountQuestion(question(), answer({ selected: [0] }))

    await w.findAll('input')[1].trigger('change')

    expect(w.emitted('select')).toEqual([[[1]]])
  })

  it('toggles options of a multiple choice', async () => {
    const w = await mountQuestion(question({ multiple: true }), answer({ selected: [2] }))

    await w.findAll('input')[0].trigger('change')
    await w.findAll('input')[2].trigger('change')

    expect(w.emitted('select')).toEqual([[[0, 2]], [[]]])
  })

  it('disables Check without a selection', async () => {
    const w = await mountQuestion()

    expect(w.get('button').text()).toBe('Check')
    expect(w.get('button').attributes('disabled')).toBeDefined()
  })

  it('emits check with a selection', async () => {
    const w = await mountQuestion(question(), answer({ selected: [1] }))

    await w.get('button').trigger('click')

    expect(w.emitted('check')).toHaveLength(1)
  })

  it('locks a checked question', async () => {
    const w = await mountQuestion(question(), answer({ selected: [1], checked: true }))

    expect(w.findAll('input').every((i) => i.attributes('disabled') !== undefined)).toBe(true)
    expect(w.find('button').exists()).toBe(false)
    await w.findAll('input')[0].trigger('change')
    expect(w.emitted('select')).toBeUndefined()
  })

  // marks returns, per option, whether it is highlighted as correct and its
  // verdict icon.
  function marks(w: Awaited<ReturnType<typeof mountQuestion>>) {
    return w.findAll('.option').map((o) => ({
      right: o.classes('right'),
      icon: o.find('.verdict').exists() ? o.get('.verdict').text() : '',
      status: o.find('.hidden').exists() ? o.get('.hidden').text() : '',
    }))
  }

  it('highlights the correct option and puts icons on the selected one', async () => {
    const w = await mountQuestion(question(), answer({ selected: [1], checked: true }))

    expect(marks(w)).toEqual([
      { right: true, icon: '', status: 'Correct option, not selected' },
      { right: false, icon: '✘', status: 'Your choice, incorrect' },
      { right: false, icon: '', status: '' },
    ])
    expect(w.get('.result').text()).toBe('Incorrect')
  })

  it('shows no marks before checking', async () => {
    const w = await mountQuestion(question(), answer({ selected: [1] }))

    expect(marks(w).every((m) => !m.right && !m.icon && !m.status)).toBe(true)
  })

  it('highlights a missed correct option without an icon', async () => {
    const q = question({
      multiple: true,
      options: [
        { text: 'A', correct: true, feedback: '' },
        { text: 'B', correct: true, feedback: '' },
        { text: 'C', correct: false, feedback: '' },
      ],
    })
    const w = await mountQuestion(q, answer({ selected: [0], checked: true }))

    expect(marks(w)).toEqual([
      { right: true, icon: '✔', status: 'Your choice, correct' },
      { right: true, icon: '', status: 'Correct option, not selected' },
      { right: false, icon: '', status: '' },
    ])
    expect(w.get('.result').text()).toBe('Incorrect')
  })

  it('shows the feedback of selected options and the explanation', async () => {
    const w = await mountQuestion(
      question({ explanation: 'Because A' }),
      answer({ selected: [1], checked: true }),
    )

    expect(w.findAll('.feedback').map((f) => f.text())).toEqual(['Not B'])
    expect(w.get('.explanation').text()).toBe('Because A')
  })

  it('shows no feedback or explanation before checking', async () => {
    const w = await mountQuestion(question({ explanation: 'Because A' }), answer({ selected: [1] }))

    expect(w.find('.feedback').exists()).toBe(false)
    expect(w.find('.explanation').exists()).toBe(false)
    expect(w.find('.result').exists()).toBe(false)
  })

  it('shows Correct for a correct answer', async () => {
    const w = await mountQuestion(
      question(),
      answer({ selected: [0], checked: true, correct: true }),
    )

    expect(w.get('.result').text()).toBe('Correct')
  })
})
