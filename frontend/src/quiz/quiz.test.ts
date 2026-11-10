import { describe, expect, it } from 'vitest'
import { library } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import ru from '../i18n/ru.json'
import { freshAnswers, isCorrect, result } from './quiz'

function question(correct: boolean[], multiple = false): library.Question {
  return library.Question.createFrom({
    kind: 'choice',
    prompt: '',
    explanation: '',
    multiple,
    options: correct.map((c, i) => ({ text: `o${i}`, correct: c, feedback: '' })),
  })
}

describe('isCorrect', () => {
  it('accepts the single correct option', () => {
    expect(isCorrect(question([false, true]), [1])).toBe(true)
  })

  it('rejects a wrong single option', () => {
    expect(isCorrect(question([false, true]), [0])).toBe(false)
  })

  it('rejects a partial selection', () => {
    expect(isCorrect(question([true, true, false], true), [0])).toBe(false)
  })

  it('rejects an extra selection', () => {
    expect(isCorrect(question([true, false, false], true), [0, 2])).toBe(false)
  })

  it('accepts the exact set of correct options', () => {
    expect(isCorrect(question([true, false, true], true), [2, 0])).toBe(true)
  })

  it('rejects an empty selection', () => {
    expect(isCorrect(question([true, false]), [])).toBe(false)
  })
})

describe('result', () => {
  it('maps answers to map states', () => {
    expect(result({ selected: [0], checked: false, correct: false })).toBe('unchecked')
    expect(result({ selected: [0], checked: true, correct: true })).toBe('correct')
    expect(result({ selected: [0], checked: true, correct: false })).toBe('incorrect')
  })
})

it('has the same quiz strings in every locale', () => {
  expect(Object.keys(ru.quiz).sort()).toEqual(Object.keys(en.quiz).sort())
})

it('freshAnswers has an empty answer per question', () => {
  const quiz = library.Quiz.createFrom({
    questions: [question([true, false]), question([true, false])],
  })

  expect(freshAnswers(quiz)).toEqual([
    { selected: [], checked: false, correct: false },
    { selected: [], checked: false, correct: false },
  ])
})
