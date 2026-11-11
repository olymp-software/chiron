import type { library } from '../../wailsjs/go/models'

// Result is what the question map shows for a question.
export type Result = 'unchecked' | 'correct' | 'incorrect'

// Answer is the learner's work on one question: the selected option indexes
// and, once checked, whether the answer is correct. A checked answer is locked.
export interface Answer {
  selected: number[]
  checked: boolean
  correct: boolean
}

// freshAnswers returns an unchecked answer with no selection per question.
export function freshAnswers(quiz: library.Quiz): Answer[] {
  return quiz.questions.map(() => ({ selected: [], checked: false, correct: false }))
}

// isCorrect reports whether selected, a set of option indexes, is exactly the
// set of correct options of q; there is no partial credit.
export function isCorrect(q: library.Question, selected: readonly number[]): boolean {
  const chosen = new Set(selected)
  return q.options.every((o, i) => o.correct === chosen.has(i))
}

// result returns the map state of an answer.
export function result(a: Answer): Result {
  if (!a.checked) {
    return 'unchecked'
  }
  return a.correct ? 'correct' : 'incorrect'
}
