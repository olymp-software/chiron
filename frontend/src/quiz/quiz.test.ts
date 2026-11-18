import { describe, expect, it } from 'vitest'
import { library } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import ru from '../i18n/ru.json'
import { arrange, freshAnswers, isCorrect, result } from './quiz'

function question(correct: boolean[], multiple = false, shuffle = false): library.Question {
  return library.Question.createFrom({
    kind: 'choice',
    prompt: '',
    explanation: '',
    multiple,
    shuffle,
    options: correct.map((c, i) => ({ text: `o${i}`, correct: c, feedback: '' })),
  })
}

// first always picks the first remaining item, so Fisher-Yates turns
// [a, b, c] into [b, c, a].
const first = () => 0

// written is a quiz of three questions q0..q2 with options A (correct, with
// feedback), B and C.
function written(quizShuffle: boolean, questionShuffle: boolean): library.Quiz {
  return library.Quiz.createFrom({
    shuffle: quizShuffle,
    questions: [0, 1, 2].map((n) => ({
      kind: 'choice',
      prompt: `q${n}`,
      explanation: '',
      multiple: false,
      shuffle: questionShuffle,
      options: [
        { text: 'A', correct: true, feedback: 'Yes, A' },
        { text: 'B', correct: false, feedback: 'Not B' },
        { text: 'C', correct: false, feedback: '' },
      ],
    })),
  })
}

function prompts(quiz: library.Quiz): string[] {
  return quiz.questions.map((q) => q.prompt)
}

function texts(q: library.Question): string[] {
  return q.options.map((o) => o.text)
}

describe('arrange', () => {
  it('keeps the written order when both flags are false', () => {
    const quiz = arrange(written(false, false), first)

    expect(prompts(quiz)).toEqual(['q0', 'q1', 'q2'])
    expect(quiz.questions.map(texts)).toEqual([
      ['A', 'B', 'C'],
      ['A', 'B', 'C'],
      ['A', 'B', 'C'],
    ])
  })

  it('shuffles only the questions when only the quiz flag is set', () => {
    const quiz = arrange(written(true, false), first)

    expect(prompts(quiz)).toEqual(['q1', 'q2', 'q0'])
    expect(quiz.questions.every((q) => texts(q).join() === 'A,B,C')).toBe(true)
  })

  it('shuffles only the options when only the question flag is set', () => {
    const quiz = arrange(written(false, true), first)

    expect(prompts(quiz)).toEqual(['q0', 'q1', 'q2'])
    expect(quiz.questions.every((q) => texts(q).join() === 'B,C,A')).toBe(true)
  })

  it('keeps each option with its correctness and feedback', () => {
    const q = arrange(written(true, true), first).questions[0]

    expect(q.options).toEqual([
      { text: 'B', correct: false, feedback: 'Not B' },
      { text: 'C', correct: false, feedback: '' },
      { text: 'A', correct: true, feedback: 'Yes, A' },
    ])
    expect(isCorrect(q, [2])).toBe(true)
    expect(isCorrect(q, [0])).toBe(false)
  })

  it('leaves the input untouched', () => {
    const quiz = written(true, true)
    const before = JSON.stringify(quiz)

    arrange(quiz, first)

    expect(JSON.stringify(quiz)).toBe(before)
  })

  it('returns copies of the same classes', () => {
    const quiz = written(true, true)

    const arranged = arrange(quiz, first)

    expect(arranged).not.toBe(quiz)
    expect(arranged).toBeInstanceOf(library.Quiz)
    expect(arranged.questions[0]).not.toBe(quiz.questions[1])
    expect(arranged.questions[0]).toBeInstanceOf(library.Question)
  })
})

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
